package domain

// PrometheusRuleFile represents a Prometheus rule file structure
type PrometheusRuleFile struct {
	Groups []PrometheusRuleGroup `yaml:"groups"`
}

// PrometheusRuleGroup represents a group of Prometheus rules
type PrometheusRuleGroup struct {
	Name     string            `yaml:"name"`
	Interval string            `yaml:"interval,omitempty"`
	Rules    []PrometheusRule  `yaml:"rules"`
}

// PrometheusRule represents a single Prometheus alerting rule
type PrometheusRule struct {
	Alert       string            `yaml:"alert"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

// NewPrometheusRuleFile creates a new Prometheus rule file with given groups
func NewPrometheusRuleFile(groups []PrometheusRuleGroup) *PrometheusRuleFile {
	return &PrometheusRuleFile{
		Groups: groups,
	}
}

// NewPrometheusRuleGroup creates a new Prometheus rule group
func NewPrometheusRuleGroup(name string, rules []PrometheusRule) *PrometheusRuleGroup {
	return &PrometheusRuleGroup{
		Name:  name,
		Rules: rules,
	}
}
