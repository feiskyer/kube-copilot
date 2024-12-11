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
package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/feiskyer/kube-copilot/pkg/assistants"
	"github.com/feiskyer/kube-copilot/pkg/tools"
	kubetools "github.com/feiskyer/kube-copilot/pkg/tools"
	"github.com/feiskyer/kube-copilot/pkg/utils"
	"github.com/feiskyer/kube-copilot/pkg/workflows"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const executeSystemPrompt = `As a technical expert in Kubernetes and cloud-native networking, your task follows a specific Chain of Thought methodology to ensure thoroughness and accuracy while adhering to the constraints provided.
Available Tools:
- kubectl: Useful for executing kubectl commands. Remember to use '--sort-by=memory' or '--sort-by=cpu' when running 'kubectl top' command.  Input: a kubectl command. Output: the result of the command.
- python: This is a Python interpreter. Use it for executing Python code with the Kubernetes Python SDK client. Ensure the results are output using "print(...)". The input is a Python script, and the output will be the stdout and stderr of this script.
- trivy: Useful for executing trivy image command to scan images for vulnerabilities. Input: an image for security scanning. Output: the vulnerabilities found in the image.

The steps you take are as follows:

1. Problem Identification: Begin by clearly defining the problem you're addressing. When diagnostics or troubleshooting is needed, specify the symptoms or issues observed that prompted the analysis. This helps to narrow down the potential causes and guides the subsequent steps.
2. Diagnostic Commands: Utilize 'python' tool to gather information about the state of the Kubernetes resources, network policies, and other related configurations. Detail why each command is chosen and what information it is expected to yield. In cases where 'trivy' is applicable, explain how it will be used to analyze container images for vulnerabilities.
3. Interpretation of Outputs: Analyze the outputs from the executed commands. Describe what the results indicate about the health and configuration of the system and network. This is crucial for identifying any discrepancies that may be contributing to the issue at hand.
4. Troubleshooting Strategy: Based on the interpreted outputs, develop a step-by-step strategy for troubleshooting. Justify each step within the strategy, explaining how it relates to the findings from the diagnostic outputs.
5. Actionable Solutions: Propose solutions that can be carried out using 'kubectl' commands, where possible. If the solution involves a sequence of actions, explain the order and the expected outcome of each. For issues identified by 'trivy', provide recommendations for remediation based on best practices.
6. Contingency for Unavailable Tools: In the event that the necessary tools or commands are unavailable, provide an alternative set of instructions that comply with the guidelines, explaining how these can help progress the troubleshooting process.

Throughout this process, ensure that each response is concise and strictly adheres to the guidelines provided, with a clear justification for each step taken. The ultimate goal is to identify the root cause of issues within the domains of Kubernetes and cloud-native networking and to provide clear, actionable solutions, while staying within the operational constraints of 'kubectl' or 'trivy image' for diagnostics and troubleshooting and avoiding any installation operations.

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

var instructions string

func init() {
	tools.CopilotTools["trivy"] = kubetools.Trivy

	executeCmd.PersistentFlags().StringVarP(&instructions, "instructions", "", "", "instructions to execute")
	executeCmd.MarkFlagRequired("instructions")
}

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute operations based on prompt instructions",
	Run: func(cmd *cobra.Command, args []string) {
		if instructions == "" && len(args) > 0 {
			instructions = strings.Join(args, " ")
		}
		if instructions == "" {
			fmt.Println("Please provide the instructions")
			return
		}

		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: executeSystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Here are the instructions: %s", instructions),
			},
		}
		response, _, err := assistants.Assistant(model, messages, maxTokens, countTokens, verbose, maxIterations)
		if err != nil {
			color.Red(err.Error())
			return
		}

		instructions := fmt.Sprintf("Extract the execuation results for user instructions and reformat in a concise Markdown response: %s", response)
		result, err := workflows.AssistantFlow(model, instructions, verbose)
		if err != nil {
			color.Red(err.Error())
			fmt.Println(response)
			return
		}

		utils.RenderMarkdown(result)
	},
}
