package client

import (
	"testing"
)

func TestPropertyResolver_Resolve(t *testing.T) {
	tests := []struct {
		name    string
		aliases map[string]string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "numeric ID is returned as-is",
			aliases: map[string]string{"my-app": "123456789"},
			input:   "123456789",
			want:    "123456789",
			wantErr: false,
		},
		{
			name:    "registered alias resolves to ID",
			aliases: map[string]string{"my-app": "123456789"},
			input:   "my-app",
			want:    "123456789",
			wantErr: false,
		},
		{
			name:    "unregistered alias returns error",
			aliases: map[string]string{"my-app": "123456789"},
			input:   "unknown-app",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string returns error",
			aliases: map[string]string{"my-app": "123456789"},
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "nil aliases with numeric ID still works",
			aliases: nil,
			input:   "123456789",
			want:    "123456789",
			wantErr: false,
		},
		{
			name:    "nil aliases with string input returns error",
			aliases: nil,
			input:   "my-app",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewPropertyResolver(tt.aliases)
			got, err := r.Resolve(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}
