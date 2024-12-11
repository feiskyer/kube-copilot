package workflows

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/feiskyer/kube-copilot/pkg/tools"
	"github.com/feiskyer/swarm-go"
)

var (
	// auditFunc is a Swarm function that conducts a structured security audit of a Kubernetes Pod.
	trivyFunc = swarm.NewAgentFunction(
		"trivy",
		"Run trivy image scanning for a given image",
		func(args map[string]interface{}) (interface{}, error) {
			image, ok := args["image"].(string)
			if !ok {
				return nil, fmt.Errorf("image not provided")
			}

			result, err := tools.Trivy(image)
			if err != nil {
				return nil, err
			}

			return result, nil
		},
		[]swarm.Parameter{
			{Name: "image", Type: reflect.TypeOf(""), Required: true},
		},
	)

	// kubectlFunc is a Swarm function that runs kubectl command.
	kubectlFunc = swarm.NewAgentFunction(
		"kubectl",
		"Run kubectl command",
		func(args map[string]interface{}) (interface{}, error) {
			command, ok := args["command"].(string)
			if !ok {
				return nil, fmt.Errorf("command not provided")
			}

			result, err := tools.Kubectl(command)
			if err != nil {
				return nil, err
			}

			return result, nil
		},
		[]swarm.Parameter{
			{Name: "command", Type: reflect.TypeOf(""), Required: true},
		},
	)
)

// NewSwarm creates a new Swarm client.
func NewSwarm() (*swarm.Swarm, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	baseURL := os.Getenv("OPENAI_API_BASE")
	// OpenAI
	if baseURL == "" {
		return swarm.NewSwarm(swarm.NewOpenAIClient(apiKey)), nil
	}

	// Azure OpenAI
	if strings.Contains(baseURL, "azure") {
		return swarm.NewSwarm(swarm.NewAzureOpenAIClient(apiKey, baseURL)), nil
	}

	// OpenAI compatible LLM
	return swarm.NewSwarm(swarm.NewOpenAIClientWithBaseURL(apiKey, baseURL)), nil
}
