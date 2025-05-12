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

var (
	executeInstructions   string
	executeMCPFile        string
	executeDisableKubectl bool
)

func init() {
	executeCmd.PersistentFlags().StringVarP(&executeInstructions, "instructions", "i", "", "instructions to execute")
	executeCmd.PersistentFlags().StringVarP(&executeMCPFile, "mcp-config", "p", "", "MCP configuration file")
	executeCmd.PersistentFlags().BoolVarP(&executeDisableKubectl, "disable-kubectl", "d", false, "Disable kubectl tool (useful when using MCP server)")

	executeCmd.MarkFlagRequired("instructions")
}

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute operations based on prompt instructions",
	Run: func(cmd *cobra.Command, args []string) {
		if executeInstructions == "" && len(args) > 0 {
			executeInstructions = strings.Join(args, " ")
		}
		if executeInstructions == "" {
			fmt.Println("Please provide the instructions")
			return
		}

		clients, err := tools.InitTools(executeMCPFile, executeDisableKubectl, verbose)
		if err != nil {
			color.Red("Failed to initialize tools: %v", err)
			return
		}

		defer func() {
			for _, client := range clients {
				go client.Close()
			}
		}()

		flow, err := workflows.NewReActFlow(model, executeInstructions, tools.GetToolPrompt(), verbose, maxIterations)
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
