package utils

import (
	"fmt"

	"github.com/charmbracelet/glamour"
)

// RenderMarkdown renders markdown to the terminal.
func RenderMarkdown(md string) error {
	styler, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
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
