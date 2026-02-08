package agent

import (
	"CHIMERA/internal/core"
	"CHIMERA/internal/skills"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
)

type Engine struct {
	Brain         *BrainStruct
	MaxIterations int
	MaxHistory    int
}

func NewEngine(br *BrainStruct) *Engine {
	maxHistory := core.Config.LLM.MaxConversationMessages
	if maxHistory <= 0 {
		maxHistory = 30
	}
	return &Engine{
		Brain:         br,
		MaxIterations: 10,
		MaxHistory:    maxHistory,
	}
}

func (e *Engine) Run(ctx context.Context, history []api.Message, userText string, emit func(Event)) ([]api.Message, error) {
	if e == nil || e.Brain == nil {
		return history, fmt.Errorf("agent engine not initialized")
	}
	if emit == nil {
		return history, fmt.Errorf("emit function is nil")
	}

	log.Chat("USER", userText, true)
	currentInput := userText

	for i := 0; i < e.MaxIterations; i++ {
		agentResp, newHistory, err := e.Brain.ChatAgent(ctx, history, currentInput)
		if err != nil {
			var invalidErr InvalidJSONError
			if errors.As(err, &invalidErr) {
				invalidOutput := invalidErr.Clean
				if invalidOutput == "" {
					invalidOutput = invalidErr.Raw
				}
				if invalidOutput != "" {
					emit(Event{Type: EventReply, Content: invalidOutput})
					log.Chat("ASSISTANT", invalidOutput, true)
				}
				invalidForPrompt := invalidOutput
				if strings.TrimSpace(invalidForPrompt) == "" {
					invalidForPrompt = "<empty>"
				}
				currentInput = "Your previous response was not valid JSON.\nReturn ONLY valid JSON with keys: thought, tool, args, reply. Provide either a non-empty tool (with args) OR a non-empty reply."
				history = newHistory
				continue
			}

			var invalidRespErr InvalidAgentResponseError
			if errors.As(err, &invalidRespErr) {
				invalidOutput := invalidRespErr.Clean
				if invalidOutput == "" {
					invalidOutput = invalidRespErr.Raw
				}
				if invalidOutput != "" {
					emit(Event{Type: EventReply, Content: invalidOutput})
					log.Chat("ASSISTANT", invalidOutput, true)
				}
				reason := invalidRespErr.Reason
				if reason == "" {
					reason = "Your response did not include a tool call or a reply."
				}
				invalidForPrompt := invalidOutput
				if strings.TrimSpace(invalidForPrompt) == "" {
					invalidForPrompt = "<empty>"
				}
				currentInput = fmt.Sprintf("Your previous response was valid JSON but not usable. Reason: %s\nReturn ONLY valid JSON with keys: thought, tool, args, reply. Provide either a non-empty tool (with args) OR a non-empty reply.", reason)
				history = newHistory
				continue
			}
			emit(Event{Type: EventError, Error: err.Error(), Content: err.Error()})
			return history, err
		}
		history = newHistory

		if agentResp.Thought != "" {
			emit(Event{Type: EventThought, Content: agentResp.Thought})
			log.TraceBrain("Thought", agentResp.Thought, "component", "BRAIN")
		}

		if agentResp.Reply != "" {
			emit(Event{Type: EventReply, Content: agentResp.Reply})
			log.Chat("ASSISTANT", agentResp.Reply, true)
		}

		if agentResp.Tool == "" {
			break
		}

		emit(Event{Type: EventToolCall, ToolName: agentResp.Tool, Args: agentResp.Args})

		toolArgs, err := resolveToolArgs(agentResp.Tool, agentResp.Args)
		if err != nil {
			emit(Event{Type: EventError, Error: err.Error(), Content: err.Error(), ToolName: agentResp.Tool})
			return history, err
		}

		toolResult := skills.UseTool(agentResp.Tool, toolArgs...)
		outputStr := buildToolOutput(toolResult)

		emit(Event{Type: EventToolResult, ToolName: agentResp.Tool, Content: outputStr})

		currentInput = fmt.Sprintf("Tool '%s' Output: %s\nBased on this, what is the next step or final answer?", agentResp.Tool, outputStr)
	}

	history = trimConversation(history, e.MaxHistory)
	return history, nil
}

func resolveToolArgs(toolName string, args map[string]interface{}) ([]string, error) {
	toolDef, ok := skills.GetTool(toolName)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	toolArgs := make([]string, 0, len(toolDef.Parameters))
	for _, paramName := range toolDef.Parameters {
		v, exists := args[paramName]
		if !exists {
			return nil, fmt.Errorf("missing argument '%s' for tool '%s'", paramName, toolName)
		}
		toolArgs = append(toolArgs, fmt.Sprintf("%v", v))
	}
	return toolArgs, nil
}

func buildToolOutput(result *skills.ToolResult) string {
	if result == nil {
		return "Tool Error: empty result"
	}
	if result.Error != nil {
		return fmt.Sprintf("Tool Error: %v", result.Error)
	}
	if result.Stdout != "" {
		return result.Stdout
	}
	if len(result.Data) == 0 {
		return ""
	}
	jsonBytes, err := json.Marshal(result.Data)
	if err != nil {
		return fmt.Sprintf("Tool Error: %v", err)
	}
	return string(jsonBytes)
}
