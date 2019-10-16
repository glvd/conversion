package conversion

import (
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

const mysqlStatement = "%s:%s@tcp(%s)/%s?loc=%s&charset=%s&parseTime=true"

// Model ...
type model struct {
	ID        string     `xorm:"id pk"`
	CreatedAt time.Time  `xorm:"created_at created"`
	UpdatedAt time.Time  `xorm:"updated_at updated"`
	DeletedAt *time.Time `xorm:"deleted_at deleted"`
	Version   int        `xorm:"version"`
}

// Model ...
type Model struct {
	model `xorm:"extends"`
}

// DatabaseConfig ...
type dbConfig struct {
	showSQL      bool
	showExecTime bool
	useCache     bool
	dbType       string
	addr         string
	username     string
	password     string
	schema       string
	charset      string
	prefix       string
	location     string
}

// BeforeInsert ...
type BeforeInsert interface {
	BeforeInsert()
}

// IModel ...
type IModel interface {
	Table() interface{}
	ID() string
	SetID(string)
	Version() int
	SetVersion(int)
}

// ISync ...
type ISync interface {
	Sync() error
}

// ConfigOptions ...
type ConfigOptions func(config *dbConfig)

var (
	_              = mysql.Config{}
	_              = sqlite3.Error{}
	_database      *xorm.Engine
	_databaseTable map[string]ISync
)

// ShowSQLOptions ...
func ShowSQLOptions(b bool) ConfigOptions {
	return func(config *dbConfig) {
		config.showSQL = b
	}
}

// UseCacheOptions ...
func UseCacheOptions(b bool) ConfigOptions {
	return func(config *dbConfig) {
		config.useCache = b
	}
}

// SchemaOption ...
func SchemaOption(s string) ConfigOptions {
	return func(config *dbConfig) {
		config.schema = s
	}
}

// LoginOption ...
func LoginOption(addr, user, pass string) ConfigOptions {
	return func(config *dbConfig) {
		config.addr = addr
		config.username = user
		config.password = pass
	}
}

// InitMySQL ...
func InitMySQL(ops ...ConfigOptions) (*xorm.Engine, error) {
	config := &dbConfig{
		showSQL:  true,
		useCache: true,
		dbType:   "mysql",
		addr:     "localhost",
		username: "root",
		password: "111111",
		schema:   "glvd",
		location: url.QueryEscape("Asia/Shanghai"),
		charset:  "utf8mb4",
		prefix:   "",
	}
	for _, op := range ops {
		op(config)
	}

	engine, e := xorm.NewEngine(config.dbType, config.source())
	if e != nil {
		return nil, e
	}
	return engine, nil
}

// Source ...
func (d *dbConfig) source() string {
	return fmt.Sprintf(mysqlStatement,
		d.username, d.password, d.addr, d.schema, d.location, d.charset)
}

// ID ...
func (m Model) ID() string {
	return m.model.ID
}

// SetID ...
func (m *Model) SetID(id string) {
	m.model.ID = id
}

// Version ...
func (m Model) Version() int {
	return m.model.Version
}

// SetVersion ...
func (m *Model) SetVersion(v int) {
	m.model.Version = v
}

// BeforeInsert ...
func (m *Model) BeforeInsert() {
	if m.model.ID == "" {
		m.model.ID = UUID().String()
	}
}

func liteSource(name string) string {
	return fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL", name)
}

// InitSQLite3 ...
func InitSQLite3(name string) (eng *xorm.Engine, e error) {
	eng, e = xorm.NewEngine("sqlite3", liteSource(name))
	if e != nil {
		return nil, e
	}

	return eng, nil
}

// MustDatabase ...
func MustDatabase(engine *xorm.Engine, err error) *xorm.Engine {
	if err != nil {
		panic(err)
	}
	return engine
}

// RegisterDatabase ...
func RegisterDatabase(engine *xorm.Engine) {
	if _database == nil {
		_database = engine
	}
}

// registerTable ...
func registerTable(m ISync) {
	if _databaseTable == nil {
		_databaseTable = make(map[string]ISync)
	}
	_databaseTable[reflect.TypeOf(m).Name()] = m
}

// SyncTable ...
func SyncTable() (e error) {
	for _, v := range _databaseTable {
		if err := v.Sync(); err != nil {
			return err
		}
	}
	return nil
}

// InsertOrUpdate ...
func InsertOrUpdate(m IModel) (i int64, e error) {
	i, e = _database.InsertOne(m)
	if e != nil {
		return 0, e
	}
	return i, e
}

// FindResult ...
type FindResult func(rows *xorm.Rows) error

// FindAll ...
func FindAll(model IModel, f FindResult, limit int, start ...int) (e error) {
	table := _database.Table(model.Table())
	if limit == 0 {
		return nil
	} else if limit > 0 {
		table = table.Limit(limit, start...)
	}
	rows, e := table.Rows(model)
	if e != nil {
		return e
	}

	for rows.Next() {
		if err := f(rows); err != nil {
			return err
		}
	}

	return nil
}

// MustSession ...
func MustSession(session *xorm.Session) *xorm.Session {
	if session == nil {
		return _database.Where("")
	}
	return session
}

// IsExist ...
func IsExist(m IModel) bool {
	i, e := _database.
		Where("id = ?", m.ID()).
		Count(m.Table())
	if e != nil || i <= 0 {
		return false
	}
	return true
}

// UUID ...
func UUID() uuid.UUID {
	return uuid.Must(uuid.NewUUID())
}

// MustString  must string
func MustString(val, src string) string {
	if val != "" {
		return val
	}
	return src
}
