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

const reactPrompt = `As a technical expert in Kubernetes and cloud-native networking, you are required to help user to resolve their problem using a detailed chain-of-thought methodology.
Your responses must follow a strict JSON format and simulate tool execution via function calls without instructing the user to manually run any commands.

# Available Tools

- kubectl: Execute Kubernetes commands. Use options like '--sort-by=memory' or '--sort-by=cpu' with 'kubectl top' when necessary. Input: a kubectl command. Output: the command result.
- python: Run Python scripts that leverage the Kubernetes Python SDK client. Ensure that output is generated using 'print(...)'. Input: a Python script. Output: the stdout and stderr.
- trivy: Scan container images for vulnerabilities using the 'trivy image' command. Input: an image name. Output: a report of vulnerabilities.

# Guidelines

1. Analyze the user's instruction and their intent carefully to understand the issue or goal.
2. Formulate a detailed, step-by-step plan to achieve the goal and user intent. Document this plan in the 'plan' field.
3. For any troubleshooting step that requires tool execution, include a function call by populating the 'action' field with:
   - 'name': one of [kubectl, python, trivy].
   - 'input': the exact command or script, including any required context (e.g., raw YAML, error logs, image name).
4. DO NOT instruct the user to manually run any commands. All tool calls must be performed by the assistant through the 'action' field.
5. After a tool is invoked, analyze its result (which will be provided in the 'observation' field) and update your chain-of-thought accordingly.
6. Do not set the 'final_answer' field when a tool call is pending; only set 'final_answer' when no further tool calls are required.
7. Maintain a clear and concise chain-of-thought in the 'thought' field. Include a detailed, step-by-step process in the 'plan' field.
8. Your entire response must be a valid JSON object with exactly the following keys: 'question', 'thought', 'plan', 'current_step', 'action', 'observation', and 'final_answer'. Do not include any additional text or markdown formatting.

# Output Format

Your final output must strictly adhere to this JSON structure:

{
  "question": "<input question>",
  "plan": "<step-by-step plan to achieve the desired outcome>",
  "thought": "<your detailed thought process>",
  "next_step": "<the next step to be executed>",
  "action": {
      "name": "<tool to call for next step: kubectl, python, or trivy>",
      "input": "<exact command or script with all required context>"
  },
  "observation": "<result from the tool call of the action, to be filled in after action execution>",
  "final_answer": "<your final findings; only fill this when no further actions are required>"
}

# Important:
- Always use function calls via the 'action' field for tool invocations. NEVER output plain text instructions for the user to run a command manually.
- Ensure that the chain-of-thought (fields 'thought' and 'plan') is clear and concise, leading logically to the tool call if needed.
- The final answer should only be provided when all necessary tool invocations have been completed and the issue is fully resolved.

Follow these instructions strictly to ensure a seamless, automated diagnostic and troubleshooting process.
`

// SimpleFlow runs a simple workflow by following the given instructions.
func SimpleFlow(model string, systemPrompt string, instructions string, verbose bool) (string, error) {
	simpleFlow := &swarm.SimpleFlow{
		Name:     "simple-workflow",
		Model:    model,
		MaxTurns: 30,
		Verbose:  verbose,
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

// ReactAction is the JSON format for the react action.
type ReactAction struct {
	Question string `json:"question"`
	Plan     string `json:"plan,omitempty"`
	Thought  string `json:"thought,omitempty"`
	NextStep string `json:"next_step,omitempty"`
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
		Steps: []swarm.SimpleFlowStep{
			{
				Name:         "react",
				Instructions: reactPrompt,
				Inputs: map[string]interface{}{
					"instructions": instructions,
				},
				// Functions: []swarm.AgentFunction{trivyFunc, kubectlFunc, pythonFunc},
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

		var reactAction ReactAction
		if err = json.Unmarshal([]byte(result), &reactAction); err != nil {
			if verbose {
				color.Cyan("Unable to parse tool from prompt, assuming got final answer.\n\n", result)
			}
			return result, nil
		}
		if verbose {
			color.Cyan("Thought: %s\n\n", reactAction.Thought)
		}

		if iterations > maxIterations {
			color.Red("Max iterations reached")
			return reactAction.FinalAnswer, nil
		}

		if reactAction.FinalAnswer != "" {
			if verbose {
				color.Cyan("Final answer: %v\n\n", reactAction.FinalAnswer)
			}
			return reactAction.FinalAnswer, nil
		}

		if reactAction.Action.Name != "" {
			var observation string
			if verbose {
				color.Blue("Iteration %d): executing tool %s\n", iterations, reactAction.Action.Name)
				color.Cyan("Invoking %s tool with inputs: \n============\n%s\n============\n\n", reactAction.Action.Name, reactAction.Action.Input)
			}
			if toolFunc, ok := tools.CopilotTools[reactAction.Action.Name]; ok {
				ret, err := toolFunc(reactAction.Action.Input)
				observation = strings.TrimSpace(ret)
				if err != nil {
					observation = fmt.Sprintf("Tool %s failed with error %s. Considering refine the inputs for the tool.", reactAction.Action.Name, ret)
				}
			} else {
				observation = fmt.Sprintf("Tool %s is not available. Considering switch to other supported tools.", reactAction.Action.Name)
			}
			if verbose {
				color.Cyan("Observation: %s\n\n", observation)
			}

			// Start next iteration of LLM chat.
			reactAction.Observation = observation
			assistantMessage, _ := json.Marshal(reactAction)

			reactFlow = &swarm.SimpleFlow{
				Name:     "react-workflow",
				Model:    model,
				MaxTurns: 30,
				Verbose:  verbose,
				Steps: []swarm.SimpleFlowStep{
					{
						Name:         "react",
						Instructions: reactPrompt,
						Inputs: map[string]interface{}{
							"instructions": fmt.Sprintf("User input: %s\n\nObservation: %s", instructions, string(assistantMessage)),
							"chatHistory":  chatHistory,
						},
						// Functions: []swarm.AgentFunction{trivyFunc, kubectlFunc, pythonFunc},
					},
				},
			}
			// Initialize the workflow for the next iteration
			reactFlow.Initialize()
		}
	}
}
