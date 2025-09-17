package plan

// SimulationPlan is the top-level structure of a playbook YAML file.
// It directly maps to the root of the YAML document.
type SimulationPlan struct {
	Plan []SimulationStep `yaml:"simulation_plan"`
}

// SimulationStep represents a single action or delay in the playbook.
type SimulationStep struct {
	Name     string `yaml:"name"`
	Action   string `yaml:"action"`
	Category string `yaml:"category"`
	Duration string `yaml:"duration"`

	Args map[string]interface{} `yaml:"args"` // Args used for commands that are not category-based (e.g., hop, morph).
}
