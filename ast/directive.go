package ast

type DirectiveLocation string

const (
	// Executable
	LocationQuery              DirectiveLocation = `QUERY`
	LocationMutation           DirectiveLocation = `MUTATION`
	LocationSubscription       DirectiveLocation = `SUBSCRIPTION`
	LocationField              DirectiveLocation = `FIELD`
	LocationFragmentDefinition DirectiveLocation = `FRAGMENT_DEFINITION`
	LocationFragmentSpread     DirectiveLocation = `FRAGMENT_SPREAD`
	LocationInlineFragment     DirectiveLocation = `INLINE_FRAGMENT`

	// Type System
	LocationSchema               DirectiveLocation = `SCHEMA`
	LocationScalar               DirectiveLocation = `SCALAR`
	LocationObject               DirectiveLocation = `OBJECT`
	LocationFieldDefinition      DirectiveLocation = `FIELD_DEFINITION`
	LocationArgumentDefinition   DirectiveLocation = `ARGUMENT_DEFINITION`
	LocationInterface            DirectiveLocation = `INTERFACE`
	LocationUnion                DirectiveLocation = `UNION`
	LocationEnum                 DirectiveLocation = `ENUM`
	LocationEnumValue            DirectiveLocation = `ENUM_VALUE`
	LocationInputObject          DirectiveLocation = `INPUT_OBJECT`
	LocationInputFieldDefinition DirectiveLocation = `INPUT_FIELD_DEFINITION`
)

type Directive struct {
	Name      string
	Arguments []Argument

	// Requires validation
	ParentDefinition *Definition
	Definition       *DirectiveDefinition
	Location         DirectiveLocation
}

type Directives []*Directive

func (d Directives) Get(name string) *Directive {
	for _, directive := range d {
		if directive.Name == name {
			return directive
		}
	}
	return nil
}

func (d Directive) GetArg(name string) *Argument {
	for i := range d.Arguments {
		if d.Arguments[i].Name == name {
			return &d.Arguments[i]
		}
	}
	return nil
}
