package domain

import "time"

type WorkflowState string

const (
	StateAccountConnected  WorkflowState = "ACCOUNT_CONNECTED"
	StateAssetsSynced      WorkflowState = "ASSETS_SYNCED"
	StatePreflightPassed   WorkflowState = "PREFLIGHT_PASSED"
	StateSeedPending       WorkflowState = "SEED_PENDING"
	StateSeedCreating      WorkflowState = "SEED_CREATING"
	StateSeedRunning       WorkflowState = "SEED_RUNNING"
	StateSeedCompleted     WorkflowState = "SEED_COMPLETED"
	StateMainPending       WorkflowState = "MAIN_PENDING"
	StateMainValidating    WorkflowState = "MAIN_VALIDATING"
	StateMainCreating      WorkflowState = "MAIN_CREATING"
	StateMainRunning       WorkflowState = "MAIN_RUNNING"
	StatePaused            WorkflowState = "PAUSED"
	StateManualReview      WorkflowState = "NEEDS_MANUAL_REVIEW"
	StateFailed            WorkflowState = "FAILED"
)

type CampaignWorkflow struct {
	ID                string        `json:"id"`
	OrganizationID    string        `json:"organizationId"`
	AdAccountID       string        `json:"adAccountId"`
	PageID            string        `json:"pageId"`
	PixelID           string        `json:"pixelId,omitempty"`
	PixelEvent        string        `json:"pixelEvent,omitempty"`
	ExistingPostID    string        `json:"existingPostId,omitempty"`
	SeedCampaignID    string        `json:"seedCampaignId,omitempty"`
	MainCampaignID    string        `json:"mainCampaignId,omitempty"`
	SeedSpendLimitUSD float64       `json:"seedSpendLimitUsd"`
	State             WorkflowState `json:"state"`
	LastError         string        `json:"lastError,omitempty"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type BulkWorkflowRequest struct {
	OrganizationID    string   `json:"organizationId"`
	AdAccountIDs      []string `json:"adAccountIds"`
	PageID            string   `json:"pageId"`
	PixelID           string   `json:"pixelId,omitempty"`
	PixelEvent        string   `json:"pixelEvent,omitempty"`
	ExistingPostID    string   `json:"existingPostId,omitempty"`
	SeedSpendLimitUSD float64  `json:"seedSpendLimitUsd"`
}

type CampaignMetrics struct {
	SpendUSD            float64   `json:"spendUsd"`
	Registrations       int       `json:"registrations"`
	Deposits            int       `json:"deposits"`
	CostPerRegistration float64   `json:"costPerRegistration"`
	CostPerDeposit      float64   `json:"costPerDeposit"`
	CapturedAt          time.Time `json:"capturedAt"`
}

type RuleMetric string

const (
	MetricSpend               RuleMetric = "SPEND"
	MetricCostPerRegistration RuleMetric = "COST_PER_REGISTRATION"
	MetricCostPerDeposit      RuleMetric = "COST_PER_DEPOSIT"
	MetricRegistrations       RuleMetric = "REGISTRATIONS"
	MetricDeposits            RuleMetric = "DEPOSITS"
)

type RuleOperator string

const (
	OperatorGTE RuleOperator = ">="
	OperatorLTE RuleOperator = "<="
	OperatorGT  RuleOperator = ">"
	OperatorLT  RuleOperator = "<"
)

type RuleAction string

const (
	ActionPauseSeedCreateMain RuleAction = "PAUSE_SEED_CREATE_MAIN"
	ActionPauseCampaign       RuleAction = "PAUSE_CAMPAIGN"
	ActionScaleBudget         RuleAction = "SCALE_BUDGET"
	ActionManualReview        RuleAction = "MANUAL_REVIEW"
)

type AutomationRule struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Metric         RuleMetric   `json:"metric"`
	Operator       RuleOperator `json:"operator"`
	Threshold      float64      `json:"threshold"`
	MinimumSpend   float64      `json:"minimumSpend"`
	MinimumSamples int          `json:"minimumSamples"`
	Action         RuleAction   `json:"action"`
	Enabled        bool         `json:"enabled"`
}
