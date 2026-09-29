package clickhouse

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

var regFullDataType = regexp.MustCompile(`\D*(\d+)\D?`)

var regColumnDefinition = regexp.MustCompile(`\s*(\w+)\s+[a-zA-Z0-9_]+(?:\((?:[^()]+|\((?:[^()]+|\([^()]*\))*\))*\))?\s*`)

var (
	ErrRenameColumnUnsupported = errors.New("renaming column is not supported in your clickhouse version < 20.4")
	ErrRenameIndexUnsupported  = errors.New("renaming index is not supported")
	ErrCreateIndexFailed       = errors.New("failed to create index with name")
)

type ClickhouseNestedMigrator interface {
	MigrateNestedColumn(value interface{}, field *schema.Field, childFields []*schema.Field) error
}

type Migrator struct {
	ClickhouseNestedMigrator
	migrator.Migrator
	Dialector
}

func (m Migrator) CurrentDatabase() (name string) {
	m.DB.Raw("SELECT currentDatabase()").Row().Scan(&name)
	return
}

func (m Migrator) FullDataTypeOf(field *schema.Field) (expr clause.Expr) {
	expr.SQL = m.Migrator.DataTypeOf(field)

	// why: ClickHouse does not support NULL or UNIQUE keywords, so those field flags are not emitted.
	// NULL and UNIQUE keyword is not supported in clickhouse.
	// Hence, skipping checks for field.Unique and field.NotNull

	if field.HasDefaultValue && (field.DefaultValueInterface != nil || field.DefaultValue != "") {
		if field.DefaultValueInterface != nil {
			defaultStmt := &gorm.Statement{Vars: []interface{}{field.DefaultValueInterface}}
			m.BindVarTo(defaultStmt, defaultStmt, field.DefaultValueInterface)
			expr.SQL += " DEFAULT " + m.Explain(defaultStmt.SQL.String(), field.DefaultValueInterface)
		} else if field.DefaultValue != "(-)" {
			expr.SQL += " DEFAULT " + field.DefaultValue
		}
	}

	if comment, ok := field.TagSettings["COMMENT"]; ok {
		expr.SQL += " COMMENT " + m.Explain("?", comment)
	}

	if ttl, ok := field.TagSettings["TTL"]; ok && ttl != "" {
		expr.SQL += " TTL " + ttl
	}

	// why: Build CODEC compression algorithm optionally
	// NOTE: the codec algo name is case sensitive!
	if codecstr, ok := field.TagSettings["CODEC"]; ok && codecstr != "" {
		codecSlice := strings.Split(codecstr, ",")
		codecArgsSQL := m.DefaultCompression
		if len(codecSlice) > 0 {
			codecArgsSQL = strings.Join(codecSlice, ",")
		}
		codecSQL := fmt.Sprintf(" CODEC(%s) ", codecArgsSQL)
		expr.SQL += codecSQL
	}

	return expr
}

