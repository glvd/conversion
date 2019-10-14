package conversion

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/google/uuid"
)

const mysqlStatement = "%s:%s@tcp(%s)/%s?loc=%s&charset=%s&parseTime=true"

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

// ConfigOptions ...
type ConfigOptions func(config *dbConfig)

var tables struct {
	sync.Mutex
	table map[string]interface{}
}

var _database *xorm.Engine

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
func InitMySQL(ops ...ConfigOptions) *xorm.Engine {
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
		panic(e)
	}
	return engine
}

// Source ...
func (d *dbConfig) source() string {
	return fmt.Sprintf(mysqlStatement,
		d.username, d.password, d.addr, d.schema, d.location, d.charset)
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

// Model ...
type Model struct {
	ID        string     `xorm:"id pk"`
	CreatedAt time.Time  `xorm:"created_at created"`
	UpdatedAt time.Time  `xorm:"updated_at updated"`
	DeletedAt *time.Time `xorm:"deleted_at deleted"`
	Version   int        `xorm:"version"`
}

// Table ...
func (m *Model) Table() *xorm.Session {
	panic("implement me")
}

// Modeler ...
type Modeler interface {
	Table() *xorm.Session
	GetID() string
	SetID(string)
	GetVersion() int
	SetVersion(int)
}

// BeforeInsert ...
func (m *Model) BeforeInsert() {
	if m.ID == "" {
		m.ID = UUID().String()
	}
}

// MustSession ...
func MustSession(session *xorm.Session) *xorm.Session {
	if session == nil {
		panic("session is nil")
	}
	return session
}

// IsExist ...
func IsExist(session *xorm.Session, table interface{}) bool {
	i, e := session.Table(table).
		//Where("checksum = ?", unfin.Checksum).
		//Where("type = ?", unfin.Type).
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
