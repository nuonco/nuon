package blobstore

import (
	"context"
	"reflect"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const pluginName = "blobstore"

var (
	blobType    = reflect.TypeOf(Blob{})
	blobPtrType = reflect.TypeOf((*Blob)(nil))
)

func NewPlugin(svc Service) gorm.Plugin {
	return &plugin{svc: svc}
}

type plugin struct {
	svc     Service
	hasBlob sync.Map // *schema.Schema -> bool
}

func (p *plugin) Name() string {
	return pluginName
}

func (p *plugin) Initialize(db *gorm.DB) error {
	if p.svc == nil {
		return nil
	}

	for _, register := range []func(string, func(*gorm.DB)) error{
		db.Callback().Create().Before("*").Register,
		db.Callback().Query().Before("*").Register,
		db.Callback().Update().Before("*").Register,
		db.Callback().Delete().Before("*").Register,
		db.Callback().Raw().Before("*").Register,
		db.Callback().Row().Before("*").Register,
	} {
		if err := register("blobstore:inject", p.inject); err != nil {
			return err
		}
	}

	for _, register := range []func(string, func(*gorm.DB)) error{
		db.Callback().Create().After("*").Register,
		db.Callback().Query().After("*").Register,
		db.Callback().Update().After("*").Register,
	} {
		if err := register("blobstore:bind", p.bindResults); err != nil {
			return err
		}
	}

	return nil
}

func (p *plugin) inject(tx *gorm.DB) {
	if tx == nil || tx.DryRun || tx.Statement == nil || p.svc == nil {
		return
	}
	ctx := tx.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if GetBlobService(ctx) != nil {
		return
	}
	tx.Statement.Context = WithBlobService(ctx, p.svc)
}

func (p *plugin) bindResults(tx *gorm.DB) {
	if p.svc == nil || tx == nil || tx.DryRun || tx.Error != nil || tx.Statement == nil {
		return
	}
	if !p.schemaHasBlob(tx.Statement.Schema) {
		return
	}
	v := tx.Statement.ReflectValue
	if !v.IsValid() && tx.Statement.Dest != nil {
		v = reflect.ValueOf(tx.Statement.Dest)
	}
	p.bindValue(v, map[uintptr]struct{}{})
}

func (p *plugin) schemaHasBlob(s *schema.Schema) bool {
	if s == nil {
		return false
	}
	if v, ok := p.hasBlob.Load(s); ok {
		return v.(bool)
	}
	has := false
	for _, field := range s.Fields {
		t := field.FieldType
		for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			t = t.Elem()
		}
		if t == blobType {
			has = true
			break
		}
	}
	p.hasBlob.Store(s, has)
	return has
}

func (p *plugin) bindValue(v reflect.Value, seen map[uintptr]struct{}) {
	if !v.IsValid() {
		return
	}

	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		ptr := v.Pointer()
		if _, ok := seen[ptr]; ok {
			return
		}
		seen[ptr] = struct{}{}
		if v.Type() == blobPtrType {
			if blob, ok := v.Interface().(*Blob); ok {
				blob.svc = p.svc
			}
			return
		}
		p.bindValue(v.Elem(), seen)
	case reflect.Struct:
		if v.Type() == blobType && v.CanAddr() {
			v.Addr().Interface().(*Blob).svc = p.svc
			return
		}
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).PkgPath != "" {
				continue
			}
			p.bindValue(v.Field(i), seen)
		}
	case reflect.Slice, reflect.Array:
		elem := v.Type().Elem().Kind()
		switch elem {
		case reflect.Struct, reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Array:
		default:
			return
		}
		for i := 0; i < v.Len(); i++ {
			p.bindValue(v.Index(i), seen)
		}
	}
}

func serviceFromDB(tx *gorm.DB) Service {
	if tx == nil || tx.Config == nil || tx.Config.Plugins == nil {
		return nil
	}
	p, ok := tx.Config.Plugins[pluginName]
	if !ok {
		return nil
	}
	bp, ok := p.(*plugin)
	if !ok || bp.svc == nil {
		return nil
	}
	return bp.svc
}
