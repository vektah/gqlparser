package validator

type ValidateOption struct {
	DisableSuggestion bool `yaml:"disableSuggestion"`
}

func NewDefaultValidateOption() ValidateOption {
	return ValidateOption{}
}

func (o ValidateOption) IsDisableSuggestion() bool {
	return o.DisableSuggestion
}

type ValidateOptionFactor interface {
	Apply(option ValidateOption) ValidateOption
}

type DisableSuggestion struct{}

func (DisableSuggestion) Apply(option ValidateOption) ValidateOption {
	option.DisableSuggestion = true
	return option
}
