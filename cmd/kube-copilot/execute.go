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
	"github.com/feiskyer/kube-copilot/pkg/tools"
	"github.com/feiskyer/kube-copilot/pkg/workflows"
	"github.com/spf13/cobra"
)

var instructions string

func init() {
	tools.CopilotTools["trivy"] = tools.Trivy

	executeCmd.PersistentFlags().StringVarP(&instructions, "instructions", "i", "", "instructions to execute")
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

		flow, err := workflows.NewReActFlow(model, instructions, verbose, maxIterations)
		if err != nil {
			color.Red(err.Error())
			return
		}

		response, err := flow.Run()
		if err != nil {
			color.Red(err.Error())
			return
		}
		fmt.Println(response)
	},
}
