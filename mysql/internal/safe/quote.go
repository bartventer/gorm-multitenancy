// Package safe provides utilities for safely quoting SQL identifiers and strings.
package safe

import (
	"strings"

	"gorm.io/gorm/clause"
)

type ToQuoter interface {
	QuoteTo(clause.Writer, string)
}

func QuoteRawSQLForTenant(q ToQuoter, raw, tenantID string) string {
	sql := new(strings.Builder)
	_, _ = sql.WriteString(raw)
	q.QuoteTo(sql, tenantID)
	return sql.String()
}
