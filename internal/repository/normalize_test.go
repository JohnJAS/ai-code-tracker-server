package repository

import "testing"

func TestNormalizeOrigin(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "SSH", input: "git@GitHub.com:acme/demo.git", want: "github.com/acme/demo"},
		{name: "HTTPS", input: "https://github.com/acme/demo.git/", want: "github.com/acme/demo"},
		{name: "HTTPS credentials", input: "https://token@example.com/acme/demo.git", want: "example.com/acme/demo"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeOrigin(test.input)
			if err != nil {
				t.Fatalf("NormalizeOrigin() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("NormalizeOrigin() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeOriginRejectsInvalidRemote(t *testing.T) {
	if _, err := NormalizeOrigin("not a remote"); err == nil {
		t.Fatal("NormalizeOrigin() error = nil, want error")
	}
}
