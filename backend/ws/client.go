package ws

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"chatcoal/models"

	"github.com/gofiber/contrib/websocket"
)

// parseSnowflake extracts a models.Snowflake from a JSON-decoded interface{}.
// Handles both string ("123") and number (123) representations.
func parseSnowflake(v interface{}) (models.Snowflake, bool) {
	switch val := v.(type) {
	case string:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, false
		}
		return models.Snowflake(n), true
	case float64:
		return models.Snowflake(int64(val)), true
	default:
		return 0, false
	}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 16384
)

type Client struct {
	Conn      *websocket.Conn
	Send      chan []byte
	serverIDs map[models.Snowflake]bool // set of subscribed servers
	sidMu     sync.RWMutex              // protects serverIDs
	UserID    models.Snowflake
	CloseOnce sync.Once
}

// ServerIDSnapshot returns a copy of the client's subscribed server IDs.
func (c *Client) ServerIDSnapshot() map[models.Snowflake]bool {
	c.sidMu.RLock()
	defer c.sidMu.RUnlock()
	cp := make(map[models.Snowflake]bool, len(c.serverIDs))
	for id := range c.serverIDs {
		cp[id] = true
	}
	return cp
}

// AddServerID adds a server to the client's subscriptions.
func (c *Client) AddServerID(serverID models.Snowflake) {
	c.sidMu.Lock()
	c.serverIDs[serverID] = true
	c.sidMu.Unlock()
}

// RemoveServerID removes a server from the client's subscriptions.
func (c *Client) RemoveServerID(serverID models.Snowflake) {
	c.sidMu.Lock()
	delete(c.serverIDs, serverID)
	c.sidMu.Unlock()
}

func NewClient(conn *websocket.Conn, serverIDs map[models.Snowflake]bool, userID models.Snowflake) *Client {
	return &Client{
		Conn:      conn,
		Send:      make(chan []byte, 256),
		serverIDs: serverIDs,
		UserID:    userID,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		GetHub().Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var event map[string]interface{}
		if err := json.Unmarshal(msg, &event); err != nil {
			continue
		}

		eventType, _ := event["type"].(string)
		switch eventType {
		case "typing":
			serverID, ok := parseSnowflake(event["server_id"])
			if !ok {
				continue
			}
			payload, _ := json.Marshal(map[string]interface{}{
				"type":      "typing",
				"server_id": serverID,
				"data": map[string]interface{}{
					"user_id":    c.UserID,
					"channel_id": event["channel_id"],
				},
			})
			GetHub().Broadcast(serverID, payload)

		case "dm_typing":
			dmChannelID, ok := parseSnowflake(event["dm_channel_id"])
			targetUserID, ok2 := parseSnowflake(event["target_user_id"])
			if ok && ok2 {
				payload, _ := json.Marshal(map[string]interface{}{
					"type": "dm_typing",
					"data": map[string]interface{}{
						"user_id":       c.UserID,
						"dm_channel_id": dmChannelID,
					},
				})
				GetHub().SendToUserGlobal(targetUserID, payload)
			}

		case "voice_join":
			channelID, ok := parseSnowflake(event["channel_id"])
			serverID, ok2 := parseSnowflake(event["server_id"])
			if ok && ok2 {
				GetHub().JoinVoice(c, channelID, serverID)
			}

		case "voice_leave":
			GetHub().LeaveVoice(c)

		case "webrtc_offer", "webrtc_answer", "webrtc_ice":
			c.forwardWebRTCSignal(event)

		case "subscribe":
			serverID, ok := parseSnowflake(event["server_id"])
			if ok {
				GetHub().Subscribe(c, serverID)
			}

		case "unsubscribe":
			serverID, ok := parseSnowflake(event["server_id"])
			if ok {
				GetHub().Unsubscribe(c, serverID)
			}
		}
	}
}

func (c *Client) forwardWebRTCSignal(event map[string]interface{}) {
	targetUserID, ok := parseSnowflake(event["target_user_id"])
	if !ok {
		return
	}

	data := make(map[string]interface{})
	for k, v := range event {
		if k != "type" && k != "target_user_id" {
			data[k] = v
		}
	}
	data["user_id"] = c.UserID

	payload, err := json.Marshal(map[string]interface{}{
		"type": event["type"],
		"data": data,
	})
	if err != nil {
		return
	}
	GetHub().SendToUserGlobal(targetUserID, payload)
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
