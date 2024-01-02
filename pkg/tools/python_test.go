package tools

import (
	"strings"
	"testing"
)

func TestPythonREPL(t *testing.T) {
	type args struct {
		script string
	}
	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{
			name:    "normal test",
			args:    "print('hello world')",
			want:    "hello world",
			wantErr: false,
		},
		{
			name:    "error test",
			args:    "print('hello world'",
			want:    "SyntaxError: '(' was never closed",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PythonREPL(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("PythonREPL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want && !strings.Contains(got, tt.want) {
				t.Errorf("PythonREPL() = %v, want %v", got, tt.want)
			}
		})
	}
}
