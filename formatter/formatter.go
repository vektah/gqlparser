package formatter

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/vektah/gqlparser/ast"
)

type Formatter interface {
	FormatSchema(schema *ast.Schema) error
	FormatSchemaDocument(doc *ast.SchemaDocument) error
	FormatQueryDocument(doc *ast.QueryDocument) error
}

func NewFormatter(w io.Writer) Formatter {
	return &formatter{writer: w}
}

type formatter struct {
	writer io.Writer

	indent      int
	emitBuiltin bool

	padNext  bool
	lineHead bool
}

func (f *formatter) writeString(s string) {
	_, _ = f.writer.Write([]byte(s))
}

func (f *formatter) writeIndent() *formatter {
	if f.lineHead {
		f.writeString(strings.Repeat("\t", f.indent))
	}
	f.lineHead = false
	f.padNext = false

	return f
}

func (f *formatter) WriteNewline() *formatter {
	f.writeString("\n")
	f.lineHead = true
	f.padNext = false

	return f
}

func (f *formatter) WriteWord(word string) *formatter {
	if f.lineHead {
		f.writeIndent()
	}
	if f.padNext {
		f.writeString(" ")
	}
	f.writeString(strings.TrimSpace(word))
	f.padNext = true

	return f
}

func (f *formatter) WriteString(s string) *formatter {
	if f.lineHead {
		f.writeIndent()
	}
	if f.padNext {
		f.writeString(" ")
	}
	f.writeString(s)
	f.padNext = false

	return f
}

func (f *formatter) WriteDescription(s string) *formatter {
	if s == "" {
		return f
	}

	f.WriteString(`"""`).WriteNewline()

	ss := strings.Split(s, "\n")
	for _, s := range ss {
		f.WriteString(s).WriteNewline()
	}

	f.WriteString(`"""`).WriteNewline()

	return f
}

func (f *formatter) IncrementIndent() {
	f.indent++
}

func (f *formatter) DecrementIndent() {
	f.indent--
}

func (f *formatter) NoPadding() *formatter {
	f.padNext = false

	return f
}

func (f *formatter) NeedPadding() *formatter {
	f.padNext = true

	return f
}

