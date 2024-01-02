package tools

import (
	"context"
	"fmt"
	"os"

	customsearch "google.golang.org/api/customsearch/v1"
	option "google.golang.org/api/option"
)

// GoogleSearch returns the results of a Google search for the given query.
func GoogleSearch(query string) (string, error) {
	svc, err := customsearch.NewService(context.Background(), option.WithAPIKey(os.Getenv("GOOGLE_API_KEY")))
	if err != nil {
		return "", err
	}

	resp, err := svc.Cse.List().Cx(os.Getenv("GOOGLE_CSE_ID")).Q(query).Do()
	if err != nil {
		return "", err
	}

	results := ""
	for _, result := range resp.Items {
		results += fmt.Sprintf("%s: %s\n", result.Title, result.Snippet)
	}
	return results, nil
}