func (m Migrator) CreateTable(models ...interface{}) error {
	for _, model := range m.ReorderModels(models, false) {
		tx := m.DB.Session(new(gorm.Session))
		if err := m.RunWithValue(model, func(stmt *gorm.Statement) (err error) {
			var (
				createTableSQL = "CREATE TABLE ?%s(%s %s %s) %s"
				args           = []interface{}{clause.Table{Name: stmt.Table}}
			)

			columnSlice := make([]string, 0, len(stmt.Schema.DBNames))
			for _, dbName := range stmt.Schema.DBNames {
				field := stmt.Schema.FieldsByDBName[dbName]
				if field.IgnoreMigration {
					continue
				}

				columnSlice = append(columnSlice, "? ?")
				args = append(args,
					clause.Column{Name: dbName},
					m.FullDataTypeOf(field),
				)
			}
			columnStr := strings.Join(columnSlice, ",")

			constrSlice := make([]string, 0, len(columnSlice))
			for _, check := range stmt.Schema.ParseCheckConstraints() {
				constrSlice = append(constrSlice, "CONSTRAINT ? CHECK ?")
				args = append(args,
					clause.Column{Name: check.Name},
					clause.Expr{SQL: check.Constraint},
				)
			}
			constrStr := strings.Join(constrSlice, ",")
			if len(constrSlice) > 0 {
				constrStr = ", " + constrStr
			}

			// why: Step 3. Build index SQL string
			// NOTE: clickhouse does not support for index class.
			indexSlice := make([]string, 0, 10)
			for _, index := range stmt.Schema.ParseIndexes() {
				if m.CreateIndexAfterCreateTable {
					defer func(model interface{}, indexName string) {
						// TODO (iqdf): what if there are multiple errors
						// when creating indices after create table?
						err = tx.Migrator().CreateIndex(model, indexName)
					}(model, index.Name)
					continue
				}
				// TODO(iqdf): support primary key by put it as pass the fieldname
				// as MergeTree(...) parameters. But somehow it complained.
				// Note that primary key doesn't ensure uniqueness

				indexType := m.DefaultIndexType
				if index.Type != "" {
					indexType = index.Type
				}

				buildIndexOptions := tx.Migrator().(migrator.BuildIndexOptionsInterface)
				indexOptions := buildIndexOptions.BuildIndexOptions(index.Fields, stmt)

				str := fmt.Sprintf("INDEX ? ? TYPE %s GRANULARITY %d", indexType, m.getIndexGranularityOption(index.Fields))
				indexSlice = append(indexSlice, str)
				args = append(args, clause.Expr{SQL: index.Name}, indexOptions)
			}
			indexStr := strings.Join(indexSlice, ", ")
			if len(indexSlice) > 0 {
				indexStr = ", " + indexStr
			}

			engineOpts := m.DefaultTableEngineOpts
			if tableOption, ok := m.DB.Get("gorm:table_options"); ok {
				engineOpts = fmt.Sprint(tableOption)
			}

			clusterOpts := ""
			if clusterOption, ok := m.DB.Get("gorm:table_cluster_options"); ok {
				clusterOpts = " " + fmt.Sprint(clusterOption) + " "
			}

			createTableSQL = fmt.Sprintf(createTableSQL, clusterOpts, columnStr, constrStr, indexStr, engineOpts)

			err = tx.Exec(createTableSQL, args...).Error

			return
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m Migrator) HasTable(value interface{}) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		currentDatabase := m.DB.Migrator().CurrentDatabase()
		return m.DB.Raw(
			"SELECT count(*) FROM system.tables WHERE database = ? AND name = ? AND is_temporary = ?",
			currentDatabase,
			stmt.Table,
			uint8(0)).Row().Scan(&count)
	})
	return count > 0
}

func (m Migrator) GetTables() (tableList []string, err error) {
	err = m.DB.Raw("SELECT TABLE_NAME FROM information_schema.tables where table_schema=? and table_type =1", m.CurrentDatabase()).Scan(&tableList).Error
	return
}

func (m Migrator) AddColumn(value interface{}, field string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if field := stmt.Schema.LookUpField(field); field != nil {
			if field.IgnoreMigration {
				return nil
			}

			clusterOpts := ""
			if clusterOption, ok := m.DB.Get("gorm:table_cluster_options"); ok {
				clusterOpts = " " + fmt.Sprint(clusterOption) + " "
			}
			sQL := fmt.Sprintf("ALTER TABLE ? %s ADD COLUMN ? ?", clusterOpts)
			return m.DB.Exec(
				sQL,
				clause.Table{Name: stmt.Table},
				clause.Column{Name: field.DBName},
				m.FullDataTypeOf(field),
			).Error
		}
		return fmt.Errorf("failed to look up field with name: %s", field)
	})
}

func (m Migrator) DropColumn(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if field := stmt.Schema.LookUpField(name); field != nil {
			name = field.DBName
		}
		clusterOpts := ""
		if clusterOption, ok := m.DB.Get("gorm:table_cluster_options"); ok {
			clusterOpts = " " + fmt.Sprint(clusterOption) + " "
		}
		sQL := fmt.Sprintf("ALTER TABLE ? %s DROP COLUMN ?", clusterOpts)
		return m.DB.Exec(
			sQL,
			clause.Table{Name: stmt.Table}, clause.Column{Name: name},
		).Error
	})
}

func (m Migrator) AlterColumn(value interface{}, field string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if field := stmt.Schema.LookUpField(field); field != nil {
			if field.IgnoreMigration {
				return nil
			}

			clusterOpts := ""
			if clusterOption, ok := m.DB.Get("gorm:table_cluster_options"); ok {
				clusterOpts = " " + fmt.Sprint(clusterOption) + " "
			}
			sQL := fmt.Sprintf("ALTER TABLE ? %s MODIFY COLUMN ? ?", clusterOpts)
			return m.DB.Exec(
				sQL,
				clause.Table{Name: stmt.Table},
				clause.Column{Name: field.DBName},
				m.FullDataTypeOf(field),
			).Error
		}
		return fmt.Errorf("altercolumn() failed to look up column with name: %s", field)
	})
}

