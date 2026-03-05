package ws

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"chatcoal/cache"
	"chatcoal/metrics"
	"chatcoal/models"
	"chatcoal/services"
)

const numShards = 16
const maxClients = 10_000 // max concurrent WebSocket connections

// shard holds a subset of server broadcast groups, each served by its own goroutine.
type shard struct {
	mu      sync.RWMutex
	servers map[models.Snowflake]map[*Client]bool
	inbox   chan broadcastMsg
}

func (s *shard) run(h *Hub) {
	for msg := range s.inbox {
		s.mu.RLock()
		var stale []*Client
		if clients, ok := s.servers[msg.serverID]; ok {
			for client := range clients {
				select {
				case client.Send <- msg.message:
				default:
					stale = append(stale, client)
				}
			}
		}
		s.mu.RUnlock()
		for _, client := range stale {
			h.Unregister(client)
		}
	}
}

type Hub struct {
	shards [numShards]shard

	mu            sync.RWMutex
	userClients   map[models.Snowflake]map[*Client]bool
	offlineTimers map[models.Snowflake]*time.Timer
	voiceStates   map[models.Snowflake]map[models.Snowflake]*Client // channelID -> userID -> Client
	clientVoice   map[*Client]voiceInfo                             // client -> voice channel info

	// Bounded worker pool for async tasks (presence broadcasts, voice cleanup, etc.)
	workCh chan func()

	clientCount int64 // atomic, tracks active WebSocket connections
}

const workerCount = 8
const workQueueSize = 256

type broadcastMsg struct {
	serverID models.Snowflake
	message  []byte
}

type voiceInfo struct {
	ChannelID models.Snowflake
	ServerID  models.Snowflake
}

var hub *Hub

func GetHub() *Hub {
	return hub
}

func NewHub() *Hub {
	hub = &Hub{
		userClients:   make(map[models.Snowflake]map[*Client]bool),
		offlineTimers: make(map[models.Snowflake]*time.Timer),
		voiceStates:   make(map[models.Snowflake]map[models.Snowflake]*Client),
		clientVoice:   make(map[*Client]voiceInfo),
		workCh:        make(chan func(), workQueueSize),
	}
	for i := range hub.shards {
		hub.shards[i] = shard{
			servers: make(map[models.Snowflake]map[*Client]bool),
			inbox:   make(chan broadcastMsg, 256),
		}
	}
	return hub
}

func (h *Hub) shardFor(serverID models.Snowflake) *shard {
	return &h.shards[serverID%numShards]
}

// presenceRefreshInterval is how often we re-SET all online presence keys in
// Redis so their TTLs don't silently expire while users are still connected.
const presenceRefreshInterval = 5 * time.Minute

// Run starts all shard goroutines and worker pool. Blocks forever.
func (h *Hub) Run() {
	for range workerCount {
		go func() {
			for fn := range h.workCh {
				fn()
			}
		}()
	}

	// Periodically refresh Redis presence TTLs for all connected users.
	go func() {
		ticker := time.NewTicker(presenceRefreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			h.mu.RLock()
			for userID := range h.userClients {
				cache.SetOnline(userID)
			}
			h.mu.RUnlock()
		}
	}()

	var wg sync.WaitGroup
	for i := range h.shards {
		wg.Add(1)
		go func(s *shard) {
			defer wg.Done()
			s.run(h)
		}(&h.shards[i])
	}
	wg.Wait()
}

// submit sends work to the bounded worker pool. Non-blocking: drops the task if
// the queue is full to avoid blocking the caller.
func (h *Hub) submit(fn func()) {
	select {
	case h.workCh <- fn:
	default:
		metrics.WSDroppedTasks.Add(1)
	}
}

// Register adds a client to its server broadcast groups and tracks presence.
// Returns without registering if the server is at capacity; the caller should
// check that the connection is still open after this call.
func (h *Hub) Register(client *Client) {
	if atomic.AddInt64(&h.clientCount, 1) > maxClients {
		atomic.AddInt64(&h.clientCount, -1)
		metrics.WSConnectionsRejected.Add(1)
		// Use CloseOnce so unregisterClient won't double-close Send later.
		client.CloseOnce.Do(func() {
			close(client.Send)
		})
		client.Conn.Close()
		return
	}
	metrics.WSConnections.Add(1)

	for serverID := range client.ServerIDSnapshot() {
		s := h.shardFor(serverID)
		s.mu.Lock()
		if s.servers[serverID] == nil {
			s.servers[serverID] = make(map[*Client]bool)
		}
		s.servers[serverID][client] = true
		s.mu.Unlock()
	}

	h.mu.Lock()
	if timer, ok := h.offlineTimers[client.UserID]; ok {
		timer.Stop()
		delete(h.offlineTimers, client.UserID)
	}
	wasOffline := len(h.userClients[client.UserID]) == 0
	if h.userClients[client.UserID] == nil {
		h.userClients[client.UserID] = make(map[*Client]bool)
	}
	h.userClients[client.UserID][client] = true
	userID := client.UserID
	h.mu.Unlock()

	if wasOffline {
		cache.SetOnline(userID)
		if u, err := services.GetUserByID(userID); err == nil && u.Status == "invisible" {
			h.BroadcastPresenceChange(userID, "offline")
		} else {
			services.UpdateUserStatus(userID, "online")
			h.BroadcastPresenceChange(userID, "online")
		}
	}
}

