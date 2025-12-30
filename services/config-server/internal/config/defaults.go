package config

// DefaultsConfig holds configuration for default templates
type DefaultsConfig struct {
	// AlertTemplatesFile is the path to the alert templates YAML file
	AlertTemplatesFile string `mapstructure:"alert_templates_file"`
	// ScriptTemplatesFile is the path to the script templates YAML file
	ScriptTemplatesFile string `mapstructure:"script_templates_file"`
	// ScriptsDir is the directory containing actual script files
	ScriptsDir string `mapstructure:"scripts_dir"`
}

// AlertTemplateYAML represents the structure of alert-templates.yaml
type AlertTemplateYAML struct {
	Templates []AlertTemplateEntry `yaml:"templates"`
}

// AlertTemplateEntry represents a single alert template entry in YAML
type AlertTemplateEntry struct {
	ID            string                 `yaml:"id"`
	Name          string                 `yaml:"name"`
	Description   string                 `yaml:"description"`
	Severity      string                 `yaml:"severity"`
	QueryTemplate string                 `yaml:"query_template"`
	DefaultConfig map[string]interface{} `yaml:"default_config"`
}

// ScriptTemplateYAML represents the structure of script-templates.yaml
type ScriptTemplateYAML struct {
	Templates []ScriptTemplateEntry `yaml:"templates"`
}

// ScriptTemplateEntry represents a single script template entry in YAML
type ScriptTemplateEntry struct {
	Name          string                 `yaml:"name"`
	Description   string                 `yaml:"description"`
	ScriptType    string                 `yaml:"script_type"`
	ScriptFile    string                 `yaml:"script_file"` // Relative path to script file
	Language      string                 `yaml:"language"`
	DefaultConfig map[string]interface{} `yaml:"default_config"`
	Version       string                 `yaml:"version"`
}
