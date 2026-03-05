package services

import (
	"chatcoal/models"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

func GenerateLiveKitToken(userID models.Snowflake, channelID models.Snowflake, displayName string) (string, error) {
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	if apiKey == "" || apiSecret == "" {
		return "", errors.New("LiveKit credentials not configured")
	}

	at := auth.NewAccessToken(apiKey, apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     fmt.Sprintf("channel_%d", channelID),
	}
	at.SetVideoGrant(grant).
		SetIdentity(fmt.Sprintf("%d", userID)).
		SetName(displayName).
		SetValidFor(24 * time.Hour)

	return at.ToJWT()
}

// GetVoiceStats returns the number of active LiveKit rooms and total participants.
func GetVoiceStats() (activeChannels, activeUsers int64, err error) {
	url := os.Getenv("LIVEKIT_URL")
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	if url == "" || apiKey == "" || apiSecret == "" {
		return 0, 0, nil
	}

	client := lksdk.NewRoomServiceClient(url, apiKey, apiSecret)
	rooms, err := client.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
	if err != nil {
		return 0, 0, err
	}

	for _, room := range rooms.GetRooms() {
		if room.NumParticipants > 0 {
			activeChannels++
			activeUsers += int64(room.NumParticipants)
		}
	}
	return activeChannels, activeUsers, nil
}
