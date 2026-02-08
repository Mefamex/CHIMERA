package server

import (
	"time"

	"CHIMERA/internal/core"
	"CHIMERA/internal/models"
)

func init() {
	core.RegisterLogSink(broadcastLog)
}

func broadcastLog(logType core.LogType, data map[string]any) {
	env := models.Envelope{
		Broadcast: true,
		Timestamp: time.Now(),
		Payload: models.AgentResponse{
			Type:    models.AgentTypeLog,
			LogType: string(logType),
			Data:    data,
		},
	}
	if msg, ok := data["msg"].(string); ok {
		env.Payload.Content = msg
	}
	HubInstance.Broadcast <- env
}
