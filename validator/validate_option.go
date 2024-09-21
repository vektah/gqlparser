package validator

type ValidateOption struct {
	DisableSuggestion bool `yaml:"disableSuggestion"`
}

func (o *ValidateOption) IsDisableSuggestion() bool {
	if o == nil {
		return false
	}

	return o.DisableSuggestion
}
