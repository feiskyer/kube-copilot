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
	"github.com/feiskyer/kube-copilot/pkg/utils"
	"github.com/feiskyer/kube-copilot/pkg/workflows"
	"github.com/spf13/cobra"
)

var (
	auditName      string
	auditNamespace string
)

func init() {
	auditCmd.PersistentFlags().StringVarP(&auditName, "name", "n", "", "Pod name")
	auditCmd.PersistentFlags().StringVarP(&auditNamespace, "namespace", "s", "default", "Pod namespace")
	auditCmd.MarkFlagRequired("name")
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit security issues for a Pod",
	Run: func(cmd *cobra.Command, args []string) {
		if auditName == "" && len(args) > 0 {
			auditName = args[0]
		}
		if auditName == "" {
			fmt.Println("Please provide a pod name")
			return
		}

		fmt.Printf("Auditing Pod %s/%s\n", auditNamespace, auditName)
		response, err := workflows.AuditFlow(model, auditNamespace, auditName, verbose)
		if err != nil {
			color.Red(err.Error())
			return
		}

		utils.RenderMarkdown(response)
	},
}
