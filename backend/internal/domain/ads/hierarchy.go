package ads

// EntityKind identifies a level in the canonical advertising hierarchy.
type EntityKind string

const (
	// EntityKindUnknown represents an unset hierarchy level.
	EntityKindUnknown EntityKind = ""
	// EntityKindAccount represents an advertising account.
	EntityKindAccount EntityKind = "account"
	// EntityKindCampaign represents a campaign.
	EntityKindCampaign EntityKind = "campaign"
	// EntityKindAdGroup represents a Meta ad set, Google ad group, or equivalent level.
	EntityKindAdGroup EntityKind = "ad_group"
	// EntityKindAd represents an advertisement.
	EntityKindAd EntityKind = "ad"
	// EntityKindCreative represents creative content attached to an advertisement.
	EntityKindCreative EntityKind = "creative"
)

// Parent returns the canonical parent kind and whether a parent is required.
func (k EntityKind) Parent() (EntityKind, bool) {
	switch k {
	case EntityKindCampaign:
		return EntityKindAccount, true
	case EntityKindAdGroup:
		return EntityKindCampaign, true
	case EntityKindAd:
		return EntityKindAdGroup, true
	case EntityKindCreative:
		return EntityKindAd, true
	case EntityKindAccount, EntityKindUnknown:
		return EntityKindUnknown, false
	default:
		return EntityKindUnknown, false
	}
}
