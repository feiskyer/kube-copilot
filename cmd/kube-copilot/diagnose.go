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

	"github.com/fatih/color"
	"github.com/feiskyer/kube-copilot/pkg/assistants"
	"github.com/feiskyer/kube-copilot/pkg/utils"
	"github.com/feiskyer/kube-copilot/pkg/workflows"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const diagnoseSystemPrompt = `You are a seasoned expert in Kubernetes and cloud-native networking. Utilize a Chain of Thought (CoT) process to diagnose and resolve issues. Your explanations should be in simple terms for non-technical users to understand.

Available Tools:
- kubectl: Useful for executing kubectl commands. Input: a kubectl command. Output: the result of the command.
- python: This is a Python interpreter. Use it for executing Python code with the Kubernetes Python SDK client. Ensure the results are output using "print(...)". The input is a Python script, and the output will be the stdout and stderr of this script.

Here is your process:

1. Information Gathering:
   a. Using the Kubernetes Python SDK with "python" tool, detail how you retrieve data like pod status, logs, and events. Explain the significance of each data type in understanding the cluster's state in layman's terms.
   b. Outline your plan for executing SDK calls. Describe what each call does in simple language, making it understandable for non-technical users.

2. Issue Analysis:
   a. Systematically analyze the gathered information. Describe how you identify inconsistencies or signs of issues in the cluster. Explain your thought process in determining the expected versus the actual data.
   b. Translate your findings into a narrative easy for non-technical users to follow, using analogies to explain complex concepts.

3. Configuration Verification:
   a. Explain how to verify the configurations of Pod, Service, Ingress, and NetworkPolicy resources. Simplify the explanation of each resource's role and its importance for the cluster's health.
   b. Discuss common misconfigurations and their impact on the cluster's operations, keeping explanations straightforward and free of technical jargon.

4. Network Connectivity Analysis:
   a. Describe your approach to analysing network connectivity within the cluster and to external services. Explain the importance of the chosen tools or methods.
   b. Use simple analogies to explain how network issues might manifest, making the concept easy to visualize for non-technical users.

Present your findings in this accessible format:

1. Issue: <Issue 1>
   Analysis: Describe the symptoms and your process of identifying Issue 1.
   Solution: Detail the steps to resolve Issue 1, explaining their effectiveness in simple terms.

2. Issue: <Issue 2>
   Analysis: Explain the clues leading to Issue 2 in understandable language.
   Solution: Provide a non-technical explanation for resolving Issue 2, clarifying the reasoning behind each step.

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

var diagnoseName string
var diagnoseNamespace string

func init() {
	diagnoseCmd.PersistentFlags().StringVarP(&diagnoseName, "name", "", "", "Pod name")
	diagnoseCmd.PersistentFlags().StringVarP(&diagnoseNamespace, "namespace", "n", "default", "Pod namespace")
	diagnoseCmd.MarkFlagRequired("name")
}

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose problems for a Pod",
	Run: func(cmd *cobra.Command, args []string) {
		if diagnoseName == "" && len(args) > 0 {
			diagnoseName = args[0]
		}
		if diagnoseName == "" {
			fmt.Println("Please provide a pod name")
			return
		}

		fmt.Printf("Diagnosing Pod %s/%s\n", diagnoseNamespace, diagnoseName)
		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: diagnoseSystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Your goal is to ensure that both the issues and their solutions are communicated effectively and understandably. As you diagnose issues for Pod %s in namespace %s, remember to avoid using any delete or edit commands.", diagnoseName, diagnoseNamespace),
			},
		}
		response, _, err := assistants.Assistant(model, messages, maxTokens, countTokens, verbose, maxIterations)
		if err != nil {
			color.Red(err.Error())
			return
		}

		instructions := fmt.Sprintf("Extract the final diagnose results and reformat in a concise Markdown response: %s", response)
		result, err := workflows.AssistantFlow(model, instructions, verbose)
		if err != nil {
			color.Red(err.Error())
			fmt.Println(response)
			return
		}

		utils.RenderMarkdown(result)
	},
}
