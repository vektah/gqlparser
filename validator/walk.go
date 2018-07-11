package validator

import (
	"context"
	"fmt"

	"github.com/vektah/gqlparser"
)

type Events struct {
	operationCount int

	operationVisitor []func(walker *Walker, operation *gqlparser.OperationDefinition)
	field            []func(walker *Walker, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field)
	fragment         []func(walker *Walker, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition)
	inlineFragment   []func(walker *Walker, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment)
	fragmentSpread   []func(walker *Walker, parentDef *gqlparser.Definition, fragmentDef *gqlparser.FragmentDefinition, fragmentSpread *gqlparser.FragmentSpread)
	directive        []func(walker *Walker, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation)
	directiveList    []func(walker *Walker, parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation)
	value            []func(walker *Walker, value gqlparser.Value)
}

func (o *Events) OnOperation(f func(walker *Walker, operation *gqlparser.OperationDefinition)) {
	o.operationVisitor = append(o.operationVisitor, f)
}
func (o *Events) OnField(f func(walker *Walker, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field)) {
	o.field = append(o.field, f)
}
func (o *Events) OnFragment(f func(walker *Walker, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition)) {
	o.fragment = append(o.fragment, f)
}
func (o *Events) OnInlineFragment(f func(walker *Walker, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment)) {
	o.inlineFragment = append(o.inlineFragment, f)
}
func (o *Events) OnFragmentSpread(f func(walker *Walker, parentDef *gqlparser.Definition, fragmentDef *gqlparser.FragmentDefinition, fragmentSpread *gqlparser.FragmentSpread)) {
	o.fragmentSpread = append(o.fragmentSpread, f)
}
func (o *Events) OnDirective(f func(walker *Walker, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation)) {
	o.directive = append(o.directive, f)
}
func (o *Events) OnDirectiveList(f func(walker *Walker, parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation)) {
	o.directiveList = append(o.directiveList, f)
}
func (o *Events) OnValue(f func(walker *Walker, value gqlparser.Value)) {
	o.value = append(o.value, f)
}

func Walk(schema *gqlparser.Schema, document *gqlparser.QueryDocument, observers *Events) {
	w := Walker{
		Observers: observers,
		Schema:    schema,
		Document:  document,
	}
	w.walk()
}

type Walker struct {
	Context   context.Context
	Observers *Events
	Schema    *gqlparser.Schema
	Document  *gqlparser.QueryDocument
}

func (w *Walker) walk() {
	w.Observers.operationCount = len(w.Document.Operations)

	for _, child := range w.Document.Operations {
		w.walkOperation(&child)
	}
	for _, child := range w.Document.Fragments {
		w.walkFragment(&child)
	}
}

func (w *Walker) walkOperation(operation *gqlparser.OperationDefinition) {
	for _, v := range w.Observers.operationVisitor {
		v(w, operation)
	}

	var def *gqlparser.Definition
	var loc gqlparser.DirectiveLocation
	switch operation.Operation {
	case gqlparser.Query, "":
		def = w.Schema.Query
		loc = gqlparser.LocationQuery
	case gqlparser.Mutation:
		def = w.Schema.Mutation
		loc = gqlparser.LocationMutation
	case gqlparser.Subscription:
		def = w.Schema.Subscription
		loc = gqlparser.LocationSubscription
	}

	w.walkDirectives(def, operation.Directives, loc)

	for _, v := range operation.SelectionSet {
		w.walkSelection(def, v)
	}
}

func (w *Walker) walkFragment(it *gqlparser.FragmentDefinition) {
	parentDef := w.Schema.Types[it.TypeCondition.Name()]

	w.walkDirectives(parentDef, it.Directives, gqlparser.LocationFragmentDefinition)

	for _, v := range w.Observers.fragment {
		v(w, parentDef, it)
	}

	for _, child := range it.SelectionSet {
		w.walkSelection(parentDef, child)
	}
}

func (w *Walker) walkDirectives(parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation) {
	for _, v := range w.Observers.directiveList {
		v(w, parentDef, directives, location)
	}

	for _, dir := range directives {
		def := w.Schema.Directives[dir.Name]
		for _, v := range w.Observers.directive {
			v(w, parentDef, def, &dir, location)
		}
	}
}

func (w *Walker) walkValue(value gqlparser.Value) {
	for _, v := range w.Observers.value {
		v(w, value)
	}

	switch value := value.(type) {
	case gqlparser.ListValue:
		for _, v := range value {
			w.walkValue(v)
		}
	case gqlparser.ObjectValue:
		for _, v := range value {
			w.walkValue(v.Value)
		}
	}
}

func (w *Walker) walkSelection(parentDef *gqlparser.Definition, it gqlparser.Selection) {
	switch it := it.(type) {
	case gqlparser.Field:
		var def *gqlparser.FieldDefinition
		if it.Name == "__typename" {
			def = &gqlparser.FieldDefinition{
				Name: "__typename",
				Type: gqlparser.NamedType("String"),
			}
		} else if parentDef != nil {
			def = parentDef.Field(it.Name)
		}

		for _, v := range w.Observers.field {
			v(w, parentDef, def, &it)
		}

		var nextParentDef *gqlparser.Definition
		if def != nil {
			nextParentDef = w.Schema.Types[def.Type.Name()]
		}

		for _, arg := range it.Arguments {
			w.walkValue(arg.Value)
		}

		for _, sel := range it.SelectionSet {
			w.walkSelection(nextParentDef, sel)
		}

		w.walkDirectives(nextParentDef, it.Directives, gqlparser.LocationField)

	case gqlparser.InlineFragment:
		for _, v := range w.Observers.inlineFragment {
			v(w, parentDef, &it)
		}

		var nextParentDef *gqlparser.Definition
		if it.TypeCondition.Name() != "" {
			nextParentDef = w.Schema.Types[it.TypeCondition.Name()]
		}

		w.walkDirectives(nextParentDef, it.Directives, gqlparser.LocationInlineFragment)

		for _, sel := range it.SelectionSet {
			w.walkSelection(nextParentDef, sel)
		}

	case gqlparser.FragmentSpread:
		def := w.Document.GetFragment(it.Name)

		for _, v := range w.Observers.fragmentSpread {
			v(w, parentDef, def, &it)
		}

		var nextParentDef *gqlparser.Definition
		if def != nil {
			nextParentDef = w.Schema.Types[def.TypeCondition.Name()]
		}

		w.walkDirectives(nextParentDef, it.Directives, gqlparser.LocationFragmentSpread)

		if def != nil {
			for _, sel := range def.SelectionSet {
				w.walkSelection(nextParentDef, sel)
			}
		}

	default:
		panic(fmt.Errorf("unsupported %T", it))

	}
}