// why: Only supported after ClickHouse 20.4 and above.
// See: https://github.com/ClickHouse/ClickHouse/issues/146
func (m Migrator) RenameColumn(value interface{}, oldName, newName string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if !m.DontSupportRenameColumn {
			var field *schema.Field
			if f := stmt.Schema.LookUpField(oldName); f != nil {
				oldName = f.DBName
				field = f
			}
			if f := stmt.Schema.LookUpField(newName); f != nil {
				newName = f.DBName
				field = f
			}
			if field != nil {
				clusterOpts := ""
				if clusterOption, ok := m.DB.Get("gorm:table_cluster_options"); ok {
					clusterOpts = " " + fmt.Sprint(clusterOption) + " "
				}
				sQL := fmt.Sprintf("ALTER TABLE ? %s RENAME COLUMN ? TO ?", clusterOpts)
				return m.DB.Exec(
					sQL,
					clause.Table{Name: stmt.Table},
					clause.Column{Name: oldName},
					clause.Column{Name: newName},
				).Error
			}
			return fmt.Errorf("renamecolumn() failed to look up column with name: %s", oldName)
		}
		return ErrRenameIndexUnsupported
	})
}

func (m Migrator) HasColumn(value interface{}, field string) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		currentDatabase := m.DB.Migrator().CurrentDatabase()
		name := field

		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				name = field.DBName
			}
		}

		return m.DB.Raw(
			"SELECT count(*) FROM system.columns WHERE database = ? AND table = ? AND name = ?",
			currentDatabase, stmt.Table, name,
		).Row().Scan(&count)
	})

	return count > 0
}

