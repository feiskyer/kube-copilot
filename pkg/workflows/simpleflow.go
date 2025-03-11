/*
Copyright 2023 - Present, Pengfei Ni

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

// SimpleFlow runs a simple workflow by following the given instructions.
func SimpleFlow(model string, systemPrompt string, instructions string, verbose bool) (string, error) {
	simpleFlow := &swarm.SimpleFlow{
		Name:     "simple-workflow",
		Model:    model,
		MaxTurns: 30,
		// Verbose:  verbose,
		Steps: []swarm.SimpleFlowStep{
			{
				Name:         "simple",
				Instructions: systemPrompt,
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
	simpleFlow.Initialize()
	result, _, err := simpleFlow.Run(context.Background(), client)
	if err != nil {
		return "", err
	}

	return result, nil
}

// AssistantFlow runs a simple workflow with kubernetes assistant prompt.
func AssistantFlow(model string, instructions string, verbose bool) (string, error) {
	return SimpleFlow(model, assistantPrompt, instructions, verbose)
}
