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
	"testing"
)

func TestGetTokenLimits(t *testing.T) {
	type args struct {
		model string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "gpt-3.5-turbo-0613",
			args: args{
				model: "gpt-3.5-turbo-0613",
			},
			want: 4096,
		},
		{
			name: "gpt-4",
			args: args{
				model: "gpt-4",
			},
			want: 8192,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTokenLimits(tt.args.model); got != tt.want {
				t.Errorf("GetTokenLimits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstrictPrompt(t *testing.T) {
	type args struct {
		prompt      string
		model       string
		tokenLimits int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "gpt-3.5-turbo-0613",
			args: args{
				prompt:      "This is a test prompt.",
				model:       "gpt-3.5-turbo-0613",
				tokenLimits: 512,
			},
			want: "This is a test prompt.",
		},
		{
			name: "gpt-3.5-turbo-0613",
			args: args{
				prompt:      "This is a test prompt.",
				model:       "gpt-3.5-turbo-0613",
				tokenLimits: 1,
			},
			want: "",
		},
		{
			name: "gpt-3.5-turbo-0613",
			args: args{
				prompt:      "This is a test prompt.\nhere is another.",
				model:       "gpt-3.5-turbo-0613",
				tokenLimits: 15,
			},
			want: "here is another.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConstrictPrompt(tt.args.prompt, tt.args.model, tt.args.tokenLimits); got != tt.want {
				t.Errorf("ConstrictPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
