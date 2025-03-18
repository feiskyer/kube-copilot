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

// Package workflows provides the ReAct (Reason + Act) workflow for AI assistants.
package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/feiskyer/kube-copilot/pkg/llms"
	"github.com/feiskyer/kube-copilot/pkg/tools"
	"github.com/feiskyer/swarm-go"
)

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
	if pt.CurrentStep >= len(pt.Steps)-1 && len(pt.Steps) > 0 {
		// Just mark the current (last) step as completed if not already failed
		if pt.Steps[pt.CurrentStep].Status != "failed" {
			pt.Steps[pt.CurrentStep].Status = "completed"
		}
		return false
	}

	// Track the original step index before moving
	originalStep := pt.CurrentStep

	// Mark the current step as completed if it's not already marked as failed
	// and only if we have steps (avoid index out of range)
	if len(pt.Steps) > 0 && originalStep >= 0 && originalStep < len(pt.Steps) {
		if pt.Steps[originalStep].Status != "failed" {
			pt.Steps[originalStep].Status = "completed"
		}
	}

	// First try to find the next pending step
	foundNextStep := false

	// Look for pending steps after the current step first
	for i := originalStep + 1; i < len(pt.Steps); i++ {
		// If step is pending, move to it
		if pt.Steps[i].Status == "pending" {
			pt.CurrentStep = i
			pt.Steps[i].Status = "in_progress"
			foundNextStep = true
			break
		}
	}

	// If no pending steps found after current, check from beginning up to current
	if !foundNextStep {
		for i := 0; i < originalStep; i++ {
			if pt.Steps[i].Status == "pending" {
				pt.CurrentStep = i
				pt.Steps[i].Status = "in_progress"
				foundNextStep = true
				break
			}
		}
	}

	// If still no pending steps, look for in_progress steps that aren't the current one
	if !foundNextStep {
		for i := 0; i < len(pt.Steps); i++ {
			if i != originalStep && pt.Steps[i].Status == "in_progress" {
				pt.CurrentStep = i
				foundNextStep = true
				break
			}
		}
	}

	// If we still couldn't find a pending or in_progress step, take one of two actions:
	if !foundNextStep {
		// If the original step was valid and we were simply moving to the next step in sequence,
		// follow the sequence
		if originalStep >= 0 && originalStep < len(pt.Steps)-1 {
			pt.CurrentStep = originalStep + 1
			// Only mark as in_progress if it's not already completed or failed
			if pt.Steps[pt.CurrentStep].Status != "completed" && pt.Steps[pt.CurrentStep].Status != "failed" {
				pt.Steps[pt.CurrentStep].Status = "in_progress"
			}
			return true
		}

		// If we had an invalid original step or were already at the end,
		// set to the last step in the plan
		if len(pt.Steps) > 0 {
			pt.CurrentStep = len(pt.Steps) - 1
		} else {
			pt.CurrentStep = 0 // Handle empty step list
		}
		return false
	}

	return true
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

