package casbin

import (
	"context"
	"database/sql"
	"fmt"

	"braces.dev/errtrace"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/jmoiron/sqlx"
)

// CasbinRule represents a row in casbin_rule table
type CasbinRule struct {
	ID    int            `db:"id"`
	PType string         `db:"ptype"`
	V0    sql.NullString `db:"v0"`
	V1    sql.NullString `db:"v1"`
	V2    sql.NullString `db:"v2"`
	V3    sql.NullString `db:"v3"`
	V4    sql.NullString `db:"v4"`
	V5    sql.NullString `db:"v5"`
}

// SqlxAdapter is a Casbin adapter for PostgreSQL using sqlx
type SqlxAdapter struct {
	db *sqlx.DB
}

// NewSqlxAdapter creates a new sqlx adapter for Casbin
func NewSqlxAdapter(db *sqlx.DB) (*SqlxAdapter, error) {
	adapter := &SqlxAdapter{db: db}
	return adapter, nil
}

// LoadPolicy loads all policy rules from the database
func (a *SqlxAdapter) LoadPolicy(model model.Model) error {
	var rules []CasbinRule
	query := "SELECT id, ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule"

	err := a.db.Select(&rules, query)
	if err != nil {
		return errtrace.Wrap(err)
	}

	for _, rule := range rules {
		a.loadPolicyLine(rule, model)
	}

	return nil
}

// SavePolicy saves all policy rules to the database
func (a *SqlxAdapter) SavePolicy(model model.Model) error {
	ctx := context.Background()

	// Start transaction
	tx, err := a.db.BeginTxx(ctx, nil)
	if err != nil {
		return errtrace.Wrap(err)
	}
	defer tx.Rollback()

	// Clear existing rules
	_, err = tx.Exec("DELETE FROM casbin_rule")
	if err != nil {
		return errtrace.Wrap(err)
	}

	// Save all policy rules
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			if err := a.savePolicyLine(tx, ptype, rule); err != nil {
				return errtrace.Wrap(err)
			}
		}
	}

	// Save all role rules
	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			if err := a.savePolicyLine(tx, ptype, rule); err != nil {
				return errtrace.Wrap(err)
			}
		}
	}

	return errtrace.Wrap(tx.Commit())
}

// AddPolicy adds a policy rule to the database
func (a *SqlxAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	ctx := context.Background()
	tx, err := a.db.BeginTxx(ctx, nil)
	if err != nil {
		return errtrace.Wrap(err)
	}
	defer tx.Rollback()

	if err := a.savePolicyLine(tx, ptype, rule); err != nil {
		return errtrace.Wrap(err)
	}

	return errtrace.Wrap(tx.Commit())
}

// RemovePolicy removes a policy rule from the database
func (a *SqlxAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	ctx := context.Background()
	query := a.buildDeleteQuery(ptype, rule)

	_, err := a.db.ExecContext(ctx, query)
	return errtrace.Wrap(err)
}

// RemoveFilteredPolicy removes policy rules that match the filter from the database
func (a *SqlxAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	ctx := context.Background()
	query := a.buildFilteredDeleteQuery(ptype, fieldIndex, fieldValues...)

	_, err := a.db.ExecContext(ctx, query)
	return errtrace.Wrap(err)
}

// Helper methods

func (a *SqlxAdapter) loadPolicyLine(rule CasbinRule, model model.Model) {
	lineText := rule.PType
	if rule.V0.Valid && rule.V0.String != "" {
		lineText += ", " + rule.V0.String
	}
	if rule.V1.Valid && rule.V1.String != "" {
		lineText += ", " + rule.V1.String
	}
	if rule.V2.Valid && rule.V2.String != "" {
		lineText += ", " + rule.V2.String
	}
	if rule.V3.Valid && rule.V3.String != "" {
		lineText += ", " + rule.V3.String
	}
	if rule.V4.Valid && rule.V4.String != "" {
		lineText += ", " + rule.V4.String
	}
	if rule.V5.Valid && rule.V5.String != "" {
		lineText += ", " + rule.V5.String
	}

	persist.LoadPolicyLine(lineText, model)
}

func (a *SqlxAdapter) savePolicyLine(tx *sqlx.Tx, ptype string, rule []string) error {
	query := `INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	v0, v1, v2, v3, v4, v5 := "", "", "", "", "", ""
	if len(rule) > 0 {
		v0 = rule[0]
	}
	if len(rule) > 1 {
		v1 = rule[1]
	}
	if len(rule) > 2 {
		v2 = rule[2]
	}
	if len(rule) > 3 {
		v3 = rule[3]
	}
	if len(rule) > 4 {
		v4 = rule[4]
	}
	if len(rule) > 5 {
		v5 = rule[5]
	}

	_, err := tx.Exec(query, ptype, v0, v1, v2, v3, v4, v5)
	return errtrace.Wrap(err)
}

func (a *SqlxAdapter) buildDeleteQuery(ptype string, rule []string) string {
	query := fmt.Sprintf("DELETE FROM casbin_rule WHERE ptype = '%s'", ptype)

	for i, v := range rule {
		if v != "" {
			query += fmt.Sprintf(" AND v%d = '%s'", i, v)
		}
	}

	return query
}

func (a *SqlxAdapter) buildFilteredDeleteQuery(ptype string, fieldIndex int, fieldValues ...string) string {
	query := fmt.Sprintf("DELETE FROM casbin_rule WHERE ptype = '%s'", ptype)

	for i, v := range fieldValues {
		if v != "" {
			query += fmt.Sprintf(" AND v%d = '%s'", fieldIndex+i, v)
		}
	}

	return query
}

// IsFiltered returns true if the adapter supports filtered policies
func (a *SqlxAdapter) IsFiltered() bool {
	return false
}

// LoadFilteredPolicy loads filtered policy rules from the database
func (a *SqlxAdapter) LoadFilteredPolicy(model model.Model, filter interface{}) error {
	return a.LoadPolicy(model)
}

// Ensure SqlxAdapter implements persist.Adapter interface
var _ persist.Adapter = (*SqlxAdapter)(nil)
