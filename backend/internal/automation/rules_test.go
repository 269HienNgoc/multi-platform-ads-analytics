package automation

import (
	"testing"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain"
)

func TestEvaluateWaitsForMinimumSpend(t *testing.T) {
	rule := domain.AutomationRule{
		ID: "pause-expensive",
		Metric: domain.MetricCostPerRegistration,
		Operator: domain.OperatorGT,
		Threshold: 8,
		MinimumSpend: 50,
		Action: domain.ActionPauseCampaign,
		Enabled: true,
	}

	decision := Evaluate(rule, domain.CampaignMetrics{SpendUSD: 10, CostPerRegistration: 20})
	if decision.Matched {
		t.Fatal("rule must not match before minimum spend")
	}
}

func TestEvaluateMatchesExpensiveRegistration(t *testing.T) {
	rule := domain.AutomationRule{
		ID: "pause-expensive",
		Metric: domain.MetricCostPerRegistration,
		Operator: domain.OperatorGT,
		Threshold: 8,
		MinimumSpend: 50,
		MinimumSamples: 3,
		Action: domain.ActionPauseCampaign,
		Enabled: true,
	}

	decision := Evaluate(rule, domain.CampaignMetrics{
		SpendUSD: 70,
		Registrations: 5,
		CostPerRegistration: 9.5,
	})
	if !decision.Matched || decision.Action != domain.ActionPauseCampaign {
		t.Fatalf("expected pause decision, got %#v", decision)
	}
}
