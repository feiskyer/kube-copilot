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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/feiskyer/kube-copilot/pkg/tools"
	"github.com/feiskyer/swarm-go"
)

const planPrompt = `
You are an expert Planning Agent tasked with solving Kubernetes and cloud-native networking problems efficiently through structured plans.
Your job is to:

1. Analyze the user's instruction and their intent carefully to understand the issue or goal.
2. Create a clear and actionable plan to achieve the goal and user intent. Document this plan in the 'steps' field as a structured array.
3. For any troubleshooting step that requires tool execution, include a function call by populating the 'action' field with:
   - 'name': one of [kubectl, python, trivy].
   - 'input': the exact command or script, including any required context (e.g., raw YAML, error logs, image name).
4. Track progress and adapt plans when necessary
5. Do not set the 'final_answer' field when a tool call is pending; only set 'final_answer' when no further tool calls are required.


# Available Tools

- kubectl: Execute Kubernetes commands. Use options like '--sort-by=memory' or '--sort-by=cpu' with 'kubectl top' when necessary and user '--all-namespaces' for cluster-wide information. Input: a single kubectl command (multiple commands are not supported). Output: the command result.
- python: Run Python scripts that leverage the Kubernetes Python SDK client. Ensure that output is generated using 'print(...)'. Input: a Python script (multiple scripts are not supported). Output: the stdout and stderr.
- trivy: Scan container images for vulnerabilities using the 'trivy image' command. Input: an image name. Output: a report of vulnerabilities.

# Output Format

Your final output must strictly adhere to this JSON structure:

{
  "question": "<input question>",
  "thought": "<your detailed thought process>",
  "steps": [
    {
      "name": "<descriptive name of step 1>",
      "description": "<detailed description of what this step will do>",
	  "action": {
		"name": "<tool to call for current step: kubectl, python, or trivy>",
		"input": "<exact command or script with all required context>"
		},
       "status": "<one of: pending, in_progress, completed, failed>",
	  "observation": "<result from the tool call of the action, to be filled in after action execution>",
    },
    {
      "name": "<descriptive name of step 2>",
      "description": "<detailed description of what this step will do>",
	  "action": {
		"name": "<tool to call for current step: kubectl, python, or trivy>",
		"input": "<exact command or script with all required context>"
		},
	  "observation": "<result from the tool call of the action, to be filled in after action execution>",
      "status": "<status of this step>"
    },
    ...more steps...
  ],
  "current_step_index": <index of the current step being executed, zero-based>,
  "final_answer": "<your final findings; only fill this when no further actions are required>"
}

# Important:
- Always use function calls via the 'action' field for tool invocations. NEVER output plain text instructions for the user to run a command manually.
- Ensure that the chain-of-thought (fields 'thought' and 'steps') is clear and concise, leading logically to the tool call if needed.
- The final answer should only be provided when all necessary tool invocations have been completed and the issue is fully resolved.
- The 'steps' array should contain ALL steps needed to solve the problem, with appropriate status updates as you progress.
- NEVER remove steps from the 'steps' array once added, only update their status.
- Initial step statuses should be "pending", change to "in_progress" when starting a step, and then "completed" or "failed" when done.
`

const nextStepPrompt = `You are an expert Planning Agent tasked with solving Kubernetes and cloud-native networking problems efficiently through structured plans.
Your job is to:

1. Review the tool execution results and the current plan.
2. Determine if the plan is sufficient, or if it needs refinement.
3. Choose the most efficient path forward and update the plan accordingly (e.g. update the action inputs for next step or add new steps).
4. If the task is complete, set 'final_answer' right away.

Be concise in your reasoning, then select the appropriate tool or action.

# Output Format

Your final output must strictly adhere to this JSON structure:

{
  "question": "<input question>",
  "thought": "<your detailed thought process>",
  "steps": [
    {
      "name": "<descriptive name of step 1>",
      "description": "<detailed description of what this step will do>",
	  "action": {
		"name": "<tool to call for current step: kubectl, python, or trivy>",
		"input": "<exact command or script with all required context>"
		},
       "status": "<one of: pending, in_progress, completed, failed>",
	  "observation": "<result from the tool call of the action, to be filled in after action execution>",
    },
    {
      "name": "<descriptive name of step 2>",
      "description": "<detailed description of what this step will do>",
	  "action": {
		"name": "<tool to call for current step: kubectl, python, or trivy>",
		"input": "<exact command or script with all required context>"
		},
	  "observation": "<result from the tool call of the action, to be filled in after action execution>",
      "status": "<status of this step>"
    },
    ...more steps...
  ],
  "current_step_index": <index of the current step being executed, zero-based>,
  "final_answer": "<your final findings; only fill this when no further actions are required>"
}
`

