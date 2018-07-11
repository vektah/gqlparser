package validator

import (
	"fmt"

	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/errors"
)

var operationVisitor []func(vctx *vctx, operation *gqlparser.OperationDefinition)
var fieldVisitors []func(vctx *vctx, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field)
var fragmentVisitors []func(vctx *vctx, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition)
var inlineFragmentVisitors []func(vctx *vctx, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment)
var directiveVisitors []func(vctx *vctx, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation)
var directiveListVisitors []func(vctx *vctx, parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation)

func init() {
	//fieldVisitors = append(fieldVisitors, func(vctx *vctx, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field) {
	//	fmt.Println("ENTER FIELD "+field.Name, parentDef, fieldDef)
	//})
}

type vctx struct {
	schema   *gqlparser.Schema
	document *gqlparser.QueryDocument
	errors   []errors.Validation
}

func (c *vctx) walk() {
	for _, child := range c.document.Operations {
		c.walkOperation(&child)
	}
	for _, child := range c.document.Fragments {
		c.walkFragment(&child)
	}
}

func (c *vctx) walkOperation(operation *gqlparser.OperationDefinition) {
	for _, v := range operationVisitor {
		v(c, operation)
	}

	var def *gqlparser.Definition
	var loc gqlparser.DirectiveLocation
	switch operation.Operation {
	case gqlparser.Query, "":
		def = c.schema.Query
		loc = gqlparser.LocationQuery
	case gqlparser.Mutation:
		def = c.schema.Mutation
		loc = gqlparser.LocationMutation
	case gqlparser.Subscription:
		def = c.schema.Subscription
		loc = gqlparser.LocationSubscription
	}

	c.walkDirectives(def, operation.Directives, loc)

	for _, v := range operation.SelectionSet {
		c.walkSelection(def, v)
	}
}

func (c *vctx) walkFragment(it *gqlparser.FragmentDefinition) {
	parentDef := c.schema.Types[it.TypeCondition.Name()]

	c.walkDirectives(parentDef, it.Directives, gqlparser.LocationFragmentDefinition)

	for _, v := range fragmentVisitors {
		v(c, parentDef, it)
	}

	for _, child := range it.SelectionSet {
		c.walkSelection(parentDef, child)
	}
}

func (c *vctx) walkDirectives(parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation) {
	for _, v := range directiveListVisitors {
		v(c, parentDef, directives, location)
	}

	for _, dir := range directives {
		def := c.schema.Directives[dir.Name]
		for _, v := range directiveVisitors {
			v(c, parentDef, def, &dir, location)
		}
	}
}

func (c *vctx) walkSelection(parentDef *gqlparser.Definition, it gqlparser.Selection) {
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

		for _, v := range fieldVisitors {
			v(c, parentDef, def, &it)
		}

		var nextParentDef *gqlparser.Definition
		if def != nil {
			nextParentDef = c.schema.Types[def.Type.Name()]
		}

		for _, sel := range it.SelectionSet {
			c.walkSelection(nextParentDef, sel)
		}

		c.walkDirectives(nextParentDef, it.Directives, gqlparser.LocationField)

	case gqlparser.InlineFragment:
		for _, v := range inlineFragmentVisitors {
			v(c, parentDef, &it)
		}

		var nextParentDef *gqlparser.Definition
		if it.TypeCondition.Name() != "" {
			nextParentDef = c.schema.Types[it.TypeCondition.Name()]
		}

		c.walkDirectives(nextParentDef, it.Directives, gqlparser.LocationInlineFragment)

		for _, sel := range it.SelectionSet {
			c.walkSelection(nextParentDef, sel)
		}

	case gqlparser.FragmentSpread:
		def := c.document.GetFragment(it.Name)

		var nextParentDef *gqlparser.Definition
		if def != nil {
			nextParentDef = c.schema.Types[def.TypeCondition.Name()]
		}

		c.walkDirectives(nextParentDef, it.Directives, gqlparser.LocationFragmentSpread)

		if def != nil {
			for _, sel := range def.SelectionSet {
				c.walkSelection(nextParentDef, sel)
			}
		}

	default:
		panic(fmt.Errorf("unsupported %T", it))

	}
}
