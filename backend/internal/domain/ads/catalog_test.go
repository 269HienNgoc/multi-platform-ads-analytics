package ads

import "testing"

func TestStatus_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{name: "active", status: StatusActive, expected: true},
		{name: "paused", status: StatusPaused, expected: true},
		{name: "archived", status: StatusArchived, expected: true},
		{name: "unknown", status: StatusUnknown, expected: false},
		{name: "unsupported", status: Status("deleted"), expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if actual := test.status.IsValid(); actual != test.expected {
				t.Errorf("IsValid() = %t, expected %t", actual, test.expected)
			}
		})
	}
}
