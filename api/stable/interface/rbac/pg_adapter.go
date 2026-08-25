package rbac

import (
	"fmt"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/jmoiron/sqlx"
)

// pgAdapter is a Casbin adapter backed by PostgreSQL via sqlx.
// It stores policies in the casbin_rule table created by migration 00013.
type pgAdapter struct {
	db *sqlx.DB
}

type casbinRule struct {
	Ptype string `db:"ptype"`
	V0    string `db:"v0"`
	V1    string `db:"v1"`
	V2    string `db:"v2"`
	V3    string `db:"v3"`
	V4    string `db:"v4"`
	V5    string `db:"v5"`
}

func newPGAdapter(db *sqlx.DB) persist.BatchAdapter {
	return &pgAdapter{db: db}
}

func (a *pgAdapter) LoadPolicy(m model.Model) error {
	rows := []casbinRule{}
	if err := a.db.Select(&rows, `SELECT ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule`); err != nil {
		return fmt.Errorf("casbin load policy: %w", err)
	}
	for _, row := range rows {
		persist.LoadPolicyLine(row.toLine(), m)
	}
	return nil
}

func (a *pgAdapter) SavePolicy(m model.Model) error {
	tx, err := a.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM casbin_rule`); err != nil {
		return fmt.Errorf("casbin save policy delete: %w", err)
	}

	var rules []casbinRule
	for ptype, assertions := range m["p"] {
		for _, rule := range assertions.Policy {
			rules = append(rules, newRule(ptype, rule))
		}
	}
	for ptype, assertions := range m["g"] {
		for _, rule := range assertions.Policy {
			rules = append(rules, newRule(ptype, rule))
		}
	}

	for _, r := range rules {
		if _, err := tx.NamedExec(
			`INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES (:ptype, :v0, :v1, :v2, :v3, :v4, :v5)`,
			r,
		); err != nil {
			return fmt.Errorf("casbin save policy insert: %w", err)
		}
	}
	return tx.Commit()
}

func (a *pgAdapter) AddPolicy(sec, ptype string, rule []string) error {
	r := newRule(ptype, rule)
	_, err := a.db.NamedExec(
		`INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES (:ptype, :v0, :v1, :v2, :v3, :v4, :v5)
		 ON CONFLICT DO NOTHING`,
		r,
	)
	return err
}

func (a *pgAdapter) RemovePolicy(sec, ptype string, rule []string) error {
	r := newRule(ptype, rule)
	_, err := a.db.Exec(
		`DELETE FROM casbin_rule WHERE ptype=$1 AND v0=$2 AND v1=$3 AND v2=$4`,
		r.Ptype, r.V0, r.V1, r.V2,
	)
	return err
}

func (a *pgAdapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	query := `DELETE FROM casbin_rule WHERE ptype=$1`
	args := []interface{}{ptype}
	cols := []string{"v0", "v1", "v2", "v3", "v4", "v5"}
	for i, v := range fieldValues {
		if v != "" {
			query += fmt.Sprintf(" AND %s=$%d", cols[fieldIndex+i], len(args)+1)
			args = append(args, v)
		}
	}
	_, err := a.db.Exec(query, args...)
	return err
}

func (a *pgAdapter) AddPolicies(sec, ptype string, rules [][]string) error {
	for _, rule := range rules {
		if err := a.AddPolicy(sec, ptype, rule); err != nil {
			return err
		}
	}
	return nil
}

func (a *pgAdapter) RemovePolicies(sec, ptype string, rules [][]string) error {
	for _, rule := range rules {
		if err := a.RemovePolicy(sec, ptype, rule); err != nil {
			return err
		}
	}
	return nil
}

func newRule(ptype string, rule []string) casbinRule {
	r := casbinRule{Ptype: ptype}
	if len(rule) > 0 {
		r.V0 = rule[0]
	}
	if len(rule) > 1 {
		r.V1 = rule[1]
	}
	if len(rule) > 2 {
		r.V2 = rule[2]
	}
	if len(rule) > 3 {
		r.V3 = rule[3]
	}
	if len(rule) > 4 {
		r.V4 = rule[4]
	}
	if len(rule) > 5 {
		r.V5 = rule[5]
	}
	return r
}

func (r casbinRule) toLine() string {
	line := r.Ptype
	for _, v := range []string{r.V0, r.V1, r.V2, r.V3, r.V4, r.V5} {
		if v == "" {
			break
		}
		line += ", " + v
	}
	return line
}
