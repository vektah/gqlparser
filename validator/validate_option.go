package validator

type ValidateOption struct {
	Suggestion SuggestionOption
}

type SuggestionOption struct {
	DisableTypeNamesSuggestion  bool
	DisableFieldNamesSuggestion bool
}
