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
