package validator

import "github.com/vektah/gqlparser"

func init() {
	directiveDecoratedVisitors = append(directiveDecoratedVisitors, knownDirectives)
}

func knownDirectives(ctx *vctx, target interface{}, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive) {
	if directiveDef == nil {
		ctx.errors = append(ctx.errors, Error(
			Rule("KnownDirectives"),
			Message(`Unknown directive "%s".`, directive.Name),
		))
		return
	}

	candidateLocation := getDirectiveLocation(target)
	if candidateLocation == "" {
		return
	}

	for _, loc := range directiveDef.Locations {
		if loc == candidateLocation {
			return
		}
	}

	ctx.errors = append(ctx.errors, Error(
		Rule("KnownDirectives"),
		Message(`Directive "%s" may not be used on %s.`, directive.Name, candidateLocation),
	))
}

func getDirectiveLocation(def interface{}) gqlparser.DirectiveLocation {

	switch def := def.(type) {
	case *gqlparser.OperationDefinition:
		switch def.Operation {
		case gqlparser.Query:
			return gqlparser.LocationQuery
		case gqlparser.Mutation:
			return gqlparser.LocationMutation
		case gqlparser.Subscription:
			return gqlparser.LocationSubscription
		}

	case *gqlparser.Field:
		return gqlparser.LocationField

	case *gqlparser.FragmentSpread:
		return gqlparser.LocationFragmentSpread

	case *gqlparser.InlineFragment:
		return gqlparser.LocationInlineFragment

	case *gqlparser.FragmentDefinition:
		return gqlparser.LocationFragmentDefinition

	case *gqlparser.SchemaDefinition:
		return gqlparser.LocationSchema

	case *gqlparser.Definition:
		switch def.Kind {
		case gqlparser.Scalar:
			return gqlparser.LocationScalar
		case gqlparser.Object:
			return gqlparser.LocationObject
		case gqlparser.Interface:
			return gqlparser.LocationInterface
		case gqlparser.Union:
			return gqlparser.LocationUnion
		case gqlparser.Enum:
			return gqlparser.LocationEnum
		case gqlparser.InputObject:
			return gqlparser.LocationInputObject
		}

	case *gqlparser.FieldDefinition:
		return gqlparser.LocationFieldDefinition

	case *gqlparser.EnumValueDefinition:
		return gqlparser.LocationEnumValue

		// TODO case Kind.INPUT_VALUE_DEFINITION:
	}

	return ""
}
