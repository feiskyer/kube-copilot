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
	"github.com/feiskyer/kube-copilot/pkg/kubernetes"
	"github.com/feiskyer/kube-copilot/pkg/utils"
	"github.com/feiskyer/kube-copilot/pkg/workflows"
	"github.com/spf13/cobra"
)

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

		response, err := workflows.AnalysisFlow(model, manifests, verbose)
		if err != nil {
			color.Red(err.Error())
			return
		}

		utils.RenderMarkdown(response)
	},
}
