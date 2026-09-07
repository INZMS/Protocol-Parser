// Package datascope carries the server-resolved row scope from the HTTP layer
// into repositories.  The browser never supplies this value.
package datascope

import "context"

type Scope struct {
	All             bool
	OrganizationIDs []int64
}

type contextKey struct{}

func With(ctx context.Context, value Scope) context.Context { return context.WithValue(ctx, contextKey{}, value) }

func From(ctx context.Context) Scope {
	value, _ := ctx.Value(contextKey{}).(Scope)
	return value
}

// Clause returns a parameterised IN restriction for a table's organisation
// column.  A non-admin without an assigned organisation deliberately receives
// 1=0 rather than an implicit global scope.
func Clause(ctx context.Context, column string) (string, []any) {
	scope := From(ctx)
	if scope.All {
		return "", nil
	}
	if len(scope.OrganizationIDs) == 0 {
		return "1=0", nil
	}
	placeholders := "?"
	args := make([]any, 0, len(scope.OrganizationIDs))
	args = append(args, scope.OrganizationIDs[0])
	for _, id := range scope.OrganizationIDs[1:] {
		placeholders += ",?"
		args = append(args, id)
	}
	return column + " IN (" + placeholders + ")", args
}

func Allows(ctx context.Context, organizationID int64) bool {
	scope := From(ctx)
	if scope.All {
		return true
	}
	for _, id := range scope.OrganizationIDs {
		if id == organizationID {
			return true
		}
	}
	return false
}
