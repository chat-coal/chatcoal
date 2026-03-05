package routes

import (
	"chatcoal/middleware"
	"chatcoal/models"
	"chatcoal/services"
	"chatcoal/ws"
	"encoding/json"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// authHandshake reads the first WebSocket message and expects:
//
//	{"type":"auth","token":"<firebase-id-token>"}
//
// Returns the authenticated user, or closes the connection and returns nil on failure.
// This avoids passing tokens in the URL query string where they appear in server logs.
func authHandshake(c *websocket.Conn) *models.User {
	c.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, raw, err := c.ReadMessage()
	if err != nil {
		return nil
	}
	c.SetReadDeadline(time.Time{})

	var msg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil || msg.Type != "auth" || msg.Token == "" {
		return nil
	}

	firebaseUID, err := middleware.VerifyTokenUnified(msg.Token)
	if err != nil {
		return nil
	}

	user, err := services.GetUserByFirebaseUID(firebaseUID)
	if err != nil || user == nil {
		return nil
	}
	return user
}

func SetupWebSocket(app *fiber.App, hub *ws.Hub) {
	app.Use("/ws", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}
		return c.Next()
	})

	// Multiplexed connection — subscribes to all user's servers
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		user := authHandshake(c)
		if user == nil {
			c.Close()
			return
		}

		serverIDs := make(map[models.Snowflake]bool)
		if servers, err := services.GetServersByUserID(user.ID); err == nil {
			for _, s := range servers {
				serverIDs[s.ID] = true
			}
		}

		client := ws.NewClient(c, serverIDs, user.ID)
		hub.Register(client)

		go client.WritePump()
		client.ReadPump()
	}))
}
