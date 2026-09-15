// Package automation contains provider-neutral campaign workflow concepts.
package automation

import "time"

// WorkflowState is a durable state in the campaign automation lifecycle.
type WorkflowState string

const (
	// StateAccountConnected means provider credentials and the account are available.
	StateAccountConnected WorkflowState = "ACCOUNT_CONNECTED"
	// StateAssetsSynced means required provider assets have been discovered.
	StateAssetsSynced WorkflowState = "ASSETS_SYNCED"
	// StatePreflightPassed means deterministic readiness checks succeeded.
	StatePreflightPassed WorkflowState = "PREFLIGHT_PASSED"
	// StateSeedPending means the seed campaign is queued for creation.
	StateSeedPending WorkflowState = "SEED_PENDING"
	// StateSeedCreating means seed campaign creation is in progress.
	StateSeedCreating WorkflowState = "SEED_CREATING"
	// StateSeedRunning means the seed campaign is collecting engagement.
	StateSeedRunning WorkflowState = "SEED_RUNNING"
	// StateSeedCompleted means the seed spend threshold was reached.
	StateSeedCompleted WorkflowState = "SEED_COMPLETED"
	// StateMainPending means the conversion campaign is queued.
	StateMainPending WorkflowState = "MAIN_PENDING"
	// StateMainValidating means the main campaign is being checked.
	StateMainValidating WorkflowState = "MAIN_VALIDATING"
	// StateMainCreating means main campaign creation is in progress.
	StateMainCreating WorkflowState = "MAIN_CREATING"
	// StateMainRunning means the main campaign is active.
	StateMainRunning WorkflowState = "MAIN_RUNNING"
	// StatePaused means provider delivery has been paused.
	StatePaused WorkflowState = "PAUSED"
	// StateManualReview means a human decision is required.
	StateManualReview WorkflowState = "NEEDS_MANUAL_REVIEW"
	// StateFailed means the workflow cannot proceed automatically.
	StateFailed WorkflowState = "FAILED"
)

// IsValid reports whether state is part of the supported workflow lifecycle.
func (s WorkflowState) IsValid() bool {
	switch s {
	case StateAccountConnected, StateAssetsSynced, StatePreflightPassed,
		StateSeedPending, StateSeedCreating, StateSeedRunning, StateSeedCompleted,
		StateMainPending, StateMainValidating, StateMainCreating, StateMainRunning,
		StatePaused, StateManualReview, StateFailed:
		return true
	default:
		return false
	}
}

// CampaignWorkflow tracks one independent automation flow per advertising account.
type CampaignWorkflow struct {
	ID                     string        `json:"id"`
	OrganizationID         string        `json:"organization_id"`
	AdAccountID            string        `json:"ad_account_id"`
	PageExternalID         string        `json:"page_external_id"`
	PixelExternalID        string        `json:"pixel_external_id,omitempty"`
	PixelEvent             string        `json:"pixel_event,omitempty"`
	ExistingPostID         string        `json:"existing_post_id,omitempty"`
	SeedCampaignExternalID string        `json:"seed_campaign_external_id,omitempty"`
	MainCampaignExternalID string        `json:"main_campaign_external_id,omitempty"`
	SeedSpendLimitUSD      float64       `json:"seed_spend_limit_usd"`
	State                  WorkflowState `json:"state"`
	LastError              string        `json:"last_error,omitempty"`
	CreatedAt              time.Time     `json:"created_at"`
	UpdatedAt              time.Time     `json:"updated_at"`
}

// CampaignMetrics is a point-in-time measurement used by workflow and rule decisions.
type CampaignMetrics struct {
	SpendUSD            float64   `json:"spend_usd"`
	Registrations       int64     `json:"registrations"`
	Deposits            int64     `json:"deposits"`
	CostPerRegistration float64   `json:"cost_per_registration"`
	CostPerDeposit      float64   `json:"cost_per_deposit"`
	CapturedAt          time.Time `json:"captured_at"`
}

// RuleMetric identifies the metric read by an automation rule.
type RuleMetric string

const (
	// MetricSpend compares total spend.
	MetricSpend RuleMetric = "SPEND"
	// MetricCostPerRegistration compares acquisition cost per registration.
	MetricCostPerRegistration RuleMetric = "COST_PER_REGISTRATION"
	// MetricCostPerDeposit compares acquisition cost per deposit.
	MetricCostPerDeposit RuleMetric = "COST_PER_DEPOSIT"
	// MetricRegistrations compares registration volume.
	MetricRegistrations RuleMetric = "REGISTRATIONS"
	// MetricDeposits compares deposit volume.
	MetricDeposits RuleMetric = "DEPOSITS"
)

// RuleOperator identifies a supported numeric comparison.
type RuleOperator string

const (
	// OperatorGTE matches values greater than or equal to the threshold.
	OperatorGTE RuleOperator = ">="
	// OperatorLTE matches values less than or equal to the threshold.
	OperatorLTE RuleOperator = "<="
	// OperatorGT matches values greater than the threshold.
	OperatorGT RuleOperator = ">"
	// OperatorLT matches values less than the threshold.
	OperatorLT RuleOperator = "<"
)

// RuleAction identifies a deterministic action selected by a matching rule.
type RuleAction string

const (
	// ActionPauseSeedCreateMain advances a qualified seed into main-campaign creation.
	ActionPauseSeedCreateMain RuleAction = "PAUSE_SEED_CREATE_MAIN"
	// ActionPauseCampaign stops provider delivery.
	ActionPauseCampaign RuleAction = "PAUSE_CAMPAIGN"
	// ActionScaleBudget changes budget within configured guardrails.
	ActionScaleBudget RuleAction = "SCALE_BUDGET"
	// ActionManualReview routes the workflow to a human operator.
	ActionManualReview RuleAction = "MANUAL_REVIEW"
)

// Rule describes one guarded metric comparison.
type Rule struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Metric         RuleMetric   `json:"metric"`
	Operator       RuleOperator `json:"operator"`
	Threshold      float64      `json:"threshold"`
	MinimumSpend   float64      `json:"minimum_spend"`
	MinimumSamples int64        `json:"minimum_samples"`
	Action         RuleAction   `json:"action"`
	Enabled        bool         `json:"enabled"`
}
