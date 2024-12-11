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
