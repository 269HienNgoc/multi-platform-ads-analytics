// Package ads contains provider-neutral advertising domain concepts.
package ads

import (
	"errors"
	"strings"
)

// Platform identifies an advertising provider without leaking provider SDK types.
type Platform string

const (
	// PlatformUnknown represents an unset or unsupported provider.
	PlatformUnknown Platform = ""
	// PlatformMeta represents Meta Ads.
	PlatformMeta Platform = "meta"
	// PlatformTikTok represents TikTok Ads.
	PlatformTikTok Platform = "tiktok"
	// PlatformGoogle represents Google Ads.
	PlatformGoogle Platform = "google"
)

// ErrUnsupportedPlatform is returned for provider names outside the canonical set.
var ErrUnsupportedPlatform = errors.New("ads: unsupported platform")

// ParsePlatform converts an external provider name into a canonical platform.
func ParsePlatform(value string) (Platform, error) {
	platform := Platform(strings.ToLower(strings.TrimSpace(value)))
	switch platform {
	case PlatformMeta, PlatformTikTok, PlatformGoogle:
		return platform, nil
	case PlatformUnknown:
		return PlatformUnknown, ErrUnsupportedPlatform
	default:
		return PlatformUnknown, ErrUnsupportedPlatform
	}
}

// IsValid reports whether the platform belongs to the canonical set.
func (p Platform) IsValid() bool {
	switch p {
	case PlatformMeta, PlatformTikTok, PlatformGoogle:
		return true
	case PlatformUnknown:
		return false
	default:
		return false
	}
}
