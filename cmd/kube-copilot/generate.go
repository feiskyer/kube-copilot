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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/feiskyer/kube-copilot/pkg/assistants"
	"github.com/feiskyer/kube-copilot/pkg/kubernetes"
	"github.com/feiskyer/kube-copilot/pkg/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const generateSystemPrompt = `As a skilled technical specialist in Kubernetes and cloud-native technologies, your task is to create Kubernetes YAML manifests following these steps:

1. Review the provided instructions to generate the Kubernetes YAML manifests. Ensure these manifests adhere to current security protocols and best practices. If an instruction lacks a specific image, choose the most commonly used one from reputable sources.
2. Utilize your expertise to scrutinize the YAML manifests. Conduct a thorough step-by-step analysis to identify any issues. Resolve these issues, ensuring the YAML manifests are accurate and secure.
3. After fixing and verifying the manifests, compile them in their raw form. For multiple YAML files, use '---' as a separator.

While refining the YAML manifests, adopt this Chain of Thought:

- Understand the intended use and environment for each manifest as per instructions.
- Assess the security aspects of each component, aligning with standards and best practices.
- Document and justify any discrepancies or issues you find, sequentially.
- Implement solutions that enhance the manifests' performance and security, using best practices and recommended images.
- Ensure the final manifests are syntactically correct, properly formatted, and deployment-ready.

Present **only the final YAML manifests** in raw format, without additional comments or annotations.
`

var generatePrompt string

func init() {
	generateCmd.PersistentFlags().StringVarP(&generatePrompt, "prompt", "p", "", "Prompts to generate Kubernetes manifests")
	generateCmd.MarkFlagRequired("prompt")
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Kubernetes manifests",
	Run: func(cmd *cobra.Command, args []string) {
		if generatePrompt == "" {
			color.Red("Please specify a prompt")
			return
		}

		var err error
		var response string
		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: generateSystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Task: Generate a Kubernetes manifest for %s", generatePrompt),
			},
		}

		response, _, err = assistants.Assistant(model, messages, maxTokens, countTokens, verbose, maxIterations)
		if err != nil {
			color.Red(err.Error())
			return
		}

		// Extract the yaml from the response
		yaml := response
		if strings.Contains(response, "```") {
			yaml = utils.ExtractYaml(response)
		}
		fmt.Printf("\nGenerated manifests:\n\n")
		color.New(color.FgGreen).Printf("%s\n\n", yaml)

		// apply the yaml to kubernetes cluster
		color.New(color.FgRed).Printf("Do you approve to apply the generated manifests to cluster? (y/n)")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			approve := scanner.Text()
			if strings.ToLower(approve) != "y" && strings.ToLower(approve) != "yes" {
				break
			}

			if err := kubernetes.ApplyYaml(yaml); err != nil {
				color.Red(err.Error())
				return
			}

			color.New(color.FgGreen).Printf("Applied the generated manifests to cluster successfully!")
			break
		}
	},
}
