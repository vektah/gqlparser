package validator

import (
	"context"
	"fmt"

	"github.com/vektah/gqlparser/ast"
)

type Events struct {
	operationVisitor      []func(walker *Walker, operation *ast.OperationDefinition)
	operationLeaveVisitor []func(walker *Walker, operation *ast.OperationDefinition)
	field                 []func(walker *Walker, parentDef *ast.Definition, fieldDef *ast.FieldDefinition, field *ast.Field)
	fragment              []func(walker *Walker, parentDef *ast.Definition, fragment *ast.FragmentDefinition)
	inlineFragment        []func(walker *Walker, parentDef *ast.Definition, inlineFragment *ast.InlineFragment)
	fragmentSpread        []func(walker *Walker, parentDef *ast.Definition, fragmentDef *ast.FragmentDefinition, fragmentSpread *ast.FragmentSpread)
	directive             []func(walker *Walker, parentDef *ast.Definition, directiveDef *ast.DirectiveDefinition, directive *ast.Directive, location ast.DirectiveLocation)
	directiveList         []func(walker *Walker, parentDef *ast.Definition, directives []ast.Directive, location ast.DirectiveLocation)
	value                 []func(walker *Walker, valueType ast.Type, def *ast.Definition, value ast.Value)
	variable              []func(walker *Walker, valueType ast.Type, def *ast.Definition, variable ast.VariableDefinition)
}

func (o *Events) OnOperation(f func(walker *Walker, operation *ast.OperationDefinition)) {
	o.operationVisitor = append(o.operationVisitor, f)
}
func (o *Events) OnOperationLeave(f func(walker *Walker, operation *ast.OperationDefinition)) {
	o.operationLeaveVisitor = append(o.operationLeaveVisitor, f)
}
func (o *Events) OnField(f func(walker *Walker, parentDef *ast.Definition, fieldDef *ast.FieldDefinition, field *ast.Field)) {
	o.field = append(o.field, f)
}
func (o *Events) OnFragment(f func(walker *Walker, parentDef *ast.Definition, fragment *ast.FragmentDefinition)) {
	o.fragment = append(o.fragment, f)
}
func (o *Events) OnInlineFragment(f func(walker *Walker, parentDef *ast.Definition, inlineFragment *ast.InlineFragment)) {
	o.inlineFragment = append(o.inlineFragment, f)
}
func (o *Events) OnFragmentSpread(f func(walker *Walker, parentDef *ast.Definition, fragmentDef *ast.FragmentDefinition, fragmentSpread *ast.FragmentSpread)) {
	o.fragmentSpread = append(o.fragmentSpread, f)
}
func (o *Events) OnDirective(f func(walker *Walker, parentDef *ast.Definition, directiveDef *ast.DirectiveDefinition, directive *ast.Directive, location ast.DirectiveLocation)) {
	o.directive = append(o.directive, f)
}
func (o *Events) OnDirectiveList(f func(walker *Walker, parentDef *ast.Definition, directives []ast.Directive, location ast.DirectiveLocation)) {
	o.directiveList = append(o.directiveList, f)
}
func (o *Events) OnValue(f func(walker *Walker, valueType ast.Type, def *ast.Definition, value ast.Value)) {
	o.value = append(o.value, f)
}
func (o *Events) OnVariable(f func(walker *Walker, valueType ast.Type, def *ast.Definition, variable ast.VariableDefinition)) {
	o.variable = append(o.variable, f)
}

func Walk(schema *ast.Schema, document *ast.QueryDocument, observers *Events) {
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
	Schema    *ast.Schema
	Document  *ast.QueryDocument

	validatedFragmentSpreads map[string]bool
}

func (w *Walker) walk() {
	for _, child := range w.Document.Operations {
		w.validatedFragmentSpreads = make(map[string]bool)
		w.walkOperation(&child)
	}
	for _, child := range w.Document.Fragments {
		w.validatedFragmentSpreads = make(map[string]bool)
		w.walkFragment(&child)
	}
}

func (w *Walker) walkOperation(operation *ast.OperationDefinition) {
	for _, v := range w.Observers.operationVisitor {
		v(w, operation)
	}

	var def *ast.Definition
	var loc ast.DirectiveLocation
	switch operation.Operation {
	case ast.Query, "":
		def = w.Schema.Query
		loc = ast.LocationQuery
	case ast.Mutation:
		def = w.Schema.Mutation
		loc = ast.LocationMutation
	case ast.Subscription:
		def = w.Schema.Subscription
		loc = ast.LocationSubscription
	}

	w.walkDirectives(def, operation.Directives, loc)

	for _, varDef := range operation.VariableDefinitions {
		typeDef := w.Schema.Types[varDef.Type.Name()]
		for _, v := range w.Observers.variable {
			v(w, varDef.Type, typeDef, varDef)
		}
		if varDef.DefaultValue != nil {
			w.walkValue(varDef.Type, varDef.DefaultValue)
		}
	}

	for _, v := range operation.SelectionSet {
		w.walkSelection(def, v)
	}

	for _, v := range w.Observers.operationLeaveVisitor {
		v(w, operation)
	}
}

