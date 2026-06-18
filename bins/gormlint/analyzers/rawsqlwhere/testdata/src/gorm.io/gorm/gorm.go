package gorm

type DB struct{}

func (db *DB) Where(query interface{}, args ...interface{}) *DB { return db }
func (db *DB) First(dest interface{}, conds ...interface{}) *DB { return db }
func (db *DB) Find(dest interface{}, conds ...interface{}) *DB  { return db }
func (db *DB) Preload(query string, args ...interface{}) *DB    { return db }
func (db *DB) Model(value interface{}) *DB                      { return db }
func (db *DB) Create(value interface{}) *DB                     { return db }
func (db *DB) Save(value interface{}) *DB                       { return db }
func (db *DB) Update(column string, value interface{}) *DB      { return db }
func (db *DB) Updates(values interface{}) *DB                   { return db }
func (db *DB) Delete(value interface{}, conds ...interface{}) *DB { return db }
func (db *DB) Scan(dest interface{}) *DB                        { return db }
func (db *DB) Count(count *int64) *DB                           { return db }
func (db *DB) Limit(limit int) *DB                              { return db }
func (db *DB) Order(value interface{}) *DB                      { return db }
func (db *DB) Select(query interface{}, args ...interface{}) *DB { return db }
func (db *DB) Joins(query string, args ...interface{}) *DB      { return db }
func (db *DB) Group(name string) *DB                            { return db }
func (db *DB) Having(query interface{}, args ...interface{}) *DB { return db }
func (db *DB) Last(dest interface{}, conds ...interface{}) *DB  { return db }
func (db *DB) Take(dest interface{}, conds ...interface{}) *DB  { return db }
func (db *DB) Exec(sql string, values ...interface{}) *DB       { return db }
func (db *DB) Raw(sql string, values ...interface{}) *DB        { return db }
func (db *DB) WithContext(ctx interface{}) *DB                  { return db }
func (db *DB) Pluck(column string, dest interface{}) *DB        { return db }

type Error = error
