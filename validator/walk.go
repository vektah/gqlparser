package validator

import (
	"context"
	"fmt"

	"github.com/vektah/gqlparser/ast"
)

type Events struct {
	operationVisitor      []func(walker *Walker, operation *ast.OperationDefinition)
	operationLeaveVisitor []func(walker *Walker, operation *ast.OperationDefinition)
	field                 []func(walker *Walker, field *ast.Field)
	fragment              []func(walker *Walker, fragment *ast.FragmentDefinition)
	inlineFragment        []func(walker *Walker, inlineFragment *ast.InlineFragment)
	fragmentSpread        []func(walker *Walker, fragmentSpread *ast.FragmentSpread)
	directive             []func(walker *Walker, directive *ast.Directive)
	directiveList         []func(walker *Walker, directives []*ast.Directive)
	value                 []func(walker *Walker, value *ast.Value)
}

func (o *Events) OnOperation(f func(walker *Walker, operation *ast.OperationDefinition)) {
	o.operationVisitor = append(o.operationVisitor, f)
}
func (o *Events) OnOperationLeave(f func(walker *Walker, operation *ast.OperationDefinition)) {
	o.operationLeaveVisitor = append(o.operationLeaveVisitor, f)
}
func (o *Events) OnField(f func(walker *Walker, field *ast.Field)) {
	o.field = append(o.field, f)
}
func (o *Events) OnFragment(f func(walker *Walker, fragment *ast.FragmentDefinition)) {
	o.fragment = append(o.fragment, f)
}
func (o *Events) OnInlineFragment(f func(walker *Walker, inlineFragment *ast.InlineFragment)) {
	o.inlineFragment = append(o.inlineFragment, f)
}
func (o *Events) OnFragmentSpread(f func(walker *Walker, fragmentSpread *ast.FragmentSpread)) {
	o.fragmentSpread = append(o.fragmentSpread, f)
}
func (o *Events) OnDirective(f func(walker *Walker, directive *ast.Directive)) {
	o.directive = append(o.directive, f)
}
func (o *Events) OnDirectiveList(f func(walker *Walker, directives []*ast.Directive)) {
	o.directiveList = append(o.directiveList, f)
}
func (o *Events) OnValue(f func(walker *Walker, value *ast.Value)) {
	o.value = append(o.value, f)
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
	for _, varDef := range operation.VariableDefinitions {
		varDef.Definition = w.Schema.Types[varDef.Type.Name()]

		if varDef.DefaultValue != nil {
			varDef.DefaultValue.ExpectedType = varDef.Type
			varDef.DefaultValue.Definition = w.Schema.Types[varDef.Type.Name()]
		}
	}

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
		if varDef.DefaultValue != nil {
			w.walkValue(varDef.DefaultValue)
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
	def := w.Schema.Types[it.TypeCondition.Name()]

	it.Definition = def

	w.walkDirectives(def, it.Directives, ast.LocationFragmentDefinition)

	for _, v := range w.Observers.fragment {
		v(w, it)
	}

	for _, child := range it.SelectionSet {
		w.walkSelection(def, child)
	}
}

func (w *Walker) walkDirectives(parentDef *ast.Definition, directives []*ast.Directive, location ast.DirectiveLocation) {
	for _, dir := range directives {
		def := w.Schema.Directives[dir.Name]
		dir.Definition = def
		dir.ParentDefinition = parentDef
		dir.Location = location
		for _, v := range w.Observers.directive {
			v(w, dir)
		}

		for _, arg := range dir.Arguments {
			var argDef *ast.FieldDefinition
			if def != nil {
				argDef = def.Arguments.ForName(arg.Name)
			}

			w.walkArgument(argDef, &arg)
		}
	}

	for _, v := range w.Observers.directiveList {
		v(w, directives)
	}
}

func (w *Walker) walkValue(value *ast.Value) {
	for _, v := range w.Observers.value {
		v(w, value)
	}

	if value.Kind == ast.ObjectValue {
		for _, child := range value.Children {
			if value.Definition != nil {
				fieldDef := value.Definition.Field(child.Name)
				if fieldDef != nil {
					child.Value.ExpectedType = fieldDef.Type
					child.Value.Definition = w.Schema.Types[fieldDef.Type.Name()]
				}
			}
			w.walkValue(child.Value)
		}
	}

	if value.Kind == ast.ListValue {
		for _, child := range value.Children {
			if listType, isList := value.ExpectedType.(ast.ListType); isList {
				child.Value.ExpectedType = listType.Type
				child.Value.Definition = value.Definition
			}

			w.walkValue(child.Value)
		}
	}
}

func (w *Walker) walkArgument(argDef *ast.FieldDefinition, arg *ast.Argument) {
	if argDef != nil {
		arg.Value.ExpectedType = argDef.Type
		arg.Value.Definition = w.Schema.Types[argDef.Type.Name()]
	}

	w.walkValue(arg.Value)
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

		it.Definition = def
		it.ObjectDefinition = parentDef

		for _, v := range w.Observers.field {
			v(w, &it)
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
		it.ObjectDefinition = parentDef
		for _, v := range w.Observers.inlineFragment {
			v(w, &it)
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
		it.Definition = def
		it.ObjectDefinition = parentDef

		for _, v := range w.Observers.fragmentSpread {
			v(w, &it)
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
