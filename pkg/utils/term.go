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
package utils

import (
	"fmt"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

// RenderMarkdown renders markdown to the terminal.
func RenderMarkdown(md string) error {
	width, _, _ := term.GetSize(0)
	styler, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		fmt.Println(md)
		return err
	}

	out, err := styler.Render(md)
	if err != nil {
		fmt.Println(md)
		return err
	}

	fmt.Println(out)
	return nil
}
