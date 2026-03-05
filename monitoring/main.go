package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed dashboard.html
var staticFiles embed.FS

// CacheSnapshot mirrors the backend JSON shape.
type CacheSnapshot struct {
	User           int64 `json:"user"`
	Member         int64 `json:"member"`
	UserServers    int64 `json:"user_servers"`
	ServerChannels int64 `json:"server_channels"`
	Presigned      int64 `json:"presigned"`
}

type PlatformSnapshot struct {
	TotalServers  int64 `json:"total_servers"`
	TotalChannels int64 `json:"total_channels"`
}

type VoiceSnapshot struct {
	ActiveChannels int64 `json:"active_channels"`
	ActiveUsers    int64 `json:"active_users"`
}

type Snapshot struct {
	WSConnections         int64            `json:"ws_connections"`
	WSConnectionsRejected int64            `json:"ws_connections_rejected_total"`
	WSDroppedTasks        int64            `json:"ws_dropped_tasks_total"`
	WSShardQueueDepths    []int            `json:"ws_shard_queue_depths"`
	CacheHits             CacheSnapshot    `json:"cache_hits"`
	CacheMisses           CacheSnapshot    `json:"cache_misses"`
	Platform              PlatformSnapshot `json:"platform"`
	Voice                 VoiceSnapshot    `json:"voice"`
}

type TimestampedSnapshot struct {
	Timestamp int64    `json:"ts"`
	Snapshot  Snapshot `json:"snapshot"`
}

const ringSize = 120

type Store struct {
	mu      sync.RWMutex
	ring    [ringSize]TimestampedSnapshot
	head    int
	count   int
	healthy bool
}

func (s *Store) Add(snap Snapshot) {
	s.mu.Lock()
	s.ring[s.head] = TimestampedSnapshot{
		Timestamp: time.Now().UnixMilli(),
		Snapshot:  snap,
	}
	s.head = (s.head + 1) % ringSize
	if s.count < ringSize {
		s.count++
	}
	s.healthy = true
	s.mu.Unlock()
}

func (s *Store) SetUnhealthy() {
	s.mu.Lock()
	s.healthy = false
	s.mu.Unlock()
}

func (s *Store) History() ([]TimestampedSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]TimestampedSnapshot, s.count)
	for i := 0; i < s.count; i++ {
		idx := (s.head - s.count + i + ringSize) % ringSize
		out[i] = s.ring[idx]
	}
	return out, s.healthy
}

func poll(store *Store, metricsURL, token string, interval time.Duration) {
	client := &http.Client{Timeout: 5 * time.Second}
	for {
		req, err := http.NewRequest("GET", metricsURL, nil)
		if err != nil {
			store.SetUnhealthy()
			time.Sleep(interval)
			continue
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			store.SetUnhealthy()
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(interval)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			store.SetUnhealthy()
			time.Sleep(interval)
			continue
		}
		var snap Snapshot
		if err := json.Unmarshal(body, &snap); err != nil {
			store.SetUnhealthy()
			time.Sleep(interval)
			continue
		}
		store.Add(snap)
		time.Sleep(interval)
	}
}

func loadEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

func main() {
	loadEnv(".env")
	metricsURL := os.Getenv("CHATCOAL_METRICS_URL")
	if metricsURL == "" {
		metricsURL = "http://localhost:3000/internal/metrics"
	}
	token := os.Getenv("METRICS_TOKEN")
	port := os.Getenv("PORT")
	if port == "" {
		port = "9091"
	}
	intervalSec := 5
	if s := os.Getenv("POLL_INTERVAL"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			intervalSec = n
		}
	}

	store := &Store{}
	go poll(store, metricsURL, token, time.Duration(intervalSec)*time.Second)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFiles.ReadFile("dashboard.html")
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		history, healthy := store.History()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"history": history,
			"healthy": healthy,
		})
	})

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("chatcoal monitoring listening on %s\n", addr)
	fmt.Printf("Polling %s every %ds\n", metricsURL, intervalSec)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
