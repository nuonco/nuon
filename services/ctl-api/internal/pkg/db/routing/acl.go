package routing

import (
	"fmt"

	"gorm.io/gorm"
)

type viewModel interface {
	UseView() bool
	ViewVersion() string
}

type TableACL struct {
	Allow map[string]struct{}
	Deny  map[string]struct{}
}

func NewTableACL(allow, deny []string) *TableACL {
	acl := &TableACL{
		Allow: make(map[string]struct{}, len(allow)),
		Deny:  make(map[string]struct{}, len(deny)),
	}
	for _, t := range allow {
		acl.Allow[t] = struct{}{}
	}
	for _, t := range deny {
		acl.Deny[t] = struct{}{}
	}
	return acl
}

func (a *TableACL) AllowsReplica(table string) bool {
	if a == nil {
		return true
	}
	if table == "" {
		return false
	}
	if _, denied := a.Deny[table]; denied {
		return false
	}
	_, allowed := a.Allow[table]
	return allowed
}

type ACLBuilder struct {
	db    *gorm.DB
	allow map[string]struct{}
	deny  map[string]struct{}
}

func NewACLBuilder(db *gorm.DB) *ACLBuilder {
	return &ACLBuilder{
		db:    db,
		allow: make(map[string]struct{}),
		deny:  make(map[string]struct{}),
	}
}

func (b *ACLBuilder) Allow(models ...interface{}) *ACLBuilder {
	for _, m := range models {
		for _, t := range tableNamesFor(b.db, m) {
			b.allow[t] = struct{}{}
		}
	}
	return b
}

func (b *ACLBuilder) Deny(models ...interface{}) *ACLBuilder {
	for _, m := range models {
		for _, t := range tableNamesFor(b.db, m) {
			b.deny[t] = struct{}{}
		}
	}
	return b
}

func (b *ACLBuilder) Build() *TableACL {
	acl := &TableACL{
		Allow: make(map[string]struct{}, len(b.allow)),
		Deny:  make(map[string]struct{}, len(b.deny)),
	}
	for t := range b.allow {
		acl.Allow[t] = struct{}{}
	}
	for t := range b.deny {
		acl.Deny[t] = struct{}{}
	}
	return acl
}

func tableNamesFor(db *gorm.DB, model interface{}) []string {
	base := ResolveTable(db, model)
	if base == "" {
		return nil
	}
	out := []string{base}
	if vm, ok := model.(viewModel); ok && vm.UseView() {
		out = append(out, fmt.Sprintf("%s_view_%s", base, vm.ViewVersion()))
	}
	return out
}

func ResolveTable(db *gorm.DB, model interface{}) string {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		if t, ok := model.(interface{ TableName() string }); ok {
			return t.TableName()
		}
		return ""
	}
	if stmt.Schema == nil {
		return ""
	}
	return stmt.Schema.Table
}
