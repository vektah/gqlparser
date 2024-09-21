package validator

type ValidateOption struct {
	Suggestion SuggestionOption `yaml:"suggestion"`
}

type SuggestionOption struct {
	DisableTypeNamesSuggestion  bool `yaml:"disableTypeNamesSuggestion"`
	DisableFieldNamesSuggestion bool `yaml:"disableFieldNamesSuggestion"`
}

func (o *ValidateOption) IsDisableTypeNamesSuggestion() bool {
	if o == nil {
		return false
	}

	return o.Suggestion.DisableTypeNamesSuggestion
}

func (o *ValidateOption) IsDisableFieldNamesSuggestion() bool {
	if o == nil {
		return false
	}

	return o.Suggestion.DisableFieldNamesSuggestion
}
