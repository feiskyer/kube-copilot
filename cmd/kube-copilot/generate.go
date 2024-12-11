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
	"github.com/feiskyer/kube-copilot/pkg/kubernetes"
	"github.com/feiskyer/kube-copilot/pkg/utils"
	"github.com/feiskyer/kube-copilot/pkg/workflows"
	"github.com/spf13/cobra"
)

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

		response, err := workflows.GeneratorFlow(model, generatePrompt, verbose)
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
