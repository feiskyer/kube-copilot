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
