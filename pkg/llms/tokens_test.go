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
	"reflect"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestGetTokenLimits(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected int
	}{
		{"gpt-3.5-turbo", "gpt-3.5-turbo", 4096},
		{"gpt-4", "gpt-4", 8192},
		{"gpt-4-turbo", "gpt-4-turbo", 128000},
		{"case insensitive", "GPT-4", 8192},
		{"unknown model", "unknown-model", 8192},
		{"claude model", "o1-mini", 128000},
		{"claude o1", "o1", 200000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTokenLimits(tt.model)
			if got != tt.expected {
				t.Errorf("GetTokenLimits(%s) = %d, want %d", tt.model, got, tt.expected)
			}
		})
	}
}

func TestNumTokensFromMessages(t *testing.T) {
	tests := []struct {
		name      string
		messages  []openai.ChatCompletionMessage
		model     string
		minTokens int // Use min/max range since exact token counts can be version-dependent
		maxTokens int
	}{
		{
			name: "empty message",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: ""},
			},
			model:     "gpt-4",
			minTokens: 3,
			maxTokens: 8,
		},
		{
			name: "simple message",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "Hello, world!"},
			},
			model:     "gpt-4",
			minTokens: 8,
			maxTokens: 12,
		},
		{
			name: "multiple messages",
			messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Tell me about AI."},
				{Role: "assistant", Content: "AI stands for artificial intelligence."},
			},
			model:     "gpt-4",
			minTokens: 25,
			maxTokens: 35,
		},
		{
			name: "message with name",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "Hello", Name: "John"},
			},
			model:     "gpt-4",
			minTokens: 8,
			maxTokens: 12,
		},
		{
			name: "o1-mini model",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "Hello, world!"},
			},
			model:     "o1-mini",
			minTokens: 8,
			maxTokens: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NumTokensFromMessages(tt.messages, tt.model)
			if got < tt.minTokens || got > tt.maxTokens {
				t.Errorf("NumTokensFromMessages() = %d, want between %d and %d",
					got, tt.minTokens, tt.maxTokens)
			}
		})
	}
}

func TestConstrictMessages(t *testing.T) {
	systemMsg := openai.ChatCompletionMessage{Role: "system", Content: "You are a helpful assistant."}

	// Create a long message that will exceed token limits
	longContent := strings.Repeat("This is a very long message that will exceed token limits. ", 1000)
	longMsg := openai.ChatCompletionMessage{Role: "user", Content: longContent}

	shortMsg1 := openai.ChatCompletionMessage{Role: "user", Content: "Short message 1"}
	shortMsg2 := openai.ChatCompletionMessage{Role: "assistant", Content: "Short reply 1"}
	shortMsg3 := openai.ChatCompletionMessage{Role: "user", Content: "Short message 2"}

	tests := []struct {
		name         string
		messages     []openai.ChatCompletionMessage
		model        string
		expectedLen  int
		checkContent bool // whether to check exact content or just length
	}{
		{
			name:         "under limit",
			messages:     []openai.ChatCompletionMessage{systemMsg, shortMsg1, shortMsg2},
			model:        "gpt-4",
			expectedLen:  3,
			checkContent: true,
		},
		{
			name:         "over limit with multiple messages",
			messages:     []openai.ChatCompletionMessage{systemMsg, longMsg, shortMsg1, shortMsg2, shortMsg3},
			model:        "gpt-3.5-turbo",
			expectedLen:  2, // Should keep system and latest message
			checkContent: false,
		},
		{
			name:         "empty messages",
			messages:     []openai.ChatCompletionMessage{},
			model:        "gpt-4",
			expectedLen:  0,
			checkContent: true,
		},
		{
			name:         "single message over limit",
			messages:     []openai.ChatCompletionMessage{longMsg},
			model:        "gpt-3.5-turbo",
			expectedLen:  1, // Still keeps the message even if it's over limit
			checkContent: true,
		},
		{
			name:         "system and long message over limit",
			messages:     []openai.ChatCompletionMessage{systemMsg, longMsg},
			model:        "gpt-3.5-turbo",
			expectedLen:  1, // Should keep only system message
			checkContent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConstrictMessages(tt.messages, tt.model)

			if len(result) != tt.expectedLen {
				t.Errorf("ConstrictMessages() returned %d messages, want %d",
					len(result), tt.expectedLen)
			}

			if tt.checkContent && tt.expectedLen > 0 && !reflect.DeepEqual(result, tt.messages) {
				t.Errorf("ConstrictMessages() modified messages when it shouldn't have")
			}

			// For over limit cases with system message, verify system is preserved
			if tt.expectedLen > 0 && len(tt.messages) > 0 &&
				tt.messages[0].Role == "system" && !tt.checkContent {
				if result[0].Role != "system" {
					t.Errorf("ConstrictMessages() didn't preserve system message")
				}
			}
		})
	}
}

func TestConstrictPrompt(t *testing.T) {
	shortPrompt := "This is a short prompt."

	// Create prompts of different lengths
	mediumPrompt := strings.Repeat("This is a medium length prompt that should still fit. ", 50)
	longPrompt := strings.Repeat("This is a very long prompt that will need to be trimmed. ", 2000)

	tests := []struct {
		name          string
		prompt        string
		model         string
		shouldBeSame  bool
		shouldBeEmpty bool
	}{
		{
			name:          "short prompt",
			prompt:        shortPrompt,
			model:         "gpt-4",
			shouldBeSame:  true,
			shouldBeEmpty: false,
		},
		{
			name:          "medium prompt",
			prompt:        mediumPrompt,
			model:         "gpt-4",
			shouldBeSame:  true,
			shouldBeEmpty: false,
		},
		{
			name:          "long prompt",
			prompt:        longPrompt,
			model:         "gpt-3.5-turbo",
			shouldBeSame:  false,
			shouldBeEmpty: false,
		},
		{
			name:          "empty prompt",
			prompt:        "",
			model:         "gpt-4",
			shouldBeSame:  true,
			shouldBeEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConstrictPrompt(tt.prompt, tt.model)

			if tt.shouldBeSame && result != tt.prompt {
				t.Errorf("ConstrictPrompt() modified prompt when it shouldn't have")
			}

			if !tt.shouldBeSame && result == tt.prompt && tt.prompt != "" {
				t.Errorf("ConstrictPrompt() didn't modify prompt when it should have")
			}

			if tt.shouldBeEmpty && result != "" {
				t.Errorf("ConstrictPrompt() returned non-empty result for empty prompt")
			}

			// Test that the modified prompt is actually shorter
			if !tt.shouldBeSame && !tt.shouldBeEmpty {
				if len(result) >= len(tt.prompt) {
					t.Errorf("ConstrictPrompt() didn't make prompt shorter")
				}
			}

			// Verify the result fits within token limit
			numTokens := NumTokensFromMessages([]openai.ChatCompletionMessage{{Content: result}}, tt.model)
			tokenLimit := GetTokenLimits(tt.model)
			if numTokens >= tokenLimit {
				t.Errorf("ConstrictPrompt() returned result with %d tokens, exceeding limit of %d",
					numTokens, tokenLimit)
			}
		})
	}
}