// Unregister removes a client from all server groups, voice, and presence tracking.
// Safe to call multiple times (guarded by sync.Once on the client).
func (h *Hub) Unregister(client *Client) {
	client.CloseOnce.Do(func() {
		h.unregisterClient(client)
	})
}

func (h *Hub) unregisterClient(client *Client) {
	atomic.AddInt64(&h.clientCount, -1)
	metrics.WSConnections.Add(-1)

	h.mu.Lock()
	_, wasInVoice := h.clientVoice[client]
	h.leaveVoiceLocked(client)

	if clients, ok := h.userClients[client.UserID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.userClients, client.UserID)
			userID := client.UserID
			h.offlineTimers[userID] = time.AfterFunc(5*time.Second, func() {
				h.mu.Lock()
				stillOffline := len(h.userClients[userID]) == 0
				delete(h.offlineTimers, userID)
				h.mu.Unlock()
				if stillOffline {
					cache.SetOffline(userID)
					if u, err := services.GetUserByID(userID); err == nil && u.Status == "invisible" {
						return
					}
					services.UpdateUserStatus(userID, "offline")
					h.BroadcastPresenceChange(userID, "offline")
				}
			})
		}
	}
	h.mu.Unlock()

	for serverID := range client.ServerIDSnapshot() {
		s := h.shardFor(serverID)
		s.mu.Lock()
		if clients, ok := s.servers[serverID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(s.servers, serverID)
			}
		}
		s.mu.Unlock()
	}

	close(client.Send)

	if wasInVoice {
		userID := client.UserID
		h.submit(func() { cache.LeaveVoiceChannel(userID) })
	}
}

// BroadcastLocal delivers to local clients only (used by Redis subscriber).
func (h *Hub) BroadcastLocal(serverID models.Snowflake, message []byte) {
	h.shardFor(serverID).inbox <- broadcastMsg{serverID: serverID, message: message}
}

// Broadcast delivers to local clients and publishes to Redis for cross-instance delivery.
func (h *Hub) Broadcast(serverID models.Snowflake, message []byte) {
	h.BroadcastLocal(serverID, message)
	cache.PublishToServer(serverID, message)
}

// Subscribe adds a client to a server broadcast group (e.g., after joining a server).
func (h *Hub) Subscribe(client *Client, serverID models.Snowflake) {
	client.AddServerID(serverID)
	s := h.shardFor(serverID)
	s.mu.Lock()
	if s.servers[serverID] == nil {
		s.servers[serverID] = make(map[*Client]bool)
	}
	s.servers[serverID][client] = true
	s.mu.Unlock()
}

// Unsubscribe removes a client from a server broadcast group (e.g., after leaving a server).
func (h *Hub) Unsubscribe(client *Client, serverID models.Snowflake) {
	client.RemoveServerID(serverID)
	s := h.shardFor(serverID)
	s.mu.Lock()
	if clients, ok := s.servers[serverID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(s.servers, serverID)
		}
	}
	s.mu.Unlock()
}

func (h *Hub) JoinVoice(client *Client, channelID, serverID models.Snowflake) {
	h.mu.Lock()
	h.leaveVoiceLocked(client)

	if h.voiceStates[channelID] == nil {
		h.voiceStates[channelID] = make(map[models.Snowflake]*Client)
	}
	h.voiceStates[channelID][client.UserID] = client
	h.clientVoice[client] = voiceInfo{ChannelID: channelID, ServerID: serverID}

	h.broadcastVoiceStateUpdate(serverID, channelID, client.UserID, "join")
	h.mu.Unlock()

	cache.JoinVoiceChannel(serverID, channelID, client.UserID)
}

func (h *Hub) LeaveVoice(client *Client) {
	h.mu.Lock()
	h.leaveVoiceLocked(client)
	h.mu.Unlock()

	userID := client.UserID
	h.submit(func() { cache.LeaveVoiceChannel(userID) })
}