func formatObservationOutputs(observation string) string {
	if observation == "" {
		return ""
	}

	lines := strings.Split(observation, "\n")
	formattedLines := make([]string, len(lines))
	for i, line := range lines {
		formattedLines[i] = "   " + line
	}
	return strings.Join(formattedLines, "\n")
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
			formattedObservation := formatObservationOutputs(step.Observation)
			sb.WriteString(fmt.Sprintf("   Observation:\n%s\n", formattedObservation))
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
		// Ensure status is set to a valid value
		status := step.Status
		if status == "" {
			status = "pending"
		}

		// Copy step with proper status
		steps = append(steps, StepDetail{
			Name:        step.Name,
			Description: step.Description,
			Status:      status,
			Action:      step.Action,
			Observation: step.Observation,
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
		// Make sure the current step is marked as in_progress
		if pt.Steps[pt.CurrentStep].Status == "pending" {
			pt.Steps[pt.CurrentStep].Status = "in_progress"
		}
	} else {
		// Default to starting at first step
		pt.CurrentStep = 0
		// Mark first step as in_progress
		if len(pt.Steps) > 0 && pt.Steps[0].Status == "pending" {
			pt.Steps[0].Status = "in_progress"
		}
	}

	// Store final answer if provided
	if reactAction.FinalAnswer != "" {
		pt.FinalAnswer = reactAction.FinalAnswer
	}

	return nil
}

// SyncStepsWithReactAction synchronizes steps from ReactAction with our tracker
func (pt *PlanTracker) SyncStepsWithReactAction(reactAction *ReactAction) {
	if reactAction == nil || len(reactAction.Steps) == 0 {
		return
	}

	// First, ensure our step arrays are of the same length
	// If reactAction has more steps, copy them to our tracker
	for i := len(pt.Steps); i < len(reactAction.Steps); i++ {
		pt.Steps = append(pt.Steps, reactAction.Steps[i])
	}

	// Record the state of steps before updating
	completedSteps := make(map[int]bool)
	failedSteps := make(map[int]bool)
	for i, step := range pt.Steps {
		if step.Status == "completed" {
			completedSteps[i] = true
		} else if step.Status == "failed" {
			failedSteps[i] = true
		}
	}

	// Update existing steps - but preserve completion status
	for i, step := range reactAction.Steps {
		if i < len(pt.Steps) {
			// Keep track of original status for comparison
			originalStatus := pt.Steps[i].Status

			// Don't override completed or failed status from our tracker
			if completedSteps[i] || failedSteps[i] {
				// Only copy additional information without changing status
				if pt.Steps[i].Action.Name == "" && step.Action.Name != "" {
					pt.Steps[i].Action = step.Action
				}

				if pt.Steps[i].Description == "" && step.Description != "" {
					pt.Steps[i].Description = step.Description
				}

				if pt.Steps[i].Name == "" && step.Name != "" {
					pt.Steps[i].Name = step.Name
				}
			} else {
				// For steps that aren't completed or failed, sync all data
				pt.Steps[i].Name = step.Name
				pt.Steps[i].Description = step.Description

				// Only update action if it has content
				if step.Action.Name != "" {
					pt.Steps[i].Action = step.Action
				}

				// Only update status if it's not empty and would be a valid transition
				if step.Status != "" {
					// Don't allow pending→completed without going through in_progress
					if !(originalStatus == "pending" && step.Status == "completed") {
						pt.Steps[i].Status = step.Status
					}
				}

				// Only update observation if not empty
				if step.Observation != "" {
					pt.Steps[i].Observation = step.Observation
				}
			}
		}
	}

	// Sync current step index if it's within bounds
	if reactAction.CurrentStepIndex >= 0 && reactAction.CurrentStepIndex < len(pt.Steps) {
		// We don't want to move backward from a completed step to an earlier step
		// UNLESS that earlier step still needs execution (e.g., has a new action)
		shouldUpdateCurrentStep := false

		// Always update if moving forward
		if reactAction.CurrentStepIndex > pt.CurrentStep {
			shouldUpdateCurrentStep = true
		} else if reactAction.CurrentStepIndex < pt.CurrentStep {
			// Only move backwards if current step is completed/failed AND
			// the target step is not completed/failed AND has an action
			if (pt.Steps[pt.CurrentStep].Status == "completed" ||
				pt.Steps[pt.CurrentStep].Status == "failed") &&
				(pt.Steps[reactAction.CurrentStepIndex].Status != "completed" &&
					pt.Steps[reactAction.CurrentStepIndex].Status != "failed") &&
				pt.Steps[reactAction.CurrentStepIndex].Action.Name != "" {
				shouldUpdateCurrentStep = true
			}
		} else {
			// Same step index, no update needed
		}

		if shouldUpdateCurrentStep {
			pt.CurrentStep = reactAction.CurrentStepIndex

			// Ensure the current step is marked as in_progress
			if pt.Steps[pt.CurrentStep].Status == "pending" {
				pt.Steps[pt.CurrentStep].Status = "in_progress"
			}
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
		MaxIterations: maxIterations,
		PlanTracker:   NewPlanTracker(),
		Verbose:       verbose,
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
		Steps: []swarm.SimpleFlowStep{
			{
				Name:         "plan-step",
				Instructions: planPrompt,
				Inputs: map[string]interface{}{
					"instructions": fmt.Sprintf("First, create a clear and actionable step-by-step plan to solve this problem: %s", r.Instructions),
				},
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

func extractReactAction(text string, reactAction *ReactAction) error {
	text = strings.TrimSpace(text)

	// For responses with prefix ```json
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimPrefix(text, "json")
		text = strings.TrimSuffix(text, "```")
	}

	// For responses with prefix <think>
	if strings.HasPrefix(text, "<think>") {
		text = strings.Split(text, "</think>")[1]
		text = strings.TrimSpace(text)
	}

	if err := json.Unmarshal([]byte(text), reactAction); err != nil {
		return fmt.Errorf("failed to parse LLM response to ReactAction: %v", err)
	}
	return nil
}

// ParsePlanResult parses the planning phase result
func (r *ReActFlow) ParsePlanResult(result string) error {
	var reactAction ReactAction
	if err := extractReactAction(result, &reactAction); err != nil {
		if r.Verbose {
			color.Red("Unable to parse response as JSON: %v\n", err)
		}

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
	if len(r.PlanTracker.Steps) == 0 || !r.PlanTracker.HasValidPlan {
		return "", fmt.Errorf("no valid plan to execute")
	}

	// Set execution timeout
	execCtx, execCancel := context.WithTimeout(ctx, r.PlanTracker.ExecutionTimeout)
	defer execCancel()

	// Keep track of iterations
	iteration := 0

	// Track step stability (detect oscillation)
	previousStepIndices := make([]int, 3) // track last 3 steps to detect oscillation
	for i := range previousStepIndices {
		previousStepIndices[i] = -1 // initialize with invalid indices
	}

	// Initialize first step if needed
	if r.PlanTracker.CurrentStep < 0 || r.PlanTracker.CurrentStep >= len(r.PlanTracker.Steps) {
		r.PlanTracker.CurrentStep = 0
		r.PlanTracker.Steps[0].Status = "in_progress"
	}

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

		// Check for step oscillation (repeating between the same steps)
		oscillationDetected := false
		if iteration >= 3 {
			// Shift previous indices
			previousStepIndices[0] = previousStepIndices[1]
			previousStepIndices[1] = previousStepIndices[2]
			previousStepIndices[2] = r.PlanTracker.CurrentStep

			// Check for A-B-A pattern
			if previousStepIndices[0] == previousStepIndices[2] &&
				previousStepIndices[0] != previousStepIndices[1] &&
				previousStepIndices[0] != -1 {
				oscillationDetected = true
				if r.Verbose {
					color.Red("Oscillation detected between steps %d and %d. Forcing forward progress.",
						previousStepIndices[0]+1, previousStepIndices[1]+1)
				}

				// Force mark the current step as completed to break the cycle
				r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "completed", "",
					"Automatic completion to break oscillation")
			}
		} else {
			// For early iterations, just record the step index
			previousStepIndices[iteration] = r.PlanTracker.CurrentStep
		}

		// Check if the plan is complete - all steps must be completed or failed
		isComplete := true
		for i, step := range r.PlanTracker.Steps {
			// If any step isn't completed or failed, the plan isn't complete
			if step.Status != "completed" && step.Status != "failed" {
				isComplete = false

				// If we find a step with pending or in_progress status that comes before
				// our current step, we should consider moving back to execute it
				if i < r.PlanTracker.CurrentStep && !oscillationDetected {
					// Only move back if we have reason to (it has an action)
					if step.Action.Name != "" {
						if r.Verbose {
							color.Yellow("Found earlier step %d in %s state with action, moving back to execute it",
								i+1, step.Status)
						}
						r.PlanTracker.CurrentStep = i
						r.PlanTracker.Steps[i].Status = "in_progress"
						break
					}
				}
			}
		}

		if isComplete {
			if r.Verbose {
				color.Green("Plan execution complete - all steps are completed or failed\n")
			}
			break
		}

		// Validate current step index is within bounds
		if r.PlanTracker.CurrentStep < 0 || r.PlanTracker.CurrentStep >= len(r.PlanTracker.Steps) {
			if r.Verbose {
				color.Red("Invalid current step index %d. Resetting to first incomplete step.",
					r.PlanTracker.CurrentStep)
			}

			// Find the first non-completed step
			for i, step := range r.PlanTracker.Steps {
				if step.Status != "completed" && step.Status != "failed" {
					r.PlanTracker.CurrentStep = i
					break
				}
			}

			// If all steps are complete/failed but we didn't detect it above, force to last step
			if r.PlanTracker.CurrentStep < 0 || r.PlanTracker.CurrentStep >= len(r.PlanTracker.Steps) {
				r.PlanTracker.CurrentStep = len(r.PlanTracker.Steps) - 1
			}
		}

		// Get the current step
		currentStep := r.PlanTracker.GetCurrentStep()
		if currentStep == nil {
			return "", fmt.Errorf("invalid current step")
		}

		// Mark the current step as in progress if it's pending
		if currentStep.Status == "pending" {
			currentStep.Status = "in_progress"
		}

		if r.Verbose {
			color.Blue("[Step %d: %s] %s [%s]\n", r.PlanTracker.CurrentStep+1,
				currentStep.Name, currentStep.Description, currentStep.Status)
		}

		if err := r.ExecuteStep(execCtx, iteration, currentStep); err != nil {
			r.PlanTracker.LastError = err.Error()
			// Mark the step as failed and try to move to the next step
			r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", "", err.Error())
			// If we can't move to the next step, we're done
			if !r.PlanTracker.MoveToNextStep() {
				if r.PlanTracker.FinalAnswer != "" {
					// If we have a final answer, consider the plan successful anyway
					break
				}
				return "", fmt.Errorf("plan execution failed: %v", err)
			}
		}

		// Check if we've reached our last step
		if r.PlanTracker.CurrentStep >= len(r.PlanTracker.Steps)-1 {
			// Only exit if the last step is completed or failed
			lastStep := r.PlanTracker.Steps[len(r.PlanTracker.Steps)-1]
			if lastStep.Status == "completed" || lastStep.Status == "failed" {
				if r.PlanTracker.FinalAnswer != "" {
					if r.Verbose {
						color.Green("Final answer: %s\n", r.PlanTracker.FinalAnswer)
					}
				} else {
					if r.Verbose {
						color.Yellow("No final answer provided, but plan execution is complete.\n")
					}
				}
				break
			} else {
				// Last step isn't completed yet, continue execution
				if r.Verbose {
					color.Yellow("At last step but status is %s, continuing execution", lastStep.Status)
				}
			}
		}

		// Increment iteration counter
		iteration++
	}

	// Generate the final summary
	return generateFinalSummary(r.PlanTracker), nil
}

// ExecuteStep executes a single step in the plan
func (r *ReActFlow) ExecuteStep(ctx context.Context, iteration int, currentStep *StepDetail) error {
	// Validate we have steps to execute
	if len(r.PlanTracker.Steps) == 0 {
		return fmt.Errorf("no steps in execution plan")
	}

	// Validate the current step index is within bounds
	if r.PlanTracker.CurrentStep < 0 || r.PlanTracker.CurrentStep >= len(r.PlanTracker.Steps) {
		return fmt.Errorf("current step index %d is out of bounds (0-%d)",
			r.PlanTracker.CurrentStep, len(r.PlanTracker.Steps)-1)
	}

	// Get the current step from our tracker
	trackerCurrentStep := r.PlanTracker.GetCurrentStep()
	if trackerCurrentStep == nil {
		return fmt.Errorf("invalid current step - nil returned from GetCurrentStep")
	}

	// If the passed step doesn't match our tracker's current step, log warning and use our tracker's step
	if currentStep != trackerCurrentStep {
		if r.Verbose {
			color.Yellow("Step mismatch: passed step doesn't match tracker's current step. Using tracker's step.")
		}
		currentStep = trackerCurrentStep
	}

	// Ensure the current step is marked as in_progress
	r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "in_progress", "", "")

	if r.Verbose {
		color.Blue("[Step %d: %s] Executing step [current status: %s]\n",
			r.PlanTracker.CurrentStep+1,
			currentStep.Name,
			currentStep.Description,
			r.PlanTracker.Steps[r.PlanTracker.CurrentStep].Status)
		color.Cyan("Current plan status:\n%s\n", r.PlanTracker.GetPlanStatus())
	}

	// Think about the step
	stepResult, err := r.ThinkAboutStep(ctx, currentStep)
	if err != nil {
		if r.Verbose {
			color.Red("Error thinking about step %d: %v\n", r.PlanTracker.CurrentStep+1, err)
		}
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", "", fmt.Sprintf("Error: %v", err))

		// Try to recover by moving to the next step
		if !r.PlanTracker.MoveToNextStep() {
			r.PlanTracker.LastError = fmt.Sprintf("Step thinking failed: %v", err)
			return err
		}
		return nil
	}

	// Parse the step result
	var stepAction ReactAction
	if err = extractReactAction(stepResult, &stepAction); err != nil {
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

	// Store original step index for comparison
	originalStepIndex := r.PlanTracker.CurrentStep

	// Sync steps from the model's response with our tracker
	r.PlanTracker.SyncStepsWithReactAction(&stepAction)

	// Verify we're still working on a valid step after sync
	if r.PlanTracker.CurrentStep < 0 || r.PlanTracker.CurrentStep >= len(r.PlanTracker.Steps) {
		if r.Verbose {
			color.Red("After step sync, current step index %d is invalid. Resetting to original step %d.",
				r.PlanTracker.CurrentStep, originalStepIndex)
		}
		// Reset to original step if current became invalid
		r.PlanTracker.CurrentStep = originalStepIndex
	}

	// Check if the model and our tracker are on different steps now
	if stepAction.CurrentStepIndex != r.PlanTracker.CurrentStep {
		if r.Verbose {
			color.Yellow("Step index drift after sync: model=%d, tracker=%d",
				stepAction.CurrentStepIndex, r.PlanTracker.CurrentStep)
		}
	}

	// Check if we have a final answer
	if stepAction.FinalAnswer != "" {
		r.PlanTracker.FinalAnswer = stepAction.FinalAnswer
		if r.Verbose {
			color.Cyan("Final answer received: %s\n", r.PlanTracker.FinalAnswer)
		}

		// Mark current step as completed
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "completed", "", "Final answer provided")

		// If this is the last step, we're done
		if r.PlanTracker.CurrentStep >= len(r.PlanTracker.Steps)-1 {
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
	// Validate our current step pointer and index
	if currentStep == nil {
		return "", fmt.Errorf("current step is nil")
	}

	// Validate the current step index is within bounds
	if r.PlanTracker.CurrentStep < 0 || r.PlanTracker.CurrentStep >= len(r.PlanTracker.Steps) {
		return "", fmt.Errorf("current step index %d is out of bounds (0-%d)",
			r.PlanTracker.CurrentStep, len(r.PlanTracker.Steps)-1)
	}

	// Get the current step from our tracker (not the parameter passed in)
	trackerCurrentStep := r.PlanTracker.GetCurrentStep()

	// Double check that our pointer and tracker step are in sync
	if currentStep != trackerCurrentStep {
		if r.Verbose {
			color.Yellow("Current step mismatch in ThinkAboutStep. Using tracker's step.")
		}
		currentStep = trackerCurrentStep
	}

	// Mark current step as in_progress if not already
	if r.PlanTracker.Steps[r.PlanTracker.CurrentStep].Status == "pending" {
		r.PlanTracker.Steps[r.PlanTracker.CurrentStep].Status = "in_progress"
	}

	// Prepare a deep copy of the steps to send to the LLM
	stepsCopy := make([]StepDetail, len(r.PlanTracker.Steps))
	copy(stepsCopy, r.PlanTracker.Steps)

	// Prepare the current ReactAction with updated steps status
	currentReactAction := ReactAction{
		Question:         r.Instructions,
		Thought:          "Executing the next step in the plan",
		Steps:            stepsCopy,
		CurrentStepIndex: r.PlanTracker.CurrentStep,
	}

	// Create a new flow for this step
	currentReactActionJSON, _ := json.MarshalIndent(currentReactAction, "", "  ")
	stepFlow := &swarm.SimpleFlow{
		Name:     "think",
		Model:    r.Model,
		MaxTurns: 30,
		Steps: []swarm.SimpleFlowStep{
			{
				Name:         "think-step",
				Instructions: reactPrompt,
				Inputs: map[string]interface{}{
					"instructions": fmt.Sprintf("User input: %s\n\nCurrent plan and status:\n%s\n\nExecute the current step (index %d) of the plan.",
						r.Instructions, string(currentReactActionJSON), r.PlanTracker.CurrentStep),
					"chatHistory": r.ChatHistory,
				},
			},
		},
	}

	// Initialize the workflow for this step
	stepFlow.Initialize()

	// Create a context with timeout for this step
	stepCtx, stepCancel := context.WithTimeout(ctx, 5*time.Minute)
	if r.Verbose {
		color.Blue("[Step %d: %s] Running the step %s [current status: %s]\n",
			r.PlanTracker.CurrentStep+1,
			currentStep.Name,
			currentStep.Description,
			r.PlanTracker.Steps[r.PlanTracker.CurrentStep].Status)
	}

	stepResult, stepChatHistory, err := stepFlow.Run(stepCtx, r.Client)
	stepCancel()

	// Update chat history
	r.ChatHistory = limitChatHistory(stepChatHistory, 20)
	if r.Verbose && err == nil {
		color.Cyan("[Step %d: %s] Step result:\n%s\n\n", r.PlanTracker.CurrentStep+1, currentStep.Name, stepResult)
	}

	return stepResult, err
}

// ExecuteToolIfNeeded executes a tool if the current step requires it
func (r *ReActFlow) ExecuteToolIfNeeded(ctx context.Context, stepAction *ReactAction) error {
	// Ensure our internal step tracker and the stepAction's index are fully synchronized
	if stepAction.CurrentStepIndex != r.PlanTracker.CurrentStep {
		if r.Verbose {
			color.Yellow("Step index mismatch: PlanTracker.CurrentStep=%d, stepAction.CurrentStepIndex=%d, syncing to PlanTracker's value",
				r.PlanTracker.CurrentStep, stepAction.CurrentStepIndex)
		}
	}

	// Always use our plan tracker's current step as the source of truth
	currentStepIndex := r.PlanTracker.CurrentStep

	// Validate bounds for both tracking mechanisms
	if currentStepIndex < 0 || currentStepIndex >= len(r.PlanTracker.Steps) {
		return fmt.Errorf("invalid current step index: %d (out of bounds)", currentStepIndex)
	}

	// Check if we have a valid action to execute
	actionExists := false
	var actionName, actionInput string

	// First check our plan tracker for action info
	if currentStepIndex < len(r.PlanTracker.Steps) &&
		r.PlanTracker.Steps[currentStepIndex].Action.Name != "" {
		actionExists = true
		actionName = r.PlanTracker.Steps[currentStepIndex].Action.Name
		actionInput = r.PlanTracker.Steps[currentStepIndex].Action.Input
	}

	// If no action in our tracker, check if stepAction provides one
	if !actionExists && currentStepIndex < len(stepAction.Steps) &&
		stepAction.Steps[currentStepIndex].Action.Name != "" {
		actionExists = true
		actionName = stepAction.Steps[currentStepIndex].Action.Name
		actionInput = stepAction.Steps[currentStepIndex].Action.Input

		// Sync this action back to our plan tracker
		r.PlanTracker.Steps[currentStepIndex].Action.Name = actionName
		r.PlanTracker.Steps[currentStepIndex].Action.Input = actionInput
	}

	// If no tool execution needed, mark step as completed and move on
	if !actionExists {
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "completed", "", "Step completed without tool execution")
		r.PlanTracker.MoveToNextStep()
		return nil
	}

	// Execute the tool and get tool response
	toolResponse := r.ExecuteTool(actionName, actionInput)

	// Get the current step - initialize with a stub first
	tempStep := StepDetail{
		Name:        fmt.Sprintf("Step %d", currentStepIndex+1),
		Description: fmt.Sprintf("Executing %s tool", actionName),
		Status:      "in_progress",
	}
	tempStep.Action.Name = actionName
	tempStep.Action.Input = actionInput

	// Only try to use stepAction's step if the index is valid
	currentStep := &tempStep
	if currentStepIndex < len(stepAction.Steps) {
		currentStep = &stepAction.Steps[currentStepIndex]
	}

	// Process the tool observation
	return r.ProcessToolObservation(ctx, currentStep, toolResponse)
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
		toolResponse := fmt.Sprintf("Tool %s is not available. Considering switch to other supported tools.", toolName)
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", toolName, toolResponse)
		return toolResponse
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
	var toolResponse string
	select {
	case toolResult := <-toolResultCh:
		toolResponse = strings.TrimSpace(toolResult.result)
		if toolResult.err != nil {
			toolResponse = fmt.Sprintf("Tool %s failed with result: %s error: %v. Considering refine the inputs for the tool.",
				toolName, toolResult.result, toolResult.err)
			r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", toolName, toolResponse)
		} else {
			// Update step with tool call info
			r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "in_progress", toolName, toolResult.result)
		}
	case <-time.After(r.PlanTracker.ExecutionTimeout):
		toolResponse = fmt.Sprintf("Tool %s execution timed out after %v seconds. Try with a simpler query or different tool.",
			toolName, r.PlanTracker.ExecutionTimeout.Seconds())
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", toolName, toolResponse)
	}

	if toolResponse == "" {
		toolResponse = "Empty result returned from the tool."
	}
	if r.Verbose {
		color.Cyan("Tool %s result:\n %s\n\n", toolName, toolResponse)
	}
	return toolResponse
}

// ProcessToolObservation processes the observation from a tool execution
func (r *ReActFlow) ProcessToolObservation(ctx context.Context, currentStep *StepDetail, toolResponse string) error {
	// Truncate the prompt to the max tokens allowed by the model.
	// This is required because the tool may have generated a long output.
	toolResponse = llms.ConstrictPrompt(toolResponse, r.Model)
	// Update the truncated toolResponse
	currentStep.Observation = toolResponse

	// Update our plan tracker with the toolResponse
	r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "in_progress", currentStep.Action.Name, toolResponse)

	// Create a new flow for processing the toolResponse observation
	observationActionJSON, _ := json.MarshalIndent(currentStep, "", "  ")
	observationFlow := &swarm.SimpleFlow{
		Name:     "tool-call",
		Model:    r.Model,
		MaxTurns: 30,
		Steps: []swarm.SimpleFlowStep{
			{
				Name:         "tool-call-step",
				Instructions: nextStepPrompt,
				Inputs: map[string]interface{}{
					"instructions": fmt.Sprintf("User input: %s\n\nCurrent plan with tool execution result:\n%s\n",
						r.Instructions, string(observationActionJSON)),
					"chatHistory": r.ChatHistory,
				},
			},
		},
	}

	// Initialize the workflow for processing the observation
	observationFlow.Initialize()
	if r.Verbose {
		color.Blue("[Step %d: %s] Processing tool observation\n", r.PlanTracker.CurrentStep+1, currentStep.Name)
	}

	// Run the observation processing
	obsCtx, obsCancel := context.WithTimeout(ctx, 5*time.Minute)
	observationResult, observationChatHistory, err := observationFlow.Run(obsCtx, r.Client)
	obsCancel()

	if err != nil {
		if r.Verbose {
			color.Red("Error processing tool observation: %v\n", err)
		}
		// Mark step with the appropriate status based on tool execution
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "failed", currentStep.Action.Name, toolResponse)

		// Try to move to the next step regardless of the error
		r.PlanTracker.MoveToNextStep()
		return nil
	}

	// Update bounded chat history
	r.ChatHistory = limitChatHistory(observationChatHistory, 20)
	if r.Verbose {
		color.Cyan("[Step %d: %s] Observation processing response:\n%s\n\n", r.PlanTracker.CurrentStep+1, currentStep.Name, observationResult)
	}

	// Parse the observation result
	var observationAction ReactAction
	if err = extractReactAction(observationResult, &observationAction); err != nil {
		if r.Verbose {
			color.Red("Unable to parse observation response as JSON: %v\n", err)
		}
		// Try to extract a final answer from the raw response
		potentialAnswer := extractAnswerFromText(observationResult)
		if potentialAnswer != "" {
			r.PlanTracker.FinalAnswer = potentialAnswer
		}

		// Mark step with the determined status and move on
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "completed", currentStep.Action.Name, toolResponse)
		r.PlanTracker.MoveToNextStep()
		return nil
	}

	// Update the step's observation with the thought from observationAction
	observationThought := observationAction.Thought
	if observationThought != "" {
		r.PlanTracker.Steps[r.PlanTracker.CurrentStep].Observation = observationThought
		if r.PlanTracker.CurrentStep < len(observationAction.Steps) {
			observationAction.Steps[r.PlanTracker.CurrentStep].Observation = observationThought
		}
	}

	// Sync steps from observation action with our tracker
	r.PlanTracker.SyncStepsWithReactAction(&observationAction)

	// Check if we have a final answer from observation processing
	if observationAction.FinalAnswer != "" {
		r.PlanTracker.FinalAnswer = observationAction.FinalAnswer
		if r.Verbose {
			color.Cyan("Final answer received from observation processing: %s\n", r.PlanTracker.FinalAnswer)
		}

		// Mark current step as completed
		r.PlanTracker.UpdateStepStatus(r.PlanTracker.CurrentStep, "completed", currentStep.Action.Name, observationThought)

		// Even if we have a final answer, we should still move to the next step
		// if we're not at the last step
		r.PlanTracker.MoveToNextStep()
		return nil
	}

	// Store the current step before any changes for comparison
	originalStepIndex := r.PlanTracker.CurrentStep

	// Get the observation action's current step index
	observationStepIndex := observationAction.CurrentStepIndex

	// First verify the observation step index is valid
	if observationStepIndex < 0 || observationStepIndex >= len(observationAction.Steps) {
		if r.Verbose {
			color.Yellow("Observation action has invalid step index %d. Using our current step %d instead.",
				observationStepIndex, r.PlanTracker.CurrentStep)
		}
		observationStepIndex = r.PlanTracker.CurrentStep
	}

	// Check if the observation indicates we should run a different action for the current step
	if observationStepIndex == originalStepIndex &&
		observationAction.Steps[observationStepIndex].Action.Name != "" {
		// Update the current step with new action info but keep the same index
		if r.Verbose {
			color.Yellow("Updating current step %d with new action: %s",
				originalStepIndex+1, observationAction.Steps[observationStepIndex].Action.Name)
		}

		r.PlanTracker.Steps[originalStepIndex].Observation = observationThought
		r.PlanTracker.Steps[originalStepIndex].Action = observationAction.Steps[observationStepIndex].Action
		r.PlanTracker.Steps[originalStepIndex].Status = "in_progress"
		return nil // Continue with the same step but with a new action
	}

	// If model suggests a later step with an action, move there
	if observationStepIndex > originalStepIndex {
		// Check if that step has an action defined
		if observationStepIndex < len(observationAction.Steps) &&
			observationAction.Steps[observationStepIndex].Action.Name != "" {

			// Mark current step as completed
			r.PlanTracker.UpdateStepStatus(originalStepIndex, "completed", currentStep.Action.Name, observationThought)

			// Make sure we're not stepping beyond our plan's steps
			if observationStepIndex < len(r.PlanTracker.Steps) {
				// Update the target step's action info before moving to it
				r.PlanTracker.Steps[observationStepIndex].Action = observationAction.Steps[observationStepIndex].Action
				r.PlanTracker.Steps[observationStepIndex].Status = "pending"

				// Jump directly to that step
				r.PlanTracker.CurrentStep = observationStepIndex

				if r.Verbose {
					color.Yellow("Jumping forward to step %d to execute action: %s",
						observationStepIndex+1, r.PlanTracker.Steps[observationStepIndex].Action.Name)
				}
				return nil
			}
		}
	}

	// Check all steps for any actions to execute
	for i := 0; i < len(observationAction.Steps); i++ {
		// Skip the current step we just processed
		if i == originalStepIndex {
			continue
		}

		// If we find a step with an action to execute, move to it
		if observationAction.Steps[i].Action.Name != "" &&
			i < len(r.PlanTracker.Steps) &&
			(r.PlanTracker.Steps[i].Status == "pending" ||
				r.PlanTracker.Steps[i].Status == "in_progress") {

			// Mark current step as completed
			r.PlanTracker.UpdateStepStatus(originalStepIndex, "completed", currentStep.Action.Name, observationThought)

			// Update the target step's action and move to it
			r.PlanTracker.Steps[i].Action = observationAction.Steps[i].Action
			r.PlanTracker.CurrentStep = i

			if r.Verbose {
				color.Yellow("Moving to step %d to execute action: %s",
					i+1, r.PlanTracker.Steps[i].Action.Name)
			}
			return nil
		}
	}

	// Default case: mark current step as completed and move to next
	r.PlanTracker.UpdateStepStatus(originalStepIndex, "completed", currentStep.Action.Name, observationThought)
	r.PlanTracker.MoveToNextStep()
	return nil
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
		observation := step.Observation
		if len(observation) > 200 {
			observation = observation[:200] + " <truncated>"
		}
		formattedObs := formatObservationOutputs(observation)
		sb.WriteString(fmt.Sprintf("Observation:\n%s\n\n", formattedObs))
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
