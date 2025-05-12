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
	"github.com/feiskyer/kube-copilot/pkg/tools"
	"github.com/feiskyer/kube-copilot/pkg/workflows"
	"github.com/spf13/cobra"
)

var (
	diagnoseName           string
	diagnoseNamespace      string
	dignoseMCPFile         string
	diagnoseDisableKubectl bool
)

func init() {
	diagnoseCmd.PersistentFlags().StringVarP(&diagnoseName, "name", "n", "", "Pod name")
	diagnoseCmd.PersistentFlags().StringVarP(&diagnoseNamespace, "namespace", "s", "default", "Pod namespace")
	diagnoseCmd.PersistentFlags().StringVarP(&dignoseMCPFile, "mcp-config", "p", "", "MCP configuration file")
	diagnoseCmd.PersistentFlags().BoolVarP(&diagnoseDisableKubectl, "disable-kubectl", "d", false, "Disable kubectl tool (useful when using MCP server)")

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

		clients, err := tools.InitTools(dignoseMCPFile, diagnoseDisableKubectl, verbose)
		if err != nil {
			color.Red("Failed to initialize tools: %v", err)
			return
		}

		defer func() {
			for _, client := range clients {
				go client.Close()
			}
		}()

		prompt := fmt.Sprintf("Diagnose the issues for Pod %s in namespace %s", diagnoseName, diagnoseNamespace)
		flow, err := workflows.NewReActFlow(model, prompt, tools.GetToolPrompt(), verbose, maxIterations)
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