const reactPrompt = `As a technical expert in Kubernetes and cloud-native networking, you are required to help user to resolve their problem using a detailed chain-of-thought methodology.
Your responses must follow a strict JSON format and simulate tool execution via function calls without instructing the user to manually run any commands.

# Available Tools

- kubectl: Execute Kubernetes commands. Use options like '--sort-by=memory' or '--sort-by=cpu' with 'kubectl top' when necessary and user '--all-namespaces' for cluster-wide information. Input: a single kubectl command (multiple commands are not supported). Output: the command result.
- python: Run Python scripts that leverage the Kubernetes Python SDK client. Ensure that output is generated using 'print(...)'. Input: a Python script (multiple scripts are not supported). Output: the stdout and stderr.
- trivy: Scan container images for vulnerabilities using the 'trivy image' command. Input: an image name. Output: a report of vulnerabilities.

# Guidelines

1. Analyze the user's instruction and their intent carefully to understand the issue or goal.
2. Formulate a detailed, step-by-step plan to achieve the goal and user intent. Document this plan in the 'steps' field as a structured array.
3. For any troubleshooting step that requires tool execution, include a function call by populating the 'action' field with:
   - 'name': one of [kubectl, python, trivy].
   - 'input': the exact command or script, including any required context (e.g., raw YAML, error logs, image name).
4. DO NOT instruct the user to manually run any commands. All tool calls must be performed by the assistant through the 'action' field.
5. After a tool is invoked, analyze its result (which will be provided in the 'observation' field) and update your chain-of-thought accordingly.
6. Do not set the 'final_answer' field when a tool call is pending; only set 'final_answer' when no further tool calls are required.
7. Maintain a clear and concise chain-of-thought in the 'thought' field. Include a detailed, step-by-step process in the 'steps' field.
8. Your entire response must be a valid JSON object with exactly the following keys: 'question', 'thought', 'steps', 'current_step_index', 'action', 'observation', and 'final_answer'. Do not include any additional text or markdown formatting.

# Output Format

Your final output must strictly adhere to this JSON structure:

{
  "question": "<input question>",
  "thought": "<your detailed thought process>",
  "steps": [
    {
      "name": "<descriptive name of step 1>",
      "description": "<detailed description of what this step will do>",
	  "action": {
		"name": "<tool to call for current step: kubectl, python, or trivy>",
		"input": "<exact command or script with all required context>"
		},
       "status": "<one of: pending, in_progress, completed, failed>",
	  "observation": "<result from the tool call of the action, to be filled in after action execution>",
    },
    {
      "name": "<descriptive name of step 2>",
      "description": "<detailed description of what this step will do>",
	  "action": {
		"name": "<tool to call for current step: kubectl, python, or trivy>",
		"input": "<exact command or script with all required context>"
		},
	  "observation": "<result from the tool call of the action, to be filled in after action execution>",
      "status": "<status of this step>"
    },
    ...more steps...
  ],
  "current_step_index": <index of the current step being executed, zero-based>,
  "final_answer": "<your final findings; only fill this when no further actions are required>"
}

# Important:
- Always use function calls via the 'action' field for tool invocations. NEVER output plain text instructions for the user to run a command manually.
- Ensure that the chain-of-thought (fields 'thought' and 'steps') is clear and concise, leading logically to the tool call if needed.
- The final answer should only be provided when all necessary tool invocations have been completed and the issue is fully resolved.
- The 'steps' array should contain ALL steps needed to solve the problem, with appropriate status updates as you progress.
- NEVER remove steps from the 'steps' array once added, only update their status.
- Initial step statuses should be "pending", change to "in_progress" when starting a step, and then "completed" or "failed" when done.

Follow these instructions strictly to ensure a seamless, automated diagnostic and troubleshooting process.
`

// ReactAction is the JSON format for the react action.
type ReactAction struct {
	Question         string       `json:"question"`
	Thought          string       `json:"thought,omitempty"`
	Steps            []StepDetail `json:"steps,omitempty"`
	CurrentStepIndex int          `json:"current_step_index,omitempty"`
	FinalAnswer      string       `json:"final_answer,omitempty"`
}

// StepDetail represents a detailed step in the plan
type StepDetail struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Action      struct {
		Name  string `json:"name"`
		Input string `json:"input"`
	} `json:"action,omitempty"`
	Observation string `json:"observation,omitempty"`
	Status      string `json:"status"` // pending, in_progress, completed, failed
}

