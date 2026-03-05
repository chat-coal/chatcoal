package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"chatcoal/models"

	"github.com/gofiber/fiber/v2/log"
)

var instanceID string

func init() {
	b := make([]byte, 8)
	rand.Read(b)
	instanceID = hex.EncodeToString(b)
}

type pubsubEnvelope struct {
	InstanceID string          `json:"i"`
	Payload    json.RawMessage `json:"p"`
}

// PublishToServer publishes a WebSocket message to all instances for a server.
func PublishToServer(serverID models.Snowflake, message []byte) {
	if Redis == nil {
		return
	}
	env, _ := json.Marshal(pubsubEnvelope{
		InstanceID: instanceID,
		Payload:    message,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	logRedisErr(Redis.Publish(ctx, fmt.Sprintf("ws:s:%d", serverID), env).Err(), "PublishToServer")
}

// PublishToUser publishes a WebSocket message to all instances for a user.
func PublishToUser(userID models.Snowflake, message []byte) {
	if Redis == nil {
		return
	}
	env, _ := json.Marshal(pubsubEnvelope{
		InstanceID: instanceID,
		Payload:    message,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	logRedisErr(Redis.Publish(ctx, fmt.Sprintf("ws:u:%d", userID), env).Err(), "PublishToUser")
}

// PubSubReceiver is implemented by the Hub to receive cross-instance messages.
type PubSubReceiver interface {
	BroadcastLocal(serverID models.Snowflake, message []byte)
	SendToUserGlobalLocal(userID models.Snowflake, data []byte)
}

var subscriberCancel context.CancelFunc

// StartSubscriber begins listening for Redis pub/sub messages and delivers them locally.
func StartSubscriber(recv PubSubReceiver) {
	if Redis == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	subscriberCancel = cancel
	go runSubscriber(ctx, recv)
	log.Info("Redis pub/sub subscriber started (instance: ", instanceID, ")")
}

// StopSubscriber gracefully shuts down the pub/sub subscriber goroutine.
func StopSubscriber() {
	if subscriberCancel != nil {
		subscriberCancel()
	}
}

func runSubscriber(ctx context.Context, recv PubSubReceiver) {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		ps := Redis.PSubscribe(ctx, "ws:s:*", "ws:u:*")

		ch := ps.Channel()
		for msg := range ch {
			var env pubsubEnvelope
			if json.Unmarshal([]byte(msg.Payload), &env) != nil {
				continue
			}
			if env.InstanceID == instanceID {
				continue
			}

			parts := strings.SplitN(msg.Channel, ":", 3)
			if len(parts) != 3 {
				continue
			}

			id, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				continue
			}

			switch parts[1] {
			case "s":
				recv.BroadcastLocal(models.Snowflake(id), env.Payload)
			case "u":
				recv.SendToUserGlobalLocal(models.Snowflake(id), env.Payload)
			}
		}

		ps.Close()

		// Check if we were asked to shut down
		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Warn("Redis pub/sub disconnected, reconnecting in ", backoff, "...")
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}

		// Exponential backoff capped at maxBackoff, reset on successful reconnect
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}