func (m Migrator) AutoMigrate(values ...interface{}) error {
	for _, value := range m.ReorderModels(values, true) {
		queryTx, execTx := m.GetQueryAndExecTx()
		if !queryTx.Migrator().HasTable(value) {
			if err := execTx.Migrator().CreateTable(value); err != nil {
				return err
			}
		} else {
			if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
				if stmt.Schema == nil {
					return errors.New("failed to get schema")
				}

				log.Printf("[Migrator.AutoMigrate] Running Migration for table \"%s\"", stmt.Table)
				columnTypes, err := queryTx.Migrator().ColumnTypes(value)
				if err != nil {
					return err
				}
				var (
					parseIndexes          = stmt.Schema.ParseIndexes()
					parseCheckConstraints = stmt.Schema.ParseCheckConstraints()
				)

				nestedFieldColSearchTerms := []string{}
				nestedFieldDBNames := []string{}
				nestedFields := []*schema.Field{}
				for _, field := range stmt.Schema.Fields {
					if strings.HasPrefix(string(field.DataType), "Nested") {
						nestedFields = append(nestedFields, field)
						nestedFieldDBNames = append(nestedFieldDBNames, field.DBName)
						nestedFieldColSearchTerms = append(nestedFieldColSearchTerms, fmt.Sprintf("%s.*", field.DBName))
					}
				}
				if len(nestedFields) > 0 {
					log.Printf("[Migrator.AutoMigrate] Found %d Nested Fields: \"%s\"", len(nestedFields), strings.Join(nestedFieldDBNames, ","))
					log.Printf("[Migrator.AutoMigrate] NestedField Facts:%+v", nestedFieldDBNames)
					log.Printf("[Migrator.AutoMigrate] NestedField Facts:%+v", nestedFieldColSearchTerms)
				}

				dbNamesOriginal := stmt.Schema.DBNames

				dbNamesFiltered := []string{}
				for _, dbName := range dbNamesOriginal {
					if !contains(nestedFieldDBNames, dbName) {
						dbNamesFiltered = append(dbNamesFiltered, dbName)
					}
				}

				for _, dbName := range dbNamesFiltered {
					var foundColumn gorm.ColumnType
					field := stmt.Schema.FieldsByDBName[dbName]
					if field.IgnoreMigration {
						continue
					}

					for _, columnType := range columnTypes {
						if columnType.Name() == dbName {
							foundColumn = columnType
							break
						}
					}

					if foundColumn == nil {
						if err = execTx.Migrator().AddColumn(value, dbName); err != nil {
							return err
						}
					} else {
						if err = execTx.Migrator().MigrateColumn(value, field, foundColumn); err != nil {
							return err
						}
					}
				}

				for _, nestedField := range nestedFields {
					dbName := nestedField.DBName
					childColumns := []gorm.ColumnType{}

					for _, colType := range columnTypes {
						fieldDBName := colType.Name()
						if !strings.Contains(fieldDBName, ".") {
						} else {
							parts := strings.Split(fieldDBName, ".")
							if parts[0] == dbName {
								childColumns = append(childColumns, colType)
							}
						}
					}

					if len(childColumns) == 0 {
						if err = execTx.Migrator().AddColumn(value, dbName); err != nil {
							return err
						}
					} else {
						field := stmt.Schema.FieldsByDBName[dbName]
						err = execTx.Migrator().(Migrator).MigrateNestedColumn(value, field, childColumns)
						if err != nil {
							return err
						}
					}

				}

				if !m.DB.DisableForeignKeyConstraintWhenMigrating && !m.DB.IgnoreRelationshipsWhenMigrating {
					for _, rel := range stmt.Schema.Relationships.Relations {
						if rel.Field.IgnoreMigration {
							continue
						}
						if constraint := rel.ParseConstraint(); constraint != nil &&
							constraint.Schema == stmt.Schema && !queryTx.Migrator().HasConstraint(value, constraint.Name) {
							if err := execTx.Migrator().CreateConstraint(value, constraint.Name); err != nil {
								return err
							}
						}
					}
				}

				for _, chk := range parseCheckConstraints {
					if !queryTx.Migrator().HasConstraint(value, chk.Name) {
						if err := execTx.Migrator().CreateConstraint(value, chk.Name); err != nil {
							return err
						}
					}
				}

				for _, idx := range parseIndexes {
					if !queryTx.Migrator().HasIndex(value, idx.Name) {
						if err := execTx.Migrator().CreateIndex(value, idx.Name); err != nil {
							return err
						}
					}
				}

				return nil
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m Migrator) MigrateColumn(value interface{}, field *schema.Field, columnType gorm.ColumnType) error {
	if field.IgnoreMigration {
		return nil
	}

	fullDataType := strings.TrimSpace(strings.ToLower(m.DB.Migrator().FullDataTypeOf(field).SQL))
	realDataType := strings.ToLower(columnType.DatabaseTypeName())

	var (
		alterColumn bool
		isSameType  = fullDataType == realDataType
	)

	if !field.PrimaryKey {
		if !strings.HasPrefix(fullDataType, realDataType) {
			aliases := m.DB.Migrator().GetTypeAliases(realDataType)
			for _, alias := range aliases {
				if strings.HasPrefix(fullDataType, alias) {
					log.Printf("[Migrator.MigrateColumn] [%s] data type is an alias - %s == %s\n", field.DBName, fullDataType, alias)
					isSameType = true
					break
				}
			}

			if !isSameType {
				alterColumn = true
			}
		}
	}

	if !isSameType {
		if length, ok := columnType.Length(); length != int64(field.Size) {
			if length > 0 && field.Size > 0 {
				log.Printf("[Migrator.MigrateColumn] [%s] column field.size has changed - %d == %d\n", field.DBName, length, field.Size)
				alterColumn = true
			} else {
				matches2 := regFullDataType.FindAllStringSubmatch(fullDataType, -1)
				if !field.PrimaryKey &&
					(len(matches2) == 1 && matches2[0][1] != fmt.Sprint(length) && ok) {
					log.Printf("[Migrator.MigrateColumn] [%s] column field.size has changed - %d == %s\n", field.DBName, length, matches2[0][1])
					alterColumn = true
				}
			}
		}

		if precision, _, ok := columnType.DecimalSize(); ok && int64(field.Precision) != precision {
			if regexp.MustCompile(fmt.Sprintf("[^0-9]%d[^0-9]", field.Precision)).MatchString(m.Migrator.DataTypeOf(field)) {
				log.Printf("[Migrator.MigrateColumn] [%s] column precision has changed\n", field.DBName)
				alterColumn = true
			}
		}
	}

	if nullable, ok := columnType.Nullable(); ok && nullable == field.NotNull {
		if !field.PrimaryKey && nullable {
			log.Printf("[Migrator.MigrateColumn] [%s] column should be made nullable\n", field.DBName)
			alterColumn = true
		}
	}

	if !field.PrimaryKey {
		// why: clickhouse NOT NULL is the default. A column needs to explicitly set NULL or wrap the col def in Nullable.o
		// to accept a NULL. as a result, all of the columns have a default value.
		// as such, in this case, if the dv="", we will do nothing. now, this is an issue. if we explicitly wanted to disallow empty strings,
		// we'd need to handle them w/ a nullable field. but this is fine for our use-case.
		currentDefaultNotNull := field.HasDefaultValue && (field.DefaultValueInterface != nil || !strings.EqualFold(field.DefaultValue, "NULL"))
		dv, dvNotNull := columnType.DefaultValue()
		if dvNotNull && !currentDefaultNotNull {
			log.Printf("[Migrator.MigrateColumn] [%s] dv=%t: default value has changed (\"%s\" -> null) - currentDefaultNotNull=%t dvNotNull=%t\n", field.DBName, field.HasDefaultValue, dv, currentDefaultNotNull, dvNotNull)
			// why: explicit override: we do nothing in this case because we do not want to support this mutation
			log.Printf("[Migrator.MigrateColumn] [%s] politely refusing to alter this column", field.DBName)
			alterColumn = false
		} else if !dvNotNull && currentDefaultNotNull {
			log.Printf("[Migrator.MigrateColumn] [%s] dv=%t: default value has changed (null -> \"%s\") - currentDefaultNotNull=%t dvNotNull=%t\n", field.DBName, field.HasDefaultValue, dv, currentDefaultNotNull, dvNotNull)
			alterColumn = true
		} else if currentDefaultNotNull || dvNotNull {
			switch field.GORMDataType {
			case schema.Time:
				if !strings.EqualFold(strings.TrimSuffix(dv, "()"), strings.TrimSuffix(field.DefaultValue, "()")) {
					alterColumn = true
				}
			case schema.Bool:
				v1, _ := strconv.ParseBool(dv)
				v2, _ := strconv.ParseBool(field.DefaultValue)
				alterColumn = v1 != v2
			default:
				alterColumn = dv != field.DefaultValue
			}
		}
	}

	if comment, ok := columnType.Comment(); ok && comment != field.Comment {
		if !field.PrimaryKey {
			log.Printf("[Migrator.MigrateColumn] [%s] comment has changed\n", field.DBName)
			alterColumn = true
		}
	}

	if alterColumn {
		log.Printf("[Migrator.MigrateColumn] preparing to alter column \"%s\" although fullDataType != realDataType: \"%s\" != \"%s\"\n", field.DBName, fullDataType, realDataType)
		if err := m.DB.Migrator().AlterColumn(value, field.DBName); err != nil {
			return err
		}
	}

	if err := m.DB.Migrator().MigrateColumnUnique(value, field, columnType); err != nil {
		return err
	}

	return nil
}

func filterWhitespace(array []string) []string {
	returnValue := []string{}
	for _, el := range array {
		trimmed := strings.TrimSpace(el)
		if trimmed != "" {
			returnValue = append(returnValue, trimmed)
		}
	}
	return returnValue
}

func cleanWhiteSpace(s string) string {
	parts := strings.Split(s, " ")
	filtered := filterWhitespace(parts)
	cleaned := strings.Join(filtered, " ")
	return cleaned
}

func (m Migrator) parseNestedField(definition string) (map[string]string, error) {
	childFields := map[string]string{}
	fieldsStr := strings.Replace(strings.TrimSpace(definition), "Nested(", "", 1)
	fieldsStr = fieldsStr[:len(fieldsStr)-1]

	cleaned := cleanWhiteSpace(fieldsStr)

	matches := []string{}
	for _, match := range regColumnDefinition.FindAllString(cleaned, -1) {
		cleanMatch := cleanWhiteSpace(match)
		matches = append(matches, cleanMatch)
	}

	// why: we use a regex because comma's are not enough
	// we want to map on "key Definition()"
	for _, field := range matches {
		parts := filterWhitespace(strings.Split(field, " "))
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(strings.Join(parts[1:], " "))
		childFields[key] = value
	}
	return childFields, nil
}

func (m Migrator) MigrateNestedColumn(value interface{}, field *schema.Field, childColumns []gorm.ColumnType) error {
	// why: The ALTER query for elements in a nested data structure has limitations.
	//   > docs: https://clickhouse.com/docs/en/sql-reference/data-types/nested-data-structures/nested
	//   > docs: https://clickhouse.com/docs/en/sql-reference/statements/alter/column#limitations

	// why: this is a "dumb"/simple strategy. a change in the order of any of the Nested field content triggers a migration.
	// TODO(fd): ensure we are cleaning up whitespace carefully so we do not run unnecessary migrations due to mere whitespace.

	allKeys := map[string]struct{}{}

	childFieldDefinitionMap, err := m.parseNestedField(string(field.DataType))
	if err != nil {
		return err
	}
	log.Printf("[Migrator.MigrateNestedColumn] Desired State")
	for column, definition := range childFieldDefinitionMap {
		allKeys[column] = struct{}{}
		log.Printf("[Migrator.MigrateNestedColumn]  >    colname:\"%s\"", column)
		log.Printf("[Migrator.MigrateNestedColumn]  > definition:\"%s\"", definition)
	}

	childFieldCurrentStateMap := map[string]string{}
	log.Printf("[Migrator.MigrateNestedColumn] Table State")
	for _, col := range childColumns {
		key := strings.Split(col.Name(), ".")[1]
		definition := col.DatabaseTypeName()
		childFieldCurrentStateMap[key] = definition
		allKeys[key] = struct{}{}
		log.Printf("[Migrator.MigrateNestedColumn]  >    colname:\"%s\" <= \"%s\"", key, col.Name())
		log.Printf("[Migrator.MigrateNestedColumn]  > definition:\"%s\"", definition)
	}

	additions := map[string]string{}
	deletions := []string{}
	modifications := map[string]string{}

	for key := range allKeys {
		desired, defOk := childFieldDefinitionMap[key]
		current, curOk := childFieldCurrentStateMap[key]

		if defOk && !curOk {
			additions[key] = desired
			break
		} else if !defOk && curOk {
			deletions = append(deletions, key)
			break
		} else if current != fmt.Sprintf("Array(%s)", desired) {
			modifications[key] = desired
			break
		}
	}

	if len(additions) == 0 && len(deletions) == 0 && len(modifications) == 0 {
		log.Println("[Migrator.MigrateNestedColumn] no changes required")
		return nil
	}

	if len(additions) > 0 {
		for key, value := range additions {
			column := strings.ToLower(fmt.Sprintf("%s.%s", field.Name, key))
			definition := fmt.Sprintf("Array(%s)", value)
			log.Printf("[Migrator.MigrateNestedColumn] Addition required")
			log.Printf("[Migrator.MigrateNestedColumn]  >     column:\"%s\" <= \"%s\"", column, key)
			log.Printf("[Migrator.MigrateNestedColumn]  > definition:\"%s\"", definition)
			emptyStruct := struct{}{}
			err := m.RunWithValue(emptyStruct, func(stmt *gorm.Statement) error {
				clusterOpts := ""
				if clusterOption, ok := m.DB.Get("gorm:table_cluster_options"); ok {
					clusterOpts = " " + fmt.Sprint(clusterOption) + " "
				}
				sQL := fmt.Sprintf("ALTER TABLE ? %s ADD COLUMN ? ?", clusterOpts)
				return m.DB.Exec(
					sQL,
					clause.Table{Name: stmt.Table},
					clause.Column{Name: column},
					clause.Expr{SQL: definition},
				).Error
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	if len(deletions) > 0 {
		for _, key := range deletions {
			column := fmt.Sprintf("%s.%s", strings.ToLower(field.Name), key)
			log.Printf("[Migrator.MigrateNestedColumn] Deletion required")
			log.Printf("[Migrator.MigrateNestedColumn]  > column:\"%s\"", column)
			emptyStruct := struct{}{}
			err := m.RunWithValue(emptyStruct, func(stmt *gorm.Statement) error {
				clusterOpts := ""
				if clusterOption, ok := m.DB.Get("gorm:table_cluster_options"); ok {
					clusterOpts = " " + fmt.Sprint(clusterOption) + " "
				}
				sQL := fmt.Sprintf("ALTER TABLE ? %s DROP COLUMN ?", clusterOpts)
				return m.DB.Exec(
					sQL,
					clause.Table{Name: stmt.Table},
					clause.Column{Name: column},
				).Error
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if len(modifications) > 0 {
		for key, value := range modifications {
			column := strings.ToLower(fmt.Sprintf("%s.%s", field.Name, key))
			definition := fmt.Sprintf("Array(%s)", value)
			log.Printf("[Migrator.MigrateNestedColumn] Modification required")
			log.Printf("[Migrator.MigrateNestedColumn]  > column:\"%s\" <= \"%s\"", column, key)
			log.Printf("[Migrator.MigrateNestedColumn]  > definition:\"%s\"", definition)
			emptyStruct := struct{}{}
			err := m.RunWithValue(emptyStruct, func(stmt *gorm.Statement) error {
				clusterOpts := ""
				if clusterOption, ok := m.DB.Get("gorm:table_cluster_options"); ok {
					clusterOpts = " " + fmt.Sprint(clusterOption) + " "
				}
				sQL := fmt.Sprintf("ALTER TABLE ? %s MODIFY COLUMN ? ?", clusterOpts)
				return m.DB.Exec(
					sQL,
					clause.Table{Name: stmt.Table},
					clause.Column{Name: column},
					clause.Expr{SQL: definition},
				).Error
			})
			if err != nil {
				log.Fatal(err)
			}

		}
	}

	return nil
}

func (m Migrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	columnTypes := make([]gorm.ColumnType, 0)
	execErr := m.RunWithValue(value, func(stmt *gorm.Statement) (err error) {
		rows, err := m.DB.Session(&gorm.Session{}).Table(stmt.Table).Limit(1).Rows()
		if err != nil {
			return err
		}

		defer func() {
			if err == nil {
				err = rows.Close()
			}
		}()

		var rawColumnTypes []*sql.ColumnType
		rawColumnTypes, err = rows.ColumnTypes()

		columnTypeSQL := "SELECT name, type, default_expression, comment, is_in_primary_key, character_octet_length, numeric_precision, numeric_precision_radix, numeric_scale, datetime_precision FROM system.columns WHERE database = ? AND table = ?"
		if m.DontSupportColumnPrecision {
			columnTypeSQL = "SELECT name, type, default_expression, comment, is_in_primary_key FROM system.columns WHERE database = ? AND table = ?"
		}
		columns, rowErr := m.DB.Raw(columnTypeSQL, m.CurrentDatabase(), stmt.Table).Rows()
		if rowErr != nil {
			return rowErr
		}

		defer columns.Close()

		for columns.Next() {
			var (
				column            migrator.ColumnType
				decimalSizeValue  *uint64
				datetimePrecision *uint64
				radixValue        *uint64
				scaleValue        *uint64
				lengthValue       *uint64
				values            = []interface{}{
					&column.NameValue, &column.DataTypeValue, &column.DefaultValueValue, &column.CommentValue, &column.PrimaryKeyValue, &lengthValue, &decimalSizeValue, &radixValue, &scaleValue, &datetimePrecision,
				}
			)

			if m.DontSupportColumnPrecision {
				values = []interface{}{&column.NameValue, &column.DataTypeValue, &column.DefaultValueValue, &column.CommentValue, &column.PrimaryKeyValue}
			}

			if scanErr := columns.Scan(values...); scanErr != nil {
				return scanErr
			}

			column.ColumnTypeValue = column.DataTypeValue

			if decimalSizeValue != nil {
				column.DecimalSizeValue.Int64 = int64(*decimalSizeValue)
				column.DecimalSizeValue.Valid = true
			} else if datetimePrecision != nil {
				column.DecimalSizeValue.Int64 = int64(*datetimePrecision)
				column.DecimalSizeValue.Valid = true
			}

			if scaleValue != nil {
				column.ScaleValue.Int64 = int64(*scaleValue)
				column.ScaleValue.Valid = true
			}

			if lengthValue != nil {
				column.LengthValue.Int64 = int64(*lengthValue)
				column.LengthValue.Valid = true
			}

			if column.DefaultValueValue.Valid {
				column.DefaultValueValue.String = strings.Trim(column.DefaultValueValue.String, "'")
			}

			if m.DontSupportEmptyDefaultValue && column.DefaultValueValue.String == "" {
				column.DefaultValueValue.Valid = false
			}

			for _, c := range rawColumnTypes {
				if c.Name() == column.NameValue.String {
					column.SQLColumnType = c
					break
				}
			}

			columnTypes = append(columnTypes, column)
		}

		return
	})

	return columnTypes, execErr
}

func (m Migrator) BuildIndexOptions(opts []schema.IndexOption, stmt *gorm.Statement) (results []interface{}) {
	for _, indexOpt := range opts {
		str := stmt.Quote(indexOpt.DBName)
		if indexOpt.Expression != "" {
			str = indexOpt.Expression
		}
		results = append(results, clause.Expr{SQL: str})
	}
	return
}

func (m Migrator) CreateIndex(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if index := stmt.Schema.LookIndex(name); index != nil {
			opts := m.BuildIndexOptions(index.Fields, stmt)
			values := []interface{}{
				clause.Table{Name: stmt.Table},
				clause.Column{Name: index.Name},
				opts,
			}

			indexType := m.DefaultIndexType
			if index.Type != "" {
				indexType = index.Type
			}

			// why: concept of UNIQUE | FULLTEXT | SPATIAL index
			// is NOT supported in clickhouse
			createIndexSQL := "ALTER TABLE ? ADD INDEX ? ? TYPE %s GRANULARITY %d" // TODO(iqdf): how to inject Granularity
			createIndexSQL = fmt.Sprintf(createIndexSQL, indexType, m.getIndexGranularityOption(index.Fields))
			return m.DB.Exec(createIndexSQL, values...).Error
		}
		return ErrCreateIndexFailed
	})
}

func (m Migrator) RenameIndex(value interface{}, oldName, newName string) error {
	// TODO(iqdf): drop index and add the index again with different name
	// DROP INDEX ?
	// ADD INDEX ? TYPE ? GRANULARITY ?
	return ErrRenameIndexUnsupported
}

func (m Migrator) DropIndex(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if idx := stmt.Schema.LookIndex(name); idx != nil {
				name = idx.Name
			}
		}
		dropIndexSQL := "ALTER TABLE ? DROP INDEX ?"
		return m.DB.Exec(dropIndexSQL,
			clause.Table{Name: stmt.Table},
			clause.Column{Name: name}).Error
	})
}

func (m Migrator) HasIndex(value interface{}, name string) bool {
	var count int
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		currentDatabase := m.DB.Migrator().CurrentDatabase()

		if idx := stmt.Schema.LookIndex(name); idx != nil {
			name = idx.Name
		}

		showCreateTableSQL := fmt.Sprintf("SHOW CREATE TABLE %s.%s", currentDatabase, stmt.Table)
		var createStmt string
		if err := m.DB.Raw(showCreateTableSQL).Row().Scan(&createStmt); err != nil {
			return err
		}

		indexNames := m.extractIndexNamesFromCreateStmt(createStmt)

		for _, indexName := range indexNames {
			if indexName == name {
				count = 1
				break
			}
		}
		return nil
	})

	return count > 0
}

func (m Migrator) getIndexGranularityOption(opts []schema.IndexOption) int {
	for _, indexOpt := range opts {
		if settingStr, ok := indexOpt.TagSettings["INDEX"]; ok {
			for _, str := range strings.Split(settingStr, ",") {
				keyVal := strings.Split(str, ":")
				if len(keyVal) > 1 && strings.ToLower(keyVal[0]) == "granularity" {
					if len(keyVal) < 2 {
						continue
					}
					num, err := strconv.Atoi(keyVal[1])
					if err != nil || num < 0 {
						continue
					}
					return num
				}
			}
		}
	}
	return m.DefaultGranularity
}

func (m Migrator) extractIndexNamesFromCreateStmt(createStmt string) []string {
	var names []string
	scanner := bufio.NewScanner(strings.NewReader(createStmt))
	state := 0
	for scanner.Scan() && state < 2 {
		line := scanner.Text()
		switch state {
		case 0:
			if strings.HasPrefix(line, "(") {
				state = 1
			}
		case 1:
			if strings.HasPrefix(line, ")") {
				state = 2
				continue
			}
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "INDEX ") {
				line = strings.TrimPrefix(line, "INDEX ")
				elems := strings.Split(line, " ")
				if len(elems) > 0 {
					names = append(names, elems[0])
				}
			}
		}
	}
	return names
}
