---
name: GORM Model Quality
description: Checks GORM struct definitions for data modeling best practices
severity-default: medium
tools: [Grep, Read, glob]
globs: ["services/ctl-api/internal/app/*.go"]
---

When reviewing GORM model structs in `services/ctl-api/internal/app/`, check for these patterns:

## Index Consistency

- **Prefer `Indexes()` method over inline `gorm:"index"` tags** for composite or non-trivial indexes
- Exception: `DeletedAt` field should have inline `gorm:"index"` for soft-delete filtering
- Look for mixed indexing strategies (some in tags, some in `Indexes()`) and suggest consolidation

## Foreign Key Associations

- **Add explicit FK tags for clarity**: `gorm:"foreignKey:FieldID;references:ID"`
- Common associations needing explicit tags: `CreatedBy Account`, `Org Org`
- Check that preloadable associations have proper tags

## Polymorphic Relationships

When a model uses `OwnerID`/`OwnerType` polymorphic pattern:
- Add DB check constraint limiting `OwnerType` to valid values
- Use `varchar(26)` with length check for `OwnerID` (matches standard ID format)
- Consider adding reverse associations on owner models: `gorm:"polymorphic:Owner;polymorphicValue:table_name"`

## JSONB Fields

- **Add `default:'[]'`** for slice/array JSONB fields to avoid null vs empty array issues
- Pattern: `gorm:"type:jsonb;serializer:json;default:'[]'"`

## Common Missing Elements

1. **`DeletedAt` index**: Should have `gorm:"index"` for soft-delete filtering performance
2. **Composite indexes for common queries**: Look at query patterns and suggest indexes like:
   - `(org_id, created_at)` for listing
   - `(org_id, owner_type, owner_id, evaluated_at)` for polymorphic latest lookups
3. **`BeforeCreate` hooks**: Should populate `ID`, `CreatedByID`, `OrgID` from context

## Reference Patterns

Good model structure example:
```go
type Example struct {
    ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26"`
    CreatedByID string                `gorm:"not null;default:null"`
    CreatedBy   Account               `gorm:"foreignKey:CreatedByID;references:ID" json:"-"`
    CreatedAt   time.Time             `gorm:"notnull"`
    UpdatedAt   time.Time             `gorm:"notnull"`
    DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`
    
    OrgID string `gorm:"notnull"`
    Org   Org    `gorm:"foreignKey:OrgID;references:ID" json:"-"`
    
    // JSONB with default
    Items []Item `gorm:"type:jsonb;serializer:json;default:'[]'"`
}

func (e *Example) Indexes(db *gorm.DB) []migrations.Index {
    return []migrations.Index{
        {Name: indexes.Name(db, &Example{}, "org_id"), Columns: []string{"org_id"}},
        // Add composite indexes for common query patterns
    }
}
```

Report findings with:
- File and line number
- Current pattern vs recommended pattern
- Reference to similar models that follow best practices
