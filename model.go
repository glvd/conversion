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
type Model struct {
	ID        string     `xorm:"id pk"`
	CreatedAt time.Time  `xorm:"created_at created"`
	UpdatedAt time.Time  `xorm:"updated_at updated"`
	DeletedAt *time.Time `xorm:"deleted_at deleted"`
	Version   int        `xorm:"version"`
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

// IModel ...
type IModel interface {
	Table() *xorm.Session
	GetID() string
	SetID(string)
	GetVersion() int
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

// GetID ...
func (m Model) GetID() string {
	return m.ID
}

// SetID ...
func (m *Model) SetID(id string) {
	m.ID = id
}

// GetVersion ...
func (m Model) GetVersion() int {
	return m.Version
}

// SetVersion ...
func (m *Model) SetVersion(v int) {
	m.Version = v
}

// BeforeInsert ...
func (m *Model) BeforeInsert() {
	if m.ID == "" {
		m.ID = UUID().String()
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

// RegisterTable ...
func RegisterTable(m ISync) {
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
	i, e = m.Table().InsertOne(m)
	if e != nil {
		return 0, e
	}
	return i, e
}

// MustSession ...
func MustSession(session *xorm.Session) *xorm.Session {
	if session == nil {
		panic("session is nil")
	}
	return session
}

// IsExist ...
func IsExist(m IModel) bool {
	i, e := m.Table().
		Where("id = ?", m.GetID()).
		Count()
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
