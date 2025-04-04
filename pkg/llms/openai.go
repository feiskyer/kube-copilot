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
package llms

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"time"

	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	*openai.Client

	Retries int
	Backoff time.Duration
}

// NewOpenAIClient returns an OpenAI client.
func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		config := openai.DefaultConfig(apiKey)
		baseURL := os.Getenv("OPENAI_API_BASE")
		if baseURL != "" {
			config.BaseURL = baseURL
		}

		return &OpenAIClient{
			Retries: 5,
			Backoff: time.Second,
			Client:  openai.NewClientWithConfig(config),
		}, nil
	}

	azureAPIKey := os.Getenv("AZURE_OPENAI_API_KEY")
	azureAPIBase := os.Getenv("AZURE_OPENAI_API_BASE")
	azureAPIVersion := os.Getenv("AZURE_OPENAI_API_VERSION")
	if azureAPIVersion == "" {
		azureAPIVersion = "2025-03-01-preview"
	}
	if azureAPIKey != "" && azureAPIBase != "" {
		config := openai.DefaultConfig(azureAPIKey)
		config.BaseURL = azureAPIBase
		config.APIVersion = azureAPIVersion
		config.APIType = openai.APITypeAzure
		config.AzureModelMapperFunc = func(model string) string {
			return regexp.MustCompile(`[.:]`).ReplaceAllString(model, "")
		}

		return &OpenAIClient{
			Retries: 5,
			Backoff: time.Second,
			Client:  openai.NewClientWithConfig(config),
		}, nil
	}

	return nil, fmt.Errorf("OPENAI_API_KEY or AZURE_OPENAI_API_KEY is not set")
}

func (c *OpenAIClient) Chat(model string, maxTokens int, prompts []openai.ChatCompletionMessage) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		Temperature: math.SmallestNonzeroFloat32,
		Messages:    prompts,
	}
	if model == "o1-mini" || model == "o3-mini" || model == "o1" || model == "o3" {
		req = openai.ChatCompletionRequest{
			Model:               model,
			MaxCompletionTokens: maxTokens,
			Messages:            prompts,
		}
	}

	backoff := c.Backoff
	for try := 0; try < c.Retries; try++ {
		resp, err := c.Client.CreateChatCompletion(context.Background(), req)
		if err == nil {
			return string(resp.Choices[0].Message.Content), nil
		}

		e := &openai.APIError{}

		if errors.As(err, &e) {
			switch e.HTTPStatusCode {
			case 401:
				return "", err
			case 429, 500:
				time.Sleep(backoff)
				backoff *= 2
				continue
			default:
				return "", err
			}
		}

		return "", err
	}

	return "", fmt.Errorf("OpenAI request throttled after retrying %d times", c.Retries)
}
