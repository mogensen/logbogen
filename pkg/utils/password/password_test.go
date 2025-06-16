package password

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := Generate(tt.password)
			if hash == tt.password {
				t.Errorf("Generate() = %v, want different from input password", hash)
			}
			if len(hash) == 0 {
				t.Error("Generate() returned empty hash")
			}
		})
	}
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := Generate("password123")
			err := Verify(hash, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