func (w *Walker) walkFragment(it *ast.FragmentDefinition) {
	parentDef := w.Schema.Types[it.TypeCondition.Name()]

	w.walkDirectives(parentDef, it.Directives, ast.LocationFragmentDefinition)

	for _, v := range w.Observers.fragment {
		v(w, parentDef, it)
	}

	for _, child := range it.SelectionSet {
		w.walkSelection(parentDef, child)
	}
}

func (w *Walker) walkDirectives(parentDef *ast.Definition, directives []ast.Directive, location ast.DirectiveLocation) {
	for _, v := range w.Observers.directiveList {
		v(w, parentDef, directives, location)
	}

	for _, dir := range directives {
		def := w.Schema.Directives[dir.Name]
		for _, v := range w.Observers.directive {
			v(w, parentDef, def, &dir, location)
		}

		for _, arg := range dir.Arguments {
			var argDef *ast.FieldDefinition
			if def != nil {
				argDef = def.Arguments.ForName(arg.Name)
			}

			w.walkArgument(argDef, &arg)
		}
	}
}

func (w *Walker) walkValue(valueType ast.Type, value ast.Value) {
	var def *ast.Definition
	if valueType != nil {
		def = w.Schema.Types[valueType.Name()]
	}

	for _, v := range w.Observers.value {
		v(w, valueType, def, value)
	}

	if obj, isObj := value.(ast.ObjectValue); isObj {
		for _, v := range obj {
			var fieldType ast.Type
			if def != nil {
				fieldDef := def.Field(v.Name)
				if fieldDef != nil {
					fieldType = fieldDef.Type
				}
			}
			w.walkValue(fieldType, v.Value)
		}
	}
}

func (w *Walker) walkArgument(argDef *ast.FieldDefinition, arg *ast.Argument) {
	var argType ast.Type
	if argDef != nil {
		argType = argDef.Type
	}

	w.walkValue(argType, arg.Value)

}

func (w *Walker) walkSelection(parentDef *ast.Definition, it ast.Selection) {
	switch it := it.(type) {
	case ast.Field:
		var def *ast.FieldDefinition
		if it.Name == "__typename" {
			def = &ast.FieldDefinition{
				Name: "__typename",
				Type: ast.NamedType("String"),
			}
		} else if parentDef != nil {
			def = parentDef.Field(it.Name)
		}

		for _, v := range w.Observers.field {
			v(w, parentDef, def, &it)
		}

		var nextParentDef *ast.Definition
		if def != nil {
			nextParentDef = w.Schema.Types[def.Type.Name()]
		}

		for _, arg := range it.Arguments {
			var argDef *ast.FieldDefinition
			if def != nil {
				argDef = def.Arguments.ForName(arg.Name)
			}

			w.walkArgument(argDef, &arg)
		}

		for _, sel := range it.SelectionSet {
			w.walkSelection(nextParentDef, sel)
		}

		w.walkDirectives(nextParentDef, it.Directives, ast.LocationField)

	case ast.InlineFragment:
		for _, v := range w.Observers.inlineFragment {
			v(w, parentDef, &it)
		}

		var nextParentDef *ast.Definition
		if it.TypeCondition.Name() != "" {
			nextParentDef = w.Schema.Types[it.TypeCondition.Name()]
		}

		w.walkDirectives(nextParentDef, it.Directives, ast.LocationInlineFragment)

		for _, sel := range it.SelectionSet {
			w.walkSelection(nextParentDef, sel)
		}

	case ast.FragmentSpread:
		def := w.Document.GetFragment(it.Name)

		for _, v := range w.Observers.fragmentSpread {
			v(w, parentDef, def, &it)
		}

		var nextParentDef *ast.Definition
		if def != nil {
			nextParentDef = w.Schema.Types[def.TypeCondition.Name()]
		}

		w.walkDirectives(nextParentDef, it.Directives, ast.LocationFragmentSpread)

		if def != nil && !w.validatedFragmentSpreads[def.Name] {
			// prevent inifinite recursion
			w.validatedFragmentSpreads[def.Name] = true

			for _, sel := range def.SelectionSet {
				w.walkSelection(nextParentDef, sel)
			}
		}

	default:
		panic(fmt.Errorf("unsupported %T", it))

	}
}