func (f *formatter) FormatSchema(schema *ast.Schema) error {
	if schema == nil {
		return nil
	}

	var inSchema bool
	startSchema := func() {
		if !inSchema {
			inSchema = true

			f.WriteWord("schema").WriteString("{").WriteNewline()
			f.IncrementIndent()
		}
	}
	if schema.Query != nil && schema.Query.Name != "Query" {
		startSchema()
		f.WriteWord("query").NoPadding().WriteString(":").NeedPadding()
		f.WriteWord(schema.Query.Name).WriteNewline()
	}
	if schema.Mutation != nil && schema.Mutation.Name != "Mutation" {
		startSchema()
		f.WriteWord("mutation").NoPadding().WriteString(":").NeedPadding()
		f.WriteWord(schema.Mutation.Name).WriteNewline()
	}
	if schema.Subscription != nil && schema.Subscription.Name != "Subscription" {
		startSchema()
		f.WriteWord("subscription").NoPadding().WriteString(":").NeedPadding()
		f.WriteWord(schema.Subscription.Name).WriteNewline()
	}
	if inSchema {
		f.DecrementIndent()
		f.WriteString("}").WriteNewline()
	}

	directiveNames := make([]string, 0, len(schema.Directives))
	for name := range schema.Directives {
		directiveNames = append(directiveNames, name)
	}
	sort.Strings(directiveNames)
	for _, name := range directiveNames {
		err := f.FormatDirectiveDefinition(schema.Directives[name])
		if err != nil {
			return err
		}
	}

	typeNames := make([]string, 0, len(schema.Types))
	for name := range schema.Types {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)
	for _, name := range typeNames {
		err := f.FormatDefinition(schema.Types[name], false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatSchemaDocument(doc *ast.SchemaDocument) error {
	// TODO emit by position based order

	if err := f.FormatSchemaDefinitionList(doc.Schema, false); err != nil {
		return err
	}
	if err := f.FormatSchemaDefinitionList(doc.SchemaExtension, true); err != nil {
		return err
	}

	if err := f.FormatDirectiveDefinitionList(doc.Directives); err != nil {
		return err
	}

	if err := f.FormatDefinitionList(doc.Definitions, false); err != nil {
		return err
	}
	if err := f.FormatDefinitionList(doc.Extensions, true); err != nil {
		return err
	}

	return nil
}

func (f *formatter) FormatQueryDocument(doc *ast.QueryDocument) error {
	// TODO emit by position based order

	if err := f.FormatOperationList(doc.Operations); err != nil {
		return err
	}

	if err := f.FormatFragmentDefinitionList(doc.Fragments); err != nil {
		return err
	}

	return nil
}

func (f *formatter) FormatSchemaDefinitionList(lists ast.SchemaDefinitionList, extention bool) error {

	if len(lists) == 0 {
		return nil
	}

	if extention {
		f.WriteWord("extend")
	}
	f.WriteWord("schema").WriteString("{").WriteNewline()
	f.IncrementIndent()

	for _, def := range lists {
		if err := f.FormatSchemaDefinition(def); err != nil {
			return err
		}
	}

	f.DecrementIndent()
	f.WriteString("}").WriteNewline()

	return nil
}

func (f *formatter) FormatSchemaDefinition(def *ast.SchemaDefinition) error {
	f.WriteDescription(def.Description)

	if err := f.FormatDirectiveList(def.Directives); err != nil {
		return err
	}

	if err := f.FormatOperationTypeDefinitionList(def.OperationTypes); err != nil {
		return err
	}

	return nil
}

func (f *formatter) FormatOperationTypeDefinitionList(lists ast.OperationTypeDefinitionList) error {
	for _, def := range lists {
		if err := f.FormatOperationTypeDefinition(def); err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatOperationTypeDefinition(def *ast.OperationTypeDefinition) error {
	f.WriteWord(string(def.Operation)).NoPadding().WriteString(":").NeedPadding()
	f.WriteWord(def.Type)

	f.WriteNewline()

	return nil
}

func (f *formatter) FormatFieldList(fieldList ast.FieldList) error {
	if len(fieldList) == 0 {
		return nil
	}

	f.WriteString("{").WriteNewline()

	f.IncrementIndent()

	for _, field := range fieldList {
		err := f.FormatFieldDefinition(field)
		if err != nil {
			return err
		}
	}

	f.DecrementIndent()

	f.WriteString("}")

	return nil
}

func (f *formatter) FormatFieldDefinition(field *ast.FieldDefinition) error {
	if !f.emitBuiltin && strings.HasPrefix(field.Name, "__") {
		return nil
	}

	f.WriteDescription(field.Description)
	f.WriteWord(field.Name).NoPadding()
	if err := f.FormatArgumentDefinitionList(field.Arguments); err != nil {
		return err
	}
	f.NoPadding().WriteString(":").NeedPadding()

	if err := f.FormatType(field.Type); err != nil {
		return err
	}

	if field.DefaultValue != nil {
		f.WriteWord("=")
		if err := f.FormatValue(field.DefaultValue); err != nil {
			return err
		}
	}

	if err := f.FormatDirectiveList(field.Directives); err != nil {
		return err
	}

	f.WriteNewline()

	return nil
}

func (f *formatter) FormatArgumentDefinitionList(lists ast.ArgumentDefinitionList) error {
	if len(lists) == 0 {
		return nil
	}

	f.WriteString("(")
	for idx, arg := range lists {
		if err := f.FormatArgumentDefinition(arg); err != nil {
			return err
		}

		if idx != len(lists)-1 {
			f.NoPadding().WriteWord(",")
		}
	}
	f.NoPadding().WriteString(")").NeedPadding()

	return nil
}

func (f *formatter) FormatArgumentDefinition(def *ast.ArgumentDefinition) error {
	f.WriteDescription(def.Description)
	f.WriteWord(def.Name).NoPadding().WriteString(":").NeedPadding()
	if err := f.FormatType(def.Type); err != nil {
		return err
	}
	if def.DefaultValue != nil {
		f.WriteWord("=")
		if err := f.FormatValue(def.DefaultValue); err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatDirectiveLocation(location ast.DirectiveLocation) error {
	f.WriteWord(string(location))

	return nil
}

func (f *formatter) FormatDirectiveDefinitionList(lists ast.DirectiveDefinitionList) error {
	if len(lists) == 0 {
		return nil
	}

	for _, dec := range lists {
		if err := f.FormatDirectiveDefinition(dec); err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatDirectiveDefinition(def *ast.DirectiveDefinition) error {
	if !f.emitBuiltin {
		switch def.Name {
		case "deprecated", "skip", "include":
			return nil
		}
	}

	f.WriteWord("directive").WriteString("@").WriteWord(def.Name)

	if len(def.Arguments) != 0 {
		f.NoPadding()

		if err := f.FormatArgumentDefinitionList(def.Arguments); err != nil {
			return err
		}
	}

	if len(def.Locations) != 0 {
		f.WriteWord("on")

		for idx, dirLoc := range def.Locations {
			if err := f.FormatDirectiveLocation(dirLoc); err != nil {
				return err
			}

			if idx != len(def.Locations)-1 {
				f.WriteWord("|")
			}
		}
	}

	f.WriteNewline()

	return nil
}

func (f *formatter) FormatDefinitionList(lists ast.DefinitionList, extend bool) error {
	if len(lists) == 0 {
		return nil
	}

	for _, dec := range lists {
		if err := f.FormatDefinition(dec, extend); err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatDefinition(def *ast.Definition, extend bool) error {
	if !f.emitBuiltin && def.BuiltIn {
		return nil
	}

	f.WriteDescription(def.Description)

	if extend {
		f.WriteWord("extend")
	}

	switch def.Kind {
	case ast.Scalar:
		f.WriteWord("scalar").WriteWord(def.Name)

	case ast.Object:
		f.WriteWord("type").WriteWord(def.Name)

	case ast.Interface:
		f.WriteWord("interface").WriteWord(def.Name)

	case ast.Union:
		f.WriteWord("union").WriteWord(def.Name)

	case ast.Enum:
		f.WriteWord("enum").WriteWord(def.Name)

	case ast.InputObject:
		f.WriteWord("input").WriteWord(def.Name)
	}

	if err := f.FormatDirectiveList(def.Directives); err != nil {
		return err
	}

	if len(def.Types) != 0 {
		f.WriteWord("=").WriteWord(strings.Join(def.Types, " | "))
	}

	if len(def.Interfaces) != 0 {
		f.WriteWord("implements").WriteWord(strings.Join(def.Interfaces, ", "))
	}

	if err := f.FormatFieldList(def.Fields); err != nil {
		return err
	}

	if err := f.FormatEnumValueList(def.EnumValues); err != nil {
		return err
	}

	f.WriteNewline()

	return nil
}

func (f *formatter) FormatEnumValueList(lists ast.EnumValueList) error {
	if len(lists) == 0 {
		return nil
	}

	f.WriteString("{").WriteNewline()

	f.IncrementIndent()
	for _, v := range lists {
		if err := f.FormatEnumValueDefinition(v); err != nil {
			return err
		}
	}
	f.DecrementIndent()

	f.WriteString("}")

	return nil
}

func (f *formatter) FormatEnumValueDefinition(def *ast.EnumValueDefinition) error {
	f.WriteDescription(def.Description)
	f.WriteWord(def.Name)
	if err := f.FormatDirectiveList(def.Directives); err != nil {
		return err
	}

	f.WriteNewline()

	return nil
}

func (f *formatter) FormatOperationList(lists ast.OperationList) error {
	for _, def := range lists {
		if err := f.FormatOperationDefinition(def); err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatOperationDefinition(def *ast.OperationDefinition) error {
	f.WriteWord(string(def.Operation))
	if def.Name != "" {
		f.WriteWord(def.Name)
	}

	if err := f.FormatVariableDefinitionList(def.VariableDefinitions); err != nil {
		return err
	}

	if err := f.FormatDirectiveList(def.Directives); err != nil {
		return err
	}

	if len(def.SelectionSet) != 0 {
		if err := f.FormatSelectionSet(def.SelectionSet); err != nil {
			return err
		}
		f.WriteNewline()
	}

	return nil
}

func (f *formatter) FormatDirectiveList(lists ast.DirectiveList) error {
	if len(lists) == 0 {
		return nil
	}

	for _, dir := range lists {
		if err := f.FormatDirective(dir); err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatDirective(dir *ast.Directive) error {
	f.WriteString("@").WriteWord(dir.Name).NoPadding()
	if err := f.FormatArgumentList(dir.Arguments); err != nil {
		return err
	}

	return nil
}

func (f *formatter) FormatArgumentList(lists ast.ArgumentList) error {
	f.WriteString("(")
	for idx, arg := range lists {
		if err := f.FormatArgument(arg); err != nil {
			return err
		}

		if idx != len(lists)-1 {
			f.NoPadding().WriteWord(",")
		}
	}
	f.WriteString(")")

	return nil
}

func (f *formatter) FormatArgument(arg *ast.Argument) error {
	f.WriteWord(arg.Name).NoPadding().WriteString(":").NeedPadding()
	f.WriteString(arg.Value.String())

	return nil
}

func (f *formatter) FormatFragmentDefinitionList(lists ast.FragmentDefinitionList) error {

	for _, def := range lists {
		if err := f.FormatFragmentDefinition(def); err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatFragmentDefinition(def *ast.FragmentDefinition) error {

	f.WriteWord("fragment").WriteWord(def.Name)

	if err := f.FormatVariableDefinitionList(def.VariableDefinition); err != nil {
		return err
	}

	f.WriteWord("on").WriteWord(def.TypeCondition)

	if err := f.FormatDirectiveList(def.Directives); err != nil {
		return err
	}

	if len(def.SelectionSet) != 0 {
		if err := f.FormatSelectionSet(def.SelectionSet); err != nil {
			return err
		}
		f.WriteNewline()
	}

	return nil
}

func (f *formatter) FormatVariableDefinitionList(lists ast.VariableDefinitionList) error {
	if len(lists) == 0 {
		return nil
	}

	f.WriteString("(")
	for idx, def := range lists {
		if err := f.FormatVariableDefinition(def); err != nil {
			return err
		}

		if idx != len(lists)-1 {
			f.NoPadding().WriteWord(",")
		}
	}
	f.NoPadding().WriteString(")").NeedPadding()

	return nil
}

func (f *formatter) FormatVariableDefinition(def *ast.VariableDefinition) error {

	f.WriteString("$").WriteWord(def.Variable).NoPadding().WriteString(":").NeedPadding()
	if err := f.FormatType(def.Type); err != nil {
		return err
	}

	if def.DefaultValue != nil {
		f.WriteWord("=")
		if err := f.FormatValue(def.DefaultValue); err != nil {
			return err
		}
	}

	// TODO https://github.com/vektah/gqlparser/issues/102
	//   VariableDefinition : Variable : Type DefaultValue? Directives[Const]?

	return nil
}

func (f *formatter) FormatSelectionSet(sets ast.SelectionSet) error {
	if len(sets) == 0 {
		return nil
	}

	f.WriteString("{").WriteNewline()
	f.IncrementIndent()

	for _, sel := range sets {
		if err := f.FormatSelection(sel); err != nil {
			return err
		}
	}

	f.DecrementIndent()
	f.WriteString("}")

	return nil
}

func (f *formatter) FormatSelection(selection ast.Selection) error {

	switch v := selection.(type) {
	case *ast.Field:
		if err := f.FormatField(v); err != nil {
			return err
		}

	case *ast.FragmentSpread:
		if err := f.FormatFragmentSpread(v); err != nil {
			return err
		}

	case *ast.InlineFragment:
		if err := f.FormatInlineFragment(v); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown Selection type: %T", selection)
	}

	f.WriteNewline()

	return nil
}

func (f *formatter) FormatField(field *ast.Field) error {

	if field.Alias != "" && field.Alias != field.Name {
		f.WriteWord(field.Alias).NoPadding().WriteString(":").NeedPadding()
	}
	f.WriteWord(field.Name)

	if len(field.Arguments) != 0 {
		f.NoPadding()
		if err := f.FormatArgumentList(field.Arguments); err != nil {
			return err
		}
		f.NeedPadding()
	}

	if err := f.FormatDirectiveList(field.Directives); err != nil {
		return err
	}

	if err := f.FormatSelectionSet(field.SelectionSet); err != nil {
		return err
	}

	return nil
}

func (f *formatter) FormatFragmentSpread(spread *ast.FragmentSpread) error {

	f.WriteWord("...").WriteWord(spread.Name)

	if err := f.FormatDirectiveList(spread.Directives); err != nil {
		return err
	}

	return nil
}

func (f *formatter) FormatInlineFragment(inline *ast.InlineFragment) error {

	f.WriteWord("...")
	if inline.TypeCondition != "" {
		f.WriteWord("on").WriteWord(inline.TypeCondition)
	}

	if err := f.FormatDirectiveList(inline.Directives); err != nil {
		return err
	}

	if err := f.FormatSelectionSet(inline.SelectionSet); err != nil {
		return err
	}

	return nil
}

func (f *formatter) FormatType(t *ast.Type) error {
	f.WriteWord(t.String())
	return nil
}

func (f *formatter) FormatValue(value *ast.Value) error {
	f.WriteString(value.String())

	return nil
}
