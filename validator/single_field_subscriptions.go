package validator

import (
	"strconv"

	"github.com/vektah/gqlparser"
)

func init() {
	operationVisitor = append(operationVisitor, singleFieldSubscriptions)
}

func singleFieldSubscriptions(ctx *vctx, operation *gqlparser.OperationDefinition) {
	if operation.Operation != gqlparser.Subscription {
		return
	}

	if len(operation.SelectionSet) != 1 {
		name := "Anonymous Subscription"
		if operation.Name != "" {
			name = `Subscription ` + strconv.Quote(operation.Name)
		}

		ctx.errors = append(ctx.errors, Error(
			Rule("SingleFieldSubscriptions"),
			Message(`%s must select only one top level field.`, name),
		))
	}
}
