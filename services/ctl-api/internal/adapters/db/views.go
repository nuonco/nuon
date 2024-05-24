package db

import (
	"gorm.io/gorm/schema"

	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	viewsSuffix string = "_view"
)

var viewModels []string = []string{
	"AppConfig",
	"ComponentConfigConnection",
}

var enableViews bool = true

func DisableViews() {
	enableViews = false
}

func EnableViews() {
	enableViews = false
}

// viewsNamer is a namer that supports the use of a view, when enabled.
// this allows us to create a `foo_views` view for a table foo, and then dynamically still migrate it, while also using
// it automatically during a non-migration session.
var _ schema.Namer = viewsNamer{}

type viewsNamer struct {
	schema.NamingStrategy
}

func (v viewsNamer) TableName(val string) string {
	// if this is a migration, we always use the main table
	if !enableViews {
		return v.NamingStrategy.TableName(val)
	}

	tableName := v.NamingStrategy.TableName(val)
	if generics.SliceContains[string](val, viewModels) {
		tableName += viewsSuffix
	}

	return tableName
}
