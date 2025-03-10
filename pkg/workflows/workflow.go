package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/feiskyer/kube-copilot/pkg/tools"
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

const reactPrompt = `As a technical expert in Kubernetes and cloud-native networking, your task follows a specific Chain of Thought methodology to ensure thoroughness and accuracy while adhering to the constraints provided.

# Available Tools

- kubectl: Useful for executing kubectl commands. Remember to use '--sort-by=memory' or '--sort-by=cpu' when running 'kubectl top' command.  Input: a kubectl command. Output: the result of the command.
- python: This is a Python interpreter. Use it for executing Python code with the Kubernetes Python SDK client. Ensure the results are output using "print(...)". The input is a Python script, and the output will be the stdout and stderr of this script.
- trivy: Useful for executing trivy image command to scan images for vulnerabilities. Input: an image for security scanning. Output: the vulnerabilities found in the image.

# Guidelines

Let's think step by step and reason through the problem.

1. Carefully analyze the user's instructions or questions to understand their goal.
2. If necessary, break down complex Kubernetes operations into a detailed, clear step-by-step process to achieve the desired outcome. Ensure 'action' is set when the step need to call a tool.
3. Call the appropriate tools to get the information needed to solve the problem.
4. Analyze the results of the tools and provide a detailed, clear step-by-step process to achieve the desired outcome.
5. Iterate the above steps if there are still some issues to solve (e.g. need to run kubectl to retrieve more information).
6. Set 'final_answer' when the problem is solved.

Throughout your response process, ensure that each response is concise and strictly adheres to the guidelines provided, with a clear justification for each step taken. 
The ultimate goal is to identify the root cause of issues within the domains of Kubernetes and cloud-native networking and to provide clear, actionable solutions, 
while staying within the operational constraints of 'kubectl', 'python', or 'trivy image' for diagnostics and troubleshooting and avoiding any installation operations.

# Output Format

Use this JSON format for responses:

{
	"question": "<input question>",
	"thought": "<your thought process>",
	"action": {
		"name": "<action to take, choose from tools [kubectl, python, trivy]. Do not set final_answer when an action is required>",
		"input": "<input for the action. ensure all contexts are added as input if required, e.g. raw YAML or image name.>"
	},
	"observation": "<result of the action, set by external tools>",
	"final_answer": "<your final findings, only set after completed all processes and no action is required>"
}
`

// AssistantFlow runs a simple workflow by following the given instructions.
func AssistantFlow(model string, instructions string, verbose bool) (string, error) {
	assistantFlow := &swarm.SimpleFlow{
		Name:     "assistant-workflow",
		Model:    model,
		MaxTurns: 30,
		Verbose:  verbose,
		System:   "You are an expert on Kubernetes helping user for the given instructions.",
		Steps: []swarm.SimpleFlowStep{
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

// ReActFlow runs a ReAct workflow by following the given instructions.
func ReActFlow(model string, instructions string, verbose bool) (string, error) {
	reactFlow := &swarm.SimpleFlow{
		Name:     "react-workflow",
		Model:    model,
		MaxTurns: 30,
		Verbose:  verbose,
		System:   reactPrompt,
		Steps: []swarm.SimpleFlowStep{
			{
				Name:         "react",
				Instructions: reactPrompt,
				Inputs: map[string]interface{}{
					"instructions": instructions,
				},
				Functions: []swarm.AgentFunction{trivyFunc, kubectlFunc, pythonFunc},
			},
		},
	}
	// Initialize and run workflow
	reactFlow.Initialize()

	// Create OpenAI client
	client, err := NewSwarm()
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	iterations := 0
	const maxIterations = 30
	for {
		iterations++
		if verbose {
			color.Blue("Iteration %d): chatting with LLM\n", iterations)
		}

		// Run the workflow
		result, chatHistory, err := reactFlow.Run(context.Background(), client)
		if err != nil {
			return "", err
		}
		if verbose {
			color.Cyan("Iteration %d): response from LLM:\n%s\n\n", iterations, result)
		}

		var toolPrompt ToolPrompt
		if err = json.Unmarshal([]byte(result), &toolPrompt); err != nil {
			if verbose {
				color.Cyan("Unable to parse tool from prompt, assuming got final answer.\n\n", result)
			}
			return result, nil
		}
		if verbose {
			color.Cyan("Thought: %s\n\n", toolPrompt.Thought)
		}

		if iterations > maxIterations {
			color.Red("Max iterations reached")
			return toolPrompt.FinalAnswer, nil
		}

		if toolPrompt.FinalAnswer != "" {
			if verbose {
				color.Cyan("Final answer: %v\n\n", toolPrompt.FinalAnswer)
			}
			return toolPrompt.FinalAnswer, nil
		}

		if toolPrompt.Action.Name != "" {
			var observation string
			if verbose {
				color.Blue("Iteration %d): executing tool %s\n", iterations, toolPrompt.Action.Name)
				color.Cyan("Invoking %s tool with inputs: \n============\n%s\n============\n\n", toolPrompt.Action.Name, toolPrompt.Action.Input)
			}
			if toolFunc, ok := tools.CopilotTools[toolPrompt.Action.Name]; ok {
				ret, err := toolFunc(toolPrompt.Action.Input)
				observation = strings.TrimSpace(ret)
				if err != nil {
					observation = fmt.Sprintf("Tool %s failed with error %s. Considering refine the inputs for the tool.", toolPrompt.Action.Name, ret)
				}
			} else {
				observation = fmt.Sprintf("Tool %s is not available. Considering switch to other supported tools.", toolPrompt.Action.Name)
			}
			if verbose {
				color.Cyan("Observation: %s\n\n", observation)
			}

			// Start next iteration of LLM chat.
			toolPrompt.Observation = observation
			assistantMessage, _ := json.Marshal(toolPrompt)

			reactFlow = &swarm.SimpleFlow{
				Name:     "react-workflow",
				Model:    model,
				MaxTurns: 30,
				Verbose:  verbose,
				System:   reactPrompt,
				Steps: []swarm.SimpleFlowStep{
					{
						Name:         "react",
						Instructions: reactPrompt,
						Inputs: map[string]interface{}{
							"instructions": fmt.Sprintf("User input: %s\n\nObservation: %s", instructions, string(assistantMessage)),
							"chatHistory":  chatHistory,
						},
						Functions: []swarm.AgentFunction{trivyFunc, kubectlFunc, pythonFunc},
					},
				},
			}
			// Initialize the workflow for the next iteration
			reactFlow.Initialize()
		}
	}
}
