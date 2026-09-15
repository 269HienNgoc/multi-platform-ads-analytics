package automation

// Decision is the deterministic result of evaluating one automation rule.
type Decision struct {
	Matched bool       `json:"matched"`
	Action  RuleAction `json:"action,omitempty"`
	RuleID  string     `json:"rule_id,omitempty"`
}

// Evaluate applies minimum-spend and sample guardrails before comparing a metric.
func Evaluate(rule AutomationRule, metrics CampaignMetrics) Decision {
	if !rule.Enabled || metrics.SpendUSD < rule.MinimumSpend {
		return Decision{}
	}

	samples := metrics.Registrations
	if rule.Metric == MetricCostPerDeposit || rule.Metric == MetricDeposits {
		samples = metrics.Deposits
	}
	if rule.MinimumSamples > 0 && samples < rule.MinimumSamples {
		return Decision{}
	}

	if !compare(metricValue(rule.Metric, metrics), rule.Operator, rule.Threshold) {
		return Decision{}
	}

	return Decision{Matched: true, Action: rule.Action, RuleID: rule.ID}
}

func metricValue(metric RuleMetric, metrics CampaignMetrics) float64 {
	switch metric {
	case MetricSpend:
		return metrics.SpendUSD
	case MetricCostPerRegistration:
		return metrics.CostPerRegistration
	case MetricCostPerDeposit:
		return metrics.CostPerDeposit
	case MetricRegistrations:
		return float64(metrics.Registrations)
	case MetricDeposits:
		return float64(metrics.Deposits)
	default:
		return 0
	}
}

func compare(value float64, operator RuleOperator, threshold float64) bool {
	switch operator {
	case OperatorGTE:
		return value >= threshold
	case OperatorLTE:
		return value <= threshold
	case OperatorGT:
		return value > threshold
	case OperatorLT:
		return value < threshold
	default:
		return false
	}
}