// PlanTracker keeps track of the execution plan and its progress
type PlanTracker struct {
	PlanID           string        `json:"plan_id"`
	Steps            []StepDetail  `json:"steps"`
	CurrentStep      int           `json:"current_step"`
	LastError        string        `json:"last_error,omitempty"`
	FinalAnswer      string        `json:"final_answer,omitempty"`
	HasValidPlan     bool          `json:"has_valid_plan"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
}

// NewPlanTracker creates a new plan tracker
func NewPlanTracker() *PlanTracker {
	return &PlanTracker{
		PlanID:           fmt.Sprintf("plan_%d", time.Now().Unix()),
		Steps:            []StepDetail{},
		CurrentStep:      0,
		ExecutionTimeout: 30 * time.Minute,
	}
}

// ParsePlan parses the plan string into structured steps
func (pt *PlanTracker) ParsePlan(planStr string) error {
	if planStr == "" {
		return fmt.Errorf("empty plan string")
	}

	lines := strings.Split(planStr, "\n")
	steps := []StepDetail{}

	stepPattern := regexp.MustCompile(`^(\d+\.|\*|Step \d+:|[-•])\s*(.+)$`)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := stepPattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			description := strings.TrimSpace(matches[2])
			steps = append(steps, StepDetail{
				Name:        fmt.Sprintf("Step %d", i+1),
				Description: description,
				Status:      "pending",
			})
		}
	}

	if len(steps) == 0 {
		// Fallback: If no steps were found, try to extract sentences as steps
		sentencePattern := regexp.MustCompile(`[.!?]+\s+|\n+|$`)
		sentences := sentencePattern.Split(planStr, -1)

		for i, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if sentence != "" && len(sentence) > 10 { // Minimum length for a sentence to be a step
				steps = append(steps, StepDetail{
					Name:        fmt.Sprintf("Step %d", i+1),
					Description: sentence,
					Status:      "pending",
				})
			}
		}
	}

	if len(steps) == 0 {
		return fmt.Errorf("no steps could be extracted from plan")
	}

	pt.Steps = steps
	pt.HasValidPlan = true
	return nil
}

// UpdateStepStatus updates the status of a step
func (pt *PlanTracker) UpdateStepStatus(stepIndex int, status string, toolCall string, result string) {
	if stepIndex >= 0 && stepIndex < len(pt.Steps) {
		pt.Steps[stepIndex].Status = status
		if toolCall != "" {
			pt.Steps[stepIndex].Action.Name = toolCall
		}
		if result != "" {
			// Truncate long results to prevent memory issues
			pt.Steps[stepIndex].Observation = result
		}
	}
}

// GetCurrentStep returns the current step
func (pt *PlanTracker) GetCurrentStep() *StepDetail {
	if pt.CurrentStep >= 0 && pt.CurrentStep < len(pt.Steps) {
		return &pt.Steps[pt.CurrentStep]
	}
	return nil
}

// MoveToNextStep moves to the next step
func (pt *PlanTracker) MoveToNextStep() bool {
	// If we're already at the last step, we can't move forward
	if pt.CurrentStep >= len(pt.Steps)-1 {
		return false
	}

	// Mark the current step as completed if it's not already marked as failed
	if pt.Steps[pt.CurrentStep].Status != "failed" {
		pt.Steps[pt.CurrentStep].Status = "completed"
	}

	// Find the next available step that isn't already completed or failed
	for i := pt.CurrentStep + 1; i < len(pt.Steps); i++ {
		// If step is pending, move to it
		if pt.Steps[i].Status == "pending" {
			pt.CurrentStep = i
			pt.Steps[i].Status = "in_progress"
			return true
		}
	}

	// If no pending steps found, consider moving to first incomplete step
	// This handles cases where we might need to retry a step or restructure the plan
	for i := 0; i < len(pt.Steps); i++ {
		if i != pt.CurrentStep && pt.Steps[i].Status != "completed" && pt.Steps[i].Status != "failed" {
			pt.CurrentStep = i
			pt.Steps[i].Status = "in_progress"
			return true
		}
	}

	// If we couldn't find any valid next step, just move to the next one
	// as a fallback (this preserves original behavior)
	pt.CurrentStep++
	if pt.CurrentStep < len(pt.Steps) {
		pt.Steps[pt.CurrentStep].Status = "in_progress"
		return true
	}

	return false
}

// IsComplete returns true if all steps are completed
func (pt *PlanTracker) IsComplete() bool {
	for _, step := range pt.Steps {
		if step.Status != "completed" && step.Status != "failed" {
			return false
		}
	}
	return len(pt.Steps) > 0
}

// GetPlanStatus returns a string representation of the plan status
func (pt *PlanTracker) GetPlanStatus() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Plan ID: %s\n\n", pt.PlanID))

	for i, step := range pt.Steps {
		statusSymbol := "⏳"
		if step.Status == "completed" {
			statusSymbol = "✅"
		} else if step.Status == "in_progress" {
			statusSymbol = "🔄"
		} else if step.Status == "failed" {
			statusSymbol = "❌"
		}

		sb.WriteString(fmt.Sprintf("%s Step %d: %s [%s]\n", statusSymbol, i+1, step.Description, step.Status))
		if step.Observation != "" {
			sb.WriteString(fmt.Sprintf("   Observation: %s\n", strings.ReplaceAll(step.Observation, "\n", " ")))
		}
	}

	return sb.String()
}

// ParsePlanFromReactAction parses the plan from ReactAction
func (pt *PlanTracker) ParsePlanFromReactAction(reactAction *ReactAction) error {
	if reactAction == nil {
		return fmt.Errorf("nil ReactAction provided")
	}

	if len(reactAction.Steps) == 0 {
		// Fallback to parsing from thought if steps is empty
		if reactAction.Thought != "" {
			return pt.ParsePlan(reactAction.Thought)
		}
		return fmt.Errorf("no steps found in ReactAction")
	}

	steps := []StepDetail{}

	for _, step := range reactAction.Steps {
		steps = append(steps, StepDetail{
			Name:        step.Name,
			Description: step.Description,
			Status:      step.Status,
		})
	}

	if len(steps) == 0 {
		return fmt.Errorf("no valid steps could be extracted")
	}

	pt.Steps = steps
	pt.HasValidPlan = true

	// If current step index is specified and valid, use it
	if reactAction.CurrentStepIndex >= 0 && reactAction.CurrentStepIndex < len(steps) {
		pt.CurrentStep = reactAction.CurrentStepIndex
	}

	return nil
}

// SyncStepsWithReactAction synchronizes steps from ReactAction with our tracker
func (pt *PlanTracker) SyncStepsWithReactAction(reactAction *ReactAction) {
	if reactAction == nil || len(reactAction.Steps) == 0 {
		return
	}

	// Update the plan if plan steps updated in reactAction
	for i, step := range reactAction.Steps {
		if step.Status == "" {
			step.Status = "pending"
		}

		if i < len(pt.Steps) {
			if pt.Steps[i].Status != "completed" && pt.Steps[i].Status != "failed" && step.Action.Name != "" {
				pt.Steps[i].Action = step.Action
			}
		} else {
			pt.Steps = append(pt.Steps, step)
		}
	}
}

// ReActFlow orchestrates the ReAct (Reason + Act) workflow
type ReActFlow struct {
	Model         string
	Instructions  string
	Verbose       bool
	MaxIterations int
	PlanTracker   *PlanTracker
	Client        *swarm.Swarm
	ChatHistory   interface{}
}

// NewReActFlow creates a new ReActFlow instance
func NewReActFlow(model string, instructions string, verbose bool, maxIterations int) (*ReActFlow, error) {
	// Create OpenAI client
	client, err := NewSwarm()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %v", err)
	}

	return &ReActFlow{
		Model:         model,
		Instructions:  instructions,
		Verbose:       verbose,
		MaxIterations: maxIterations,
		PlanTracker:   NewPlanTracker(),
		Client:        client,
		ChatHistory:   nil,
	}, nil
}

// Run executes the complete ReAct workflow
func (r *ReActFlow) Run() (string, error) {
	// Set a reasonable default response in case of early failures
	defaultResponse := "I was unable to complete the task due to technical issues. Please try again or simplify your request."

	// Set a context with timeout for the entire flow
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	// Step 1: Create initial plan
	if err := r.Plan(ctx); err != nil {
		r.PlanTracker.LastError = fmt.Sprintf("Planning phase failed: %v", err)
		return defaultResponse, err
	}

	// Step 2: Execute plan steps in a loop
	return r.ExecutePlan(ctx)
}

// Plan creates the initial plan for solving the problem
func (r *ReActFlow) Plan(ctx context.Context) error {
	if r.Verbose {
		color.Blue("Planning phase: creating a detailed plan\n")
	}

	// Initialize the first step to create a plan
	reactFlow := &swarm.SimpleFlow{
		Name:     "plan",
		Model:    r.Model,
		MaxTurns: 30,
		Verbose:  r.Verbose,
		Steps: []swarm.SimpleFlowStep{
			{
				Name:         "plan-step",
				Instructions: planPrompt,
				Inputs: map[string]interface{}{
					"instructions": fmt.Sprintf("First, create a clear and actionable step-by-step plan to solve this problem: %s", r.Instructions),
				},
				// Functions: []swarm.AgentFunction{trivyFunc, kubectlFunc, pythonFunc},
			},
		},
	}

	// Initialize and run workflow
	reactFlow.Initialize()

	result, chatHistory, err := reactFlow.Run(ctx, r.Client)
	if err != nil {
		return err
	}

	// Save chat history for future steps
	r.ChatHistory = limitChatHistory(chatHistory, 20)

	if r.Verbose {
		color.Cyan("Planning phase response:\n%s\n\n", result)
	}

	// Parse the initial plan
	return r.ParsePlanResult(result)
}

// ParsePlanResult parses the planning phase result
func (r *ReActFlow) ParsePlanResult(result string) error {
	var reactAction ReactAction
	if err := json.Unmarshal([]byte(result), &reactAction); err != nil {
		if r.Verbose {
			color.Red("Unable to parse response as JSON: %v\n", err)
		}

		// Attempt a more lenient parsing by handling different formats
		// Check if result contains a plan section we can extract
		planSection := extractPlanSection(result)
		if planSection != "" {
			err = r.PlanTracker.ParsePlan(planSection)
			if err != nil && r.Verbose {
				color.Red("Failed to parse extracted plan: %v\n", err)
			}
		}

		// If we still don't have a valid plan, return an error
		if !r.PlanTracker.HasValidPlan {
			return fmt.Errorf("couldn't create a proper plan")
		}
	} else {
		// Parse plan from the structured ReactAction
		err := r.PlanTracker.ParsePlanFromReactAction(&reactAction)
		if err != nil && r.Verbose {
			color.Red("Failed to parse plan from ReactAction: %v\n", err)

			// Fallback: Try to parse from Thought field if it exists (backwards compatibility)
			if reactAction.Thought != "" {
				err = r.PlanTracker.ParsePlan(reactAction.Thought)
				if err != nil && r.Verbose {
					color.Red("Failed to parse plan from Thought: %v\n", err)
				}
			}
		}

		// Check for final answer
		if reactAction.FinalAnswer != "" {
			r.PlanTracker.FinalAnswer = reactAction.FinalAnswer
		}
	}

	// Verify that we have a valid plan
	if !r.PlanTracker.HasValidPlan || len(r.PlanTracker.Steps) == 0 {
		if r.Verbose {
			color.Red("No valid plan could be created\n")
		}
		return fmt.Errorf("no valid plan could be created")
	}

	if r.Verbose {
		color.Cyan("Extracted plan with %d steps\n", len(r.PlanTracker.Steps))
		color.Cyan("Plan status:\n%s\n", r.PlanTracker.GetPlanStatus())
	}

	return nil
}

// ExecutePlan runs the execution phase of the workflow
func (r *ReActFlow) ExecutePlan(ctx context.Context) (string, error) {
	// Make sure we have a valid plan
	if r.PlanTracker.CurrentStep >= len(r.PlanTracker.Steps) || !r.PlanTracker.HasValidPlan {
		return "", fmt.Errorf("no valid plan to execute")
	}

	// Set execution timeout
	execCtx, execCancel := context.WithTimeout(ctx, r.PlanTracker.ExecutionTimeout)
	defer execCancel()

	// Keep track of iterations
	iteration := 0
	for {
		// Check if we've exceeded the maximum number of iterations
		if iteration >= r.MaxIterations {
			if r.Verbose {
				color.Yellow("Reached maximum number of iterations (%d)\n", r.MaxIterations)
			}
			break
		}

		// Check if we're out of time
		if execCtx.Err() != nil {
			return "", fmt.Errorf("execution timed out after %s", r.PlanTracker.ExecutionTimeout)
		}

		// Check if the plan is complete
		if r.PlanTracker.IsComplete() {
			if r.Verbose {
				color.Green("Plan execution complete\n")
			}
			break
		}

		// Get the current step
		currentStep := r.PlanTracker.GetCurrentStep()
		if currentStep == nil {
			return "", fmt.Errorf("invalid current step")
		}

		// Mark the current step as in progress
		currentStep.Status = "in_progress"
		if r.Verbose {
			color.Blue("[step: %s] %s [%s]\n", currentStep.Name, currentStep.Description, currentStep.Status)
		}

		if err := r.ExecuteStep(execCtx, iteration, currentStep); err != nil {
			r.PlanTracker.LastError = err.Error()
			// Mark the step as failed and try to move to the next step
			r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", "", err.Error())
			// If we can't move to the next step, we're done
			if !r.PlanTracker.MoveToNextStep() {
				return "", fmt.Errorf("plan execution failed: %v", err)
			}
		}

		// Check if we have a final answer
		if r.PlanTracker.FinalAnswer != "" && r.PlanTracker.IsComplete() {
			if r.Verbose {
				color.Green("Final answer: %s\n", r.PlanTracker.FinalAnswer)
			}
			break
		}

		// Increment iteration counter
		iteration++
	}

	// Generate the final summary
	return generateFinalSummary(r.PlanTracker), nil
}

// ExecuteStep executes a single step in the plan
func (r *ReActFlow) ExecuteStep(ctx context.Context, iteration int, currentStep *StepDetail) error {
	// Update step status to in_progress
	r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "in_progress", "", "")
	if r.Verbose {
		color.Blue("[step: %s] Executing step %d - %s\n", currentStep.Name, r.PlanTracker.CurrentStep+1, currentStep.Description)
		color.Cyan("Current plan status:\n%s\n", r.PlanTracker.GetPlanStatus())
	}

	// Think about the step
	stepResult, err := r.ThinkAboutStep(ctx, currentStep)
	if err != nil {
		if r.Verbose {
			color.Red("Error executing step: %v\n", err)
		}
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", "", fmt.Sprintf("Error: %v", err))

		// Try to recover by moving to the next step
		if !r.PlanTracker.MoveToNextStep() {
			r.PlanTracker.LastError = fmt.Sprintf("Step execution failed: %v", err)
			return err
		}
		return nil
	}

	// Parse the step result
	var stepAction ReactAction
	if err = json.Unmarshal([]byte(stepResult), &stepAction); err != nil {
		if r.Verbose {
			color.Red("Unable to parse step response as JSON: %v\n", err)
		}
		// Try to extract a final answer from the raw response
		potentialAnswer := extractAnswerFromText(stepResult)
		if potentialAnswer != "" {
			r.PlanTracker.FinalAnswer = potentialAnswer
		}

		// Mark step as failed
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", "", fmt.Sprintf("Error parsing response: %v", err))
		// Try to move to next step
		if !r.PlanTracker.MoveToNextStep() {
			if r.PlanTracker.FinalAnswer != "" {
				return nil
			}
			return fmt.Errorf("couldn't parse the response for step %d", r.PlanTracker.CurrentStep+1)
		}
		return nil
	}

	// Sync steps from the model's response with our tracker
	r.PlanTracker.SyncStepsWithReactAction(&stepAction)

	// Check if we have a final answer
	if stepAction.FinalAnswer != "" {
		r.PlanTracker.FinalAnswer = stepAction.FinalAnswer
		if r.Verbose {
			color.Cyan("Final answer received: %s\n", r.PlanTracker.FinalAnswer)
		}

		// Mark current step as completed
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "completed", "", "Final answer provided")

		// Mark all previous steps as completed
		for i := 0; i < r.PlanTracker.CurrentStep; i++ {
			if r.PlanTracker.Steps[i].Status != "failed" {
				r.PlanTracker.Steps[i].Status = "completed"
			}
		}

		// If this is the last step, we're done
		if r.PlanTracker.CurrentStep == len(r.PlanTracker.Steps)-1 {
			return nil
		}

		// Move to next step even if we have a final answer but more steps remain
		r.PlanTracker.MoveToNextStep()
		return nil
	}

	// Execute tool if needed
	return r.ExecuteToolIfNeeded(ctx, &stepAction)
}

// ThinkAboutStep uses the LLM to think about how to execute the current step
func (r *ReActFlow) ThinkAboutStep(ctx context.Context, currentStep *StepDetail) (string, error) {
	// Prepare the current ReactAction with updated steps status
	currentReactAction := ReactAction{
		Question:         r.Instructions,
		Thought:          "Executing the next step in the plan",
		Steps:            r.PlanTracker.Steps,
		CurrentStepIndex: r.PlanTracker.CurrentStep,
	}

	// Create a new flow for this step
	currentReactActionJSON, _ := json.MarshalIndent(currentReactAction, "", "  ")
	stepFlow := &swarm.SimpleFlow{
		Name:     "think",
		Model:    r.Model,
		MaxTurns: 30,
		Verbose:  r.Verbose,
		Steps: []swarm.SimpleFlowStep{
			{
				Name:         "think-step",
				Instructions: reactPrompt,
				Inputs: map[string]interface{}{
					"instructions": fmt.Sprintf("User input: %s\n\nCurrent plan and status:\n%s\n\nExecute the current step (index %d) of the plan.",
						r.Instructions, string(currentReactActionJSON), r.PlanTracker.CurrentStep),
					"chatHistory": r.ChatHistory,
				},
				// Functions: []swarm.AgentFunction{trivyFunc, kubectlFunc, pythonFunc},
			},
		},
	}

	// Initialize the workflow for this step
	stepFlow.Initialize()

	// Create a context with timeout for this step
	stepCtx, stepCancel := context.WithTimeout(ctx, 5*time.Minute)
	if r.Verbose {
		color.Blue("[step: %s] Running the step %s\n", currentStep.Name, currentStep.Description)
	}

	stepResult, stepChatHistory, err := stepFlow.Run(stepCtx, r.Client)
	stepCancel() // Cancel the context regardless of result

	// Update chat history
	r.ChatHistory = limitChatHistory(stepChatHistory, 20)
	if r.Verbose && err == nil {
		color.Cyan("[step: %s] Step result:\n%s\n\n", currentStep.Name, stepResult)
	}

	return stepResult, err
}

// ExecuteToolIfNeeded executes a tool if the current step requires it
func (r *ReActFlow) ExecuteToolIfNeeded(ctx context.Context, stepAction *ReactAction) error {
	// Check if we need to execute a tool
	currentStepIndex := stepAction.CurrentStepIndex
	if currentStepIndex < 0 || currentStepIndex >= len(stepAction.Steps) || stepAction.Steps[currentStepIndex].Action.Name == "" {
		// No tool execution needed, mark step as completed
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "completed", "", "Step completed without tool execution")

		// Move to next step
		r.PlanTracker.MoveToNextStep()
		return nil
	}

	// Get current step action
	currentStep := &stepAction.Steps[currentStepIndex]
	observation := r.ExecuteTool(currentStep.Action.Name, currentStep.Action.Input)

	// Process the tool observation
	return r.ProcessToolObservation(ctx, currentStep, observation)
}

// ExecuteTool executes the specified tool and returns the observation
func (r *ReActFlow) ExecuteTool(toolName string, toolInput string) string {
	if r.Verbose {
		color.Blue("Executing tool %s\n", toolName)
		color.Cyan("Invoking %s tool with inputs: \n============\n%s\n============\n\n", toolName, toolInput)
	}

	// Execute the tool with timeout
	toolFunc, ok := tools.CopilotTools[toolName]
	if !ok {
		observation := fmt.Sprintf("Tool %s is not available. Considering switch to other supported tools.", toolName)
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", toolName, observation)
		return observation
	}

	// Execute tool with timeout
	toolResultCh := make(chan struct {
		result string
		err    error
	})

	go func() {
		result, err := toolFunc(toolInput)
		toolResultCh <- struct {
			result string
			err    error
		}{result, err}
	}()

	// Wait for tool execution with timeout
	var observation string
	select {
	case toolResult := <-toolResultCh:
		observation = strings.TrimSpace(toolResult.result)
		if toolResult.err != nil {
			observation = fmt.Sprintf("Tool %s failed with error: %v. Considering refine the inputs for the tool.",
				toolName, toolResult.err)
			r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", toolName, observation)
		} else {
			// Update step with tool call info
			r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "in_progress", toolName, "")
		}
	case <-time.After(r.PlanTracker.ExecutionTimeout):
		observation = fmt.Sprintf("Tool %s execution timed out after %v seconds. Try with a simpler query or different tool.",
			toolName, r.PlanTracker.ExecutionTimeout.Seconds())
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", toolName, observation)
	}

	if r.Verbose {
		color.Cyan("Observation: %s\n\n", observation)
	}

	return observation
}

// ProcessToolObservation processes the observation from a tool execution
func (r *ReActFlow) ProcessToolObservation(ctx context.Context, currentStep *StepDetail, observation string) error {
	// Update stepAction with the observation
	currentStep.Observation = observation

	// Create a new flow for processing the observation
	observationActionJSON, _ := json.MarshalIndent(currentStep, "", "  ")
	observationFlow := &swarm.SimpleFlow{
		Name:     "tool-call",
		Model:    r.Model,
		MaxTurns: 30,
		Verbose:  r.Verbose,
		Steps: []swarm.SimpleFlowStep{
			{
				Name:         "tool-call-step",
				Instructions: nextStepPrompt,
				Inputs: map[string]interface{}{
					"instructions": fmt.Sprintf("User input: %s\n\nCurrent plan with tool execution result:\n%s\n",
						r.Instructions, string(observationActionJSON)),
					"chatHistory": r.ChatHistory,
				},
				// Functions: []swarm.AgentFunction{trivyFunc, kubectlFunc, pythonFunc},
			},
		},
	}

	// Initialize the workflow for processing the observation
	observationFlow.Initialize()

	// Run the observation processing
	obsCtx, obsCancel := context.WithTimeout(ctx, 5*time.Minute)
	if r.Verbose {
		color.Blue("[step: %s] Processing tool observation\n", currentStep.Name)
	}

	observationResult, observationChatHistory, err := observationFlow.Run(obsCtx, r.Client)
	obsCancel() // Cancel the context regardless of result

	if err != nil {
		if r.Verbose {
			color.Red("Error processing observation: %v\n", err)
		}
		// Mark step with the appropriate status based on tool execution
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, currentStep.Status, currentStep.Action.Name, observation)

		// Try to move to the next step regardless of the error
		r.PlanTracker.MoveToNextStep()
		return nil
	}

	// Update bounded chat history
	r.ChatHistory = limitChatHistory(observationChatHistory, 20)
	if r.Verbose {
		color.Cyan("[step: %s] Observation processing response:\n%s\n\n", currentStep.Name, observationResult)
	}

	// Parse the observation result
	var observationAction ReactAction
	if err = json.Unmarshal([]byte(observationResult), &observationAction); err != nil {
		if r.Verbose {
			color.Red("Unable to parse observation response as JSON: %v\n", err)
		}
		// Try to extract a final answer from the raw response
		potentialAnswer := extractAnswerFromText(observationResult)
		if potentialAnswer != "" {
			r.PlanTracker.FinalAnswer = potentialAnswer
		}

		// Mark step with the determined status and move on
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, currentStep.Status, currentStep.Action.Name, observation)
		r.PlanTracker.MoveToNextStep()
		return nil
	}

	// Sync steps from observation action with our tracker, but prevent marking multiple steps as in_progress
	r.PlanTracker.SyncStepsWithReactAction(&observationAction)

	// Ensure only one step is in_progress at a time
	for i := range r.PlanTracker.Steps {
		if i != r.PlanTracker.CurrentStep && r.PlanTracker.Steps[i].Status == "in_progress" {
			r.PlanTracker.Steps[i].Status = "pending"
		}
	}

	// Check if we have a final answer from observation processing
	if observationAction.FinalAnswer != "" && r.PlanTracker.IsComplete() {
		r.PlanTracker.FinalAnswer = observationAction.FinalAnswer
		if r.Verbose {
			color.Cyan("Final answer received from observation processing: %s\n", r.PlanTracker.FinalAnswer)
		}

		// Mark current step with the determined status
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "completed", currentStep.Action.Name, observation)

		// If this is the last step, we're done
		if r.PlanTracker.CurrentStep == len(r.PlanTracker.Steps)-1 {
			return nil
		}

		// Move to next step
		r.PlanTracker.MoveToNextStep()
		return nil
	}

	// Check if we need another action from observationAction's current step
	observationStepIndex := observationAction.CurrentStepIndex
	if observationStepIndex >= 0 && observationStepIndex < len(observationAction.Steps) &&
		observationAction.Steps[observationStepIndex].Action.Name != "" {
		// If the current step should retry with a new action, don't mark it as completed yet
		// But update its status in case it was marked as failed by the model
		if observationStepIndex == r.PlanTracker.CurrentStep {
			r.PlanTracker.Steps[r.PlanTracker.CurrentStep].Status = "in_progress"
			return nil // Continue with the current step but with new action
		}
	}

	// If we have a next step, mark current with the determined status and move on
	r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "completed", currentStep.Action.Name, observation)
	r.PlanTracker.MoveToNextStep()
	return nil
}

// extractPlanSection attempts to extract a plan section from unstructured text
func extractPlanSection(text string) string {
	// Look for common plan section indicators
	planPatterns := []string{
		`(?i)(?:^|\n)(?:plan|steps|procedure|approach)(?::|$).*?(?:\n\n|\z)`,
		`(?i)(?:^|\n)(?:I will|Let me|Here's how|First|To solve this).*?(?:\n\n|\z)`,
		`(?i)(?:^|\n)(?:\d+\.|Step \d+:).*?(?:\n\n|\z)`,
	}

	for _, pattern := range planPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindString(text)
		if matches != "" {
			return matches
		}
	}

	// If no plan section found, return the entire text as a fallback
	return text
}

