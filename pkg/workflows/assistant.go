package workflows

import (
	"context"
	"fmt"
	"os"

	"github.com/feiskyer/swarm-go"
)

const assistantPrompt = `As a Kubernetes expert, guide the user according to the given instructions to solve their problem or achieve their objective.

Understand the nature of their request, clarify any complex concepts, and provide step-by-step guidance tailored to their specific needs. Ensure that your explanations are comprehensive, using precise Kubernetes terminology and concepts.

# Steps

1. **Interpret User Intent**: Carefully analyze the user's instructions or questions to understand their goal.
2. **Concepts Explanation**: If necessary, break down complex Kubernetes concepts into simpler terms.
3. **Step-by-Step Solution**: Provide a detailed, clear step-by-step process to achieve the desired outcome.
4. **Troubleshooting**: Suggest potential solutions for common issues and pitfalls when working with Kubernetes.
5. **Best Practices**: Mention any relevant Kubernetes best practices that should be followed.

# Output Format

Provide a concise Markdown response in a clear, logical order. Each step should be concise, using bullet points or numbered lists if necessary. Include code snippets in markdown code blocks where relevant.

# Notes

- Assume the user has basic knowledge of Kubernetes.
- Use precise terminology and include explanations only as needed based on the complexity of the task.
- Ensure instructions are applicable across major cloud providers (GKE, EKS, AKS) unless specified otherwise.`

// AssistantFlow runs a simple workflow by following the given instructions.
func AssistantFlow(model string, instructions string, verbose bool) (string, error) {
	assistantFlow := &swarm.Workflow{
		Name:     "assistant-workflow",
		Model:    model,
		MaxTurns: 30,
		Verbose:  verbose,
		System:   "You are an expert on Kubernetes helping user for the given instructions.",
		Steps: []swarm.WorkflowStep{
			{
				Name:         "assistant",
				Instructions: assistantPrompt,
				Inputs: map[string]interface{}{
					"instructions": instructions,
				},
			},
		},
	}

	// Create OpenAI client
	client, err := NewSwarm()
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Initialize and run workflow
	assistantFlow.Initialize()
	result, _, err := assistantFlow.Run(context.Background(), client)
	if err != nil {
		return "", err
	}

	return result, nil
}
