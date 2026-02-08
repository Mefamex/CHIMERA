package agent

// InvalidJSONError signals that the model did not return valid JSON.
type InvalidJSONError struct {
	Raw   string
	Clean string
}

func (e InvalidJSONError) Error() string {
	return "invalid JSON from model"
}

// InvalidAgentResponseError signals that the model returned JSON but not in the required shape/content.
// This is retryable by asking the model to strictly follow the schema.
type InvalidAgentResponseError struct {
	Raw    string
	Clean  string
	Reason string
}

func (e InvalidAgentResponseError) Error() string {
	return "invalid agent response from model"
}