// leaveVoiceLocked removes a client from voice. Caller must hold h.mu.
func (h *Hub) leaveVoiceLocked(client *Client) {
	info, ok := h.clientVoice[client]
	if !ok {
		return
	}
	delete(h.clientVoice, client)
	if users, exists := h.voiceStates[info.ChannelID]; exists {
		delete(users, client.UserID)
		if len(users) == 0 {
			delete(h.voiceStates, info.ChannelID)
		}
	}
	h.broadcastVoiceStateUpdate(info.ServerID, info.ChannelID, client.UserID, "leave")
}

// broadcastVoiceStateUpdate sends a voice_state_update to all clients in the server.
// Caller must hold h.mu. Sends via shard inbox to avoid nested locks.
func (h *Hub) broadcastVoiceStateUpdate(serverID, channelID, userID models.Snowflake, action string) {
	payload, _ := json.Marshal(map[string]interface{}{
		"type":      "voice_state_update",
		"server_id": serverID,
		"data": map[string]interface{}{
			"channel_id": channelID,
			"user_id":    userID,
			"action":     action,
		},
	})
	h.shardFor(serverID).inbox <- broadcastMsg{serverID: serverID, message: payload}
	h.submit(func() { cache.PublishToServer(serverID, payload) })
}

// SendToUser sends data to a specific user within a server, locally and via Redis.
func (h *Hub) SendToUser(serverID, userID models.Snowflake, data []byte) {
	s := h.shardFor(serverID)
	s.mu.RLock()
	if clients, ok := s.servers[serverID]; ok {
		for client := range clients {
			if client.UserID == userID {
				select {
				case client.Send <- data:
				default:
				}
			}
		}
	}
	s.mu.RUnlock()
	cache.PublishToUser(userID, data)
}

// SendToUserGlobalLocal delivers to all local clients of a user (used by Redis subscriber).
func (h *Hub) SendToUserGlobalLocal(userID models.Snowflake, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.userClients[userID]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

// SendToUserGlobal delivers to all clients of a user locally and via Redis.
func (h *Hub) SendToUserGlobal(userID models.Snowflake, data []byte) {
	h.SendToUserGlobalLocal(userID, data)
	cache.PublishToUser(userID, data)
}

// BroadcastUserProfileChange notifies all servers the user belongs to about a profile change (display_name, avatar).
func (h *Hub) BroadcastUserProfileChange(userID models.Snowflake, displayName, avatarURL string) {
	h.submit(func() {
		servers, err := services.GetServersByUserID(userID)
		if err != nil {
			return
		}

		for _, server := range servers {
			payload, _ := json.Marshal(map[string]interface{}{
				"type":      "user_update",
				"server_id": server.ID,
				"data": map[string]interface{}{
					"user_id":      userID,
					"display_name": displayName,
					"avatar_url":   avatarURL,
				},
			})
			h.Broadcast(server.ID, payload)
		}
	})
}

// BroadcastPresenceChange notifies all servers the user belongs to about their status change.
func (h *Hub) BroadcastPresenceChange(userID models.Snowflake, status string) {
	h.submit(func() {
		servers, err := services.GetServersByUserID(userID)
		if err != nil {
			return
		}

		for _, server := range servers {
			payload, _ := json.Marshal(map[string]interface{}{
				"type":      "presence_update",
				"server_id": server.ID,
				"data": map[string]interface{}{
					"user_id": userID,
					"status":  status,
				},
			})
			h.Broadcast(server.ID, payload)
		}
	})
}

// GetOnlineUserIDs returns the set of all currently connected user IDs.
// Uses Redis for cross-instance visibility, falls back to local state.
func (h *Hub) GetOnlineUserIDs() map[models.Snowflake]bool {
	if result := cache.GetOnlineUserIDs(); result != nil {
		return result
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make(map[models.Snowflake]bool, len(h.userClients))
	for uid := range h.userClients {
		result[uid] = true
	}
	return result
}

// ShardQueueDepths returns the current inbox length for each shard.
func (h *Hub) ShardQueueDepths() []int {
	depths := make([]int, numShards)
	for i := range h.shards {
		depths[i] = len(h.shards[i].inbox)
	}
	return depths
}

// GetVoiceStates returns channelID -> []userID for all voice channels in the given server.
// Uses Redis for cross-instance visibility, falls back to local state.
func (h *Hub) GetVoiceStates(serverID models.Snowflake) map[models.Snowflake][]models.Snowflake {
	if result := cache.GetVoiceStatesFromRedis(serverID); result != nil {
		return result
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make(map[models.Snowflake][]models.Snowflake)
	for channelID, users := range h.voiceStates {
		for userID, client := range users {
			if info, ok := h.clientVoice[client]; ok && info.ServerID == serverID {
				result[channelID] = append(result[channelID], userID)
			}
		}
	}
	return result
}
