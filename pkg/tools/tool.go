package tools

// Tool is a function that takes an input and returns an output.
type Tool func(input string) (string, error)

// CopilotTools is a map of tool names to tools.
var CopilotTools = map[string]Tool{
	"search":  GoogleSearch,
	"python":  PythonREPL,
	"trivy":   Trivy,
	"kubectl": Kubectl,
}

// ToolPrompt is the JSON format for the prompt.
type ToolPrompt struct {
	Question string `json:"question"`
	Thought  string `json:"thought,omitempty"`
	Action   struct {
		Name  string `json:"name"`
		Input string `json:"input"`
	} `json:"action,omitempty"`
	Observation string `json:"observation,omitempty"`
	FinalAnswer string `json:"final_answer,omitempty"`
}
