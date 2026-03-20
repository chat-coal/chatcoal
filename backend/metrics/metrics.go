package metrics

import (
	"runtime"
	"sync/atomic"
)

var (
	WSConnections         atomic.Int64
	WSConnectionsRejected atomic.Int64
	WSDroppedTasks        atomic.Int64
)

type CacheCounters struct {
	User           atomic.Int64
	Member         atomic.Int64
	UserServers    atomic.Int64
	ServerChannels atomic.Int64
	Presigned      atomic.Int64
}

var (
	CacheHits   CacheCounters
	CacheMisses CacheCounters
)

type CacheSnapshot struct {
	User           int64 `json:"user"`
	Member         int64 `json:"member"`
	UserServers    int64 `json:"user_servers"`
	ServerChannels int64 `json:"server_channels"`
	Presigned      int64 `json:"presigned"`
}

type PlatformStats struct {
	TotalServers  int64 `json:"total_servers"`
	TotalChannels int64 `json:"total_channels"`
}

type VoiceStats struct {
	ActiveChannels int64 `json:"active_channels"`
	ActiveUsers    int64 `json:"active_users"`
}

type ProcessStats struct {
	Goroutines int    `json:"goroutines"`
	HeapMB     uint64 `json:"heap_mb"`
	SysMB      uint64 `json:"sys_mb"`
}

type Snapshot struct {
	WSConnections         int64         `json:"ws_connections"`
	WSConnectionsRejected int64         `json:"ws_connections_rejected_total"`
	WSDroppedTasks        int64         `json:"ws_dropped_tasks_total"`
	WSShardQueueDepths    []int         `json:"ws_shard_queue_depths"`
	CacheHits             CacheSnapshot `json:"cache_hits"`
	CacheMisses           CacheSnapshot `json:"cache_misses"`
	Platform              PlatformStats `json:"platform"`
	Voice                 VoiceStats    `json:"voice"`
	ActiveUsers           int64         `json:"active_users"`
	Process               ProcessStats  `json:"process"`
}

func Take(shardDepths []int, platform PlatformStats, voice VoiceStats, activeUsers int64) Snapshot {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return Snapshot{
		ActiveUsers: activeUsers,
		Process: ProcessStats{
			Goroutines: runtime.NumGoroutine(),
			HeapMB:     mem.HeapAlloc / (1024 * 1024),
			SysMB:      mem.Sys / (1024 * 1024),
		},
		WSConnections:         WSConnections.Load(),
		WSConnectionsRejected: WSConnectionsRejected.Load(),
		WSDroppedTasks:        WSDroppedTasks.Load(),
		WSShardQueueDepths:    shardDepths,
		CacheHits: CacheSnapshot{
			User:           CacheHits.User.Load(),
			Member:         CacheHits.Member.Load(),
			UserServers:    CacheHits.UserServers.Load(),
			ServerChannels: CacheHits.ServerChannels.Load(),
			Presigned:      CacheHits.Presigned.Load(),
		},
		CacheMisses: CacheSnapshot{
			User:           CacheMisses.User.Load(),
			Member:         CacheMisses.Member.Load(),
			UserServers:    CacheMisses.UserServers.Load(),
			ServerChannels: CacheMisses.ServerChannels.Load(),
			Presigned:      CacheMisses.Presigned.Load(),
		},
		Platform: platform,
		Voice:    voice,
	}
}
