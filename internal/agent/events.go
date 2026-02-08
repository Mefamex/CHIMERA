package agent

type EventType string

const (
	EventThought    EventType = "thought"
	EventReply      EventType = "reply"
	EventToolCall   EventType = "tool_call"
	EventToolResult EventType = "tool_result"
	EventError      EventType = "error"
)

type Event struct {
	Type     EventType
	Content  string
	ToolName string
	Args     map[string]interface{}
	Error    string
}
