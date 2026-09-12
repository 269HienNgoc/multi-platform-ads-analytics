package ads

import "testing"

func TestEntityKindParent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		kind           EntityKind
		expectedParent EntityKind
		expectedOK     bool
	}{
		{name: "account", kind: EntityKindAccount, expectedParent: EntityKindUnknown, expectedOK: false},
		{name: "campaign", kind: EntityKindCampaign, expectedParent: EntityKindAccount, expectedOK: true},
		{name: "ad group", kind: EntityKindAdGroup, expectedParent: EntityKindCampaign, expectedOK: true},
		{name: "ad", kind: EntityKindAd, expectedParent: EntityKindAdGroup, expectedOK: true},
		{name: "creative", kind: EntityKindCreative, expectedParent: EntityKindAd, expectedOK: true},
		{name: "unsupported", kind: EntityKind("unsupported"), expectedParent: EntityKindUnknown, expectedOK: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parent, ok := test.kind.Parent()
			if parent != test.expectedParent || ok != test.expectedOK {
				t.Errorf("Parent() = (%q, %v), expected (%q, %v)", parent, ok, test.expectedParent, test.expectedOK)
			}
		})
	}
}