// extractAnswerFromText attempts to extract a final answer from unstructured text
func extractAnswerFromText(text string) string {
	// Look for common answer patterns
	answerPatterns := []string{
		`(?i)(?:^|\n)(?:answer|conclusion|result|summary):?\s*(.+?)(?:\n\n|\z)`,
		`(?i)(?:^|\n)(?:finally|in conclusion|to summarize|in summary):?\s*(.+?)(?:\n\n|\z)`,
		`(?i)(?:^|\n)(?:the solution is|the result is|we found that):?\s*(.+?)(?:\n\n|\z)`,
	}

	for _, pattern := range answerPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// If no answer pattern found, take the last paragraph as a fallback
	paragraphs := strings.Split(text, "\n\n")
	if len(paragraphs) > 0 {
		return paragraphs[len(paragraphs)-1]
	}

	return text
}

// generateFinalSummary creates a summary from all completed steps
func generateFinalSummary(pt *PlanTracker) string {
	if pt.FinalAnswer != "" {
		return pt.FinalAnswer
	}

	var sb strings.Builder
	sb.WriteString("I've completed all the steps in the plan. Here's a summary of what I did:\n\n")

	for i, step := range pt.Steps {
		sb.WriteString(fmt.Sprintf("Step %d: %s [status: %s]\n", i+1, step.Description, step.Status))
		if step.Observation != "" {
			sb.WriteString(fmt.Sprintf("Observation: %s\n\n", step.Observation))
		} else {
			sb.WriteString("\n")
		}

	}

	return sb.String()
}

// limitChatHistory ensures chat history doesn't grow too large
func limitChatHistory(history interface{}, maxMessages int) interface{} {
	if history == nil {
		return nil
	}

	// Handle map type history
	if mapHistory, ok := history.(map[string]interface{}); ok {
		// Create a deep copy to avoid modifying the original
		result := make(map[string]interface{})
		for k, v := range mapHistory {
			result[k] = v
		}

		// If there's a messages array, limit its size
		if messages, ok := result["messages"].([]interface{}); ok && len(messages) > maxMessages {
			result["messages"] = messages[len(messages)-maxMessages:]
		}

		return result
	}

	// Handle slice type history
	if sliceHistory, ok := history.([]map[string]interface{}); ok {
		if len(sliceHistory) <= maxMessages {
			return sliceHistory
		}

		// Take only the last maxMessages items
		return sliceHistory[len(sliceHistory)-maxMessages:]
	}

	// Return unchanged if unknown type
	return history
}
