package collect

type (
	TaskModle struct {
		Property
		Root  string      `json:"root_script"`
		Rules []RuleModle `json:"rule"`
	}
	RuleModle struct {
		Name      string `json:"name"`
		ParseFunc string `json:"parse_script"`
	}
)
