package automation

import "testing"

func TestEvaluate(t *testing.T) {
	t.Parallel()

	rule := AutomationRule{
		ID: "pause-expensive", Metric: MetricCostPerRegistration, Operator: OperatorGT,
		Threshold: 8, MinimumSpend: 50, MinimumSamples: 3, Action: ActionPauseCampaign, Enabled: true,
	}
	tests := []struct {
		name    string
		metrics CampaignMetrics
		matched bool
	}{
		{name: "waits for spend", metrics: CampaignMetrics{SpendUSD: 10, Registrations: 5, CostPerRegistration: 20}},
		{name: "waits for samples", metrics: CampaignMetrics{SpendUSD: 70, Registrations: 2, CostPerRegistration: 20}},
		{
			name: "matches guarded metric",
			metrics: CampaignMetrics{
				SpendUSD: 70, Registrations: 5, CostPerRegistration: 9.5,
			},
			matched: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			decision := Evaluate(rule, test.metrics)
			if decision.Matched != test.matched {
				t.Errorf("Evaluate() matched = %t, expected %t", decision.Matched, test.matched)
			}
		})
	}
}
