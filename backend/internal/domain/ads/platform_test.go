package ads

import (
	"errors"
	"testing"
)

func TestParsePlatform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    Platform
		expectedErr error
	}{
		{name: "meta", input: "META", expected: PlatformMeta},
		{name: "tiktok", input: " tiktok ", expected: PlatformTikTok},
		{name: "google", input: "google", expected: PlatformGoogle},
		{name: "unknown", input: "linkedin", expected: PlatformUnknown, expectedErr: ErrUnsupportedPlatform},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			platform, err := ParsePlatform(test.input)
			if !errors.Is(err, test.expectedErr) {
				t.Fatalf("ParsePlatform() error = %v, expected %v", err, test.expectedErr)
			}
			if platform != test.expected {
				t.Errorf("ParsePlatform() = %q, expected %q", platform, test.expected)
			}
		})
	}
}
