package paging

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
)

// Query is the normalized list-query contract shared by repositories.
// String filters deliberately preserve "not supplied" separately from zero.
type Query struct {
	Page              int
	PageSize          int
	Keyword           string
	Status            string
	OrganizationID    string
	Category          string
	VehicleType       string
	InventoryStatus   string
	Protocol          string
	MaintenanceStatus string
	MaintenanceType   string
	BindingStatus     string
	DeviceNo          string
	VehicleID         string
	RecordType        string
	RoleID            string
	MenuType          string
	CompanyID         string
	SortField         string
	SortOrder         string
}

func (q Query) LimitOffset() (int, int) {
	page, size := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 200 {
		size = 200
	}
	return size, (page - 1) * size
}

// OrderBy accepts only explicitly mapped client fields, preventing SQL injection.
func (q Query) OrderBy(allowed map[string]string, fallback string) string {
	column, ok := allowed[q.SortField]
	if !ok {
		return fallback
	}
	direction := "ASC"
	if strings.EqualFold(q.SortOrder, "desc") || q.SortOrder == "descend" {
		direction = "DESC"
	}
	return column + " " + direction
}

func Int64(value string) (int64, bool) {
	v, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return v, err == nil && v > 0
}

type contextKey struct{}

type State struct {
	Query Query
	Total int
}

func WithContext(ctx context.Context, query Query) (context.Context, *State) {
	state := &State{Query: query}
	return context.WithValue(ctx, contextKey{}, state), state
}

func FromContext(ctx context.Context) *State {
	state, _ := ctx.Value(contextKey{}).(*State)
	if state == nil {
		return &State{Query: Query{Page: 1, PageSize: 10}}
	}
	return state
}

// QueryRows runs a matching count query and a limited data query. Callers
// supply only fixed SQL fragments and bound parameters; orderBy must come from
// Query.OrderBy's whitelist mapping.
func QueryRows(ctx context.Context, db *sql.DB, selectSQL, countSQL string, clauses []string, args []any, orderBy string) (*sql.Rows, error) {
	state := FromContext(ctx)
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	if err := db.QueryRowContext(ctx, countSQL+where, args...).Scan(&state.Total); err != nil {
		return nil, err
	}
	limit, offset := state.Query.LimitOffset()
	pageArgs := append(append([]any{}, args...), limit, offset)
	return db.QueryContext(ctx, selectSQL+where+" ORDER BY "+orderBy+" LIMIT ? OFFSET ?", pageArgs...)
}
