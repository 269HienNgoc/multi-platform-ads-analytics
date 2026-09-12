package automation

import "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain"

type Decision struct {
	Matched bool              `json:"matched"`
	Action  domain.RuleAction `json:"action,omitempty"`
	RuleID  string            `json:"ruleId,omitempty"`
}

func Evaluate(rule domain.AutomationRule, metrics domain.CampaignMetrics) Decision {
	if !rule.Enabled || metrics.SpendUSD < rule.MinimumSpend {
		return Decision{}
	}

	if rule.MinimumSamples > 0 {
		samples := metrics.Registrations
		if rule.Metric == domain.MetricCostPerDeposit || rule.Metric == domain.MetricDeposits {
			samples = metrics.Deposits
		}
		if samples < rule.MinimumSamples {
			return Decision{}
		}
	}

	value := metricValue(rule.Metric, metrics)
	matched := compare(value, rule.Operator, rule.Threshold)
	if !matched {
		return Decision{}
	}

	return Decision{Matched: true, Action: rule.Action, RuleID: rule.ID}
}

func metricValue(metric domain.RuleMetric, metrics domain.CampaignMetrics) float64 {
	switch metric {
	case domain.MetricSpend:
		return metrics.SpendUSD
	case domain.MetricCostPerRegistration:
		return metrics.CostPerRegistration
	case domain.MetricCostPerDeposit:
		return metrics.CostPerDeposit
	case domain.MetricRegistrations:
		return float64(metrics.Registrations)
	case domain.MetricDeposits:
		return float64(metrics.Deposits)
	default:
		return 0
	}
}

func compare(value float64, operator domain.RuleOperator, threshold float64) bool {
	switch operator {
	case domain.OperatorGTE:
		return value >= threshold
	case domain.OperatorLTE:
		return value <= threshold
	case domain.OperatorGT:
		return value > threshold
	case domain.OperatorLT:
		return value < threshold
	default:
		return false
	}
}
