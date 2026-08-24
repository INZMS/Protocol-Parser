package history

import (
	"context"
	"errors"
	"time"

	"protocol-parser-server/parser/core"
)

var ErrNotFound = errors.New("解析记录不存在")

type Query struct {
	Keyword  string
	Page     int
	PageSize int
}

type Summary struct {
	ID          int64     `json:"id"`
	Protocol    string    `json:"protocol"`
	MessageID   string    `json:"messageId"`
	MessageName string    `json:"messageName"`
	Length      int       `json:"length"`
	CreatedAt   time.Time `json:"-"`
	Time        string    `json:"time"`
}

type Record struct {
	Summary
	Result *core.ParseResult `json:"result"`
}

type Page struct {
	Items    []Summary `json:"items"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
}

type Store interface {
	Create(context.Context, *core.ParseResult) (int64, error)
	List(context.Context, Query) (*Page, error)
	Get(context.Context, int64) (*Record, error)
	Delete(context.Context, int64) error
	Clear(context.Context) error
}
