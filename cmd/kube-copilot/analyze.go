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
	"github.com/feiskyer/kube-copilot/pkg/kubernetes"
	"github.com/feiskyer/kube-copilot/pkg/llms"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const analysisSystemPrompt = `Transform technical Kubernetes analysis into accessible explanations for non-technical users using relatable analogies and a "detective solving a mystery" approach. For each identified issue, document the analysis and solution in everyday language, employing simple analogies to clarify technical points.

# Steps

1. **Identify Clues**: Treat each piece of YAML configuration data like a clue in a mystery. Explain how it helps to understand the issue, similar to a detective piecing together a case.
2. **Analysis with Analogies**: Translate your technical findings into relatable scenarios. Use everyday analogies to explain concepts, avoiding complex jargon. This makes episodes like 'pod failures' or 'service disruptions' simple to grasp.
3. **Solution as a DIY Guide**: Offer a step-by-step solution akin to guiding someone through a household fix-up. Instructions should be straightforward, logical, and accessible.
4. **Document Findings**:
   - Separate analysis and solution clearly for each issue, detailing them in non-technical language.

# Output Format

Provide the output in structured markdown, using clear and concise language.

# Examples

## 1. <title of the issue or potential problem>

- **Findings**: The YAML configuration doesn't specify the memory limit for the pod.
- **How to resolve**: Set memory limit in Pod spec.

## 2. HIGH Severity: CVE-2024-10963

- **Findings**: The Pod is running with CVE pam: Improper Hostname Interpretation in pam_access Leads to Access Control Bypass.
- **How to resolve**: Update package libpam-modules to fixed version (>=1.5.3) in the image. (leave the version number to empty if you don't know it)

# Notes

- Keep your language concise and simple.
- Ensure key points are included, e.g. CVE number, error code, versions.
- Relatable analogies should help in visualizing the problem and solution.
- Ensure explanations are self-contained, enough for newcomers without previous technical exposure to understand.
`

var analysisName string
var analysisNamespace string
var analysisResource string

func init() {
	analyzeCmd.PersistentFlags().StringVarP(&analysisName, "name", "", "", "Resource name")
	analyzeCmd.PersistentFlags().StringVarP(&analysisNamespace, "namespace", "n", "default", "Resource namespace")
	analyzeCmd.PersistentFlags().StringVarP(&analysisResource, "resource", "r", "pod", "Resource type")
	analyzeCmd.MarkFlagRequired("name")
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze issues for a given resource",
	Run: func(cmd *cobra.Command, args []string) {
		if analysisName == "" && len(args) > 0 {
			analysisName = args[0]
		}
		if analysisName == "" {
			fmt.Println("Please provide a resource name")
			return
		}

		fmt.Printf("Analysing %s %s/%s\n", analysisResource, analysisNamespace, analysisName)

		manifests, err := kubernetes.GetYaml(analysisResource, analysisName, analysisNamespace)
		if err != nil {
			color.Red(err.Error())
			return
		}

		manifests = llms.ConstrictPrompt(manifests, model, maxTokens)
		if verbose {
			color.Cyan("Got manifests for %s/%s:\n%s\n\n", analysisNamespace, analysisName, manifests)
		}

		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: analysisSystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("**Task**: Analyze and explain issues for the following kubernetes manifests:\n%s", manifests),
			},
		}
		response, _, err := assistants.Assistant(model, messages, maxTokens, countTokens, verbose, maxIterations)
		if err != nil {
			color.Red(err.Error())
			return
		}
		fmt.Println(response)
	},
}
