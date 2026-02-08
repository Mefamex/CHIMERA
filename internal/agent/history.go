package agent

import "github.com/ollama/ollama/api"

func trimConversation(history []api.Message, maxMsgs int) []api.Message {
	if maxMsgs <= 0 {
		maxMsgs = 20
	}
	if len(history) <= maxMsgs {
		return history
	}

	var newHistory []api.Message
	if len(history) > 0 && history[0].Role == "system" {
		newHistory = append(newHistory, history[0])
		start := len(history) - (maxMsgs - 1)
		if start < 1 {
			start = 1
		}
		newHistory = append(newHistory, history[start:]...)
		return newHistory
	}

	start := len(history) - maxMsgs
	return append(newHistory, history[start:]...)
}
