package mysqlL

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/flyerxp/lib/app"
	config2 "github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/nacos"
	"github.com/flyerxp/lib/utils/json"
	yaml2 "github.com/flyerxp/lib/utils/yaml"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/qustavo/sqlhooks/v2"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Mysql 容器
type SqlContainer struct {
	SqlContainer cmap.ConcurrentMap[string, *MysqlClient]
	MysqlConf    cmap.ConcurrentMap[string, config2.MidMysqlConf]
}

// Mysql 客户端
type MysqlClient struct {
	Poll   *sync.Pool
	CurrDb *sqlx.DB
}

var MysqlEngine *SqlContainer

type MysqlLog struct {
}

func (m *MysqlLog) Print(ctx context.Context, v ...interface{}) {
	zapLog := make([]zap.Field, len(v))
	for i := range v {
		switch v[i].(type) {
		case error:
			zapLog[i] = zap.Error(v[i].(error))
		case string:
			zapLog[i] = zap.String("mysql driver error", v[i].(string))
		default:
			zapLog[i] = zap.Any("mysql driver error", v[i])
		}
	}
	logger.AddError(ctx, zapLog...)
}
func GetEngine(ctx context.Context, name string) (*MysqlClient, error) {
	if MysqlEngine == nil {
		MysqlEngine = new(SqlContainer)
		var confList []config2.MidMysqlConf
		MysqlEngine.MysqlConf = cmap.New[config2.MidMysqlConf]()
		MysqlEngine.SqlContainer = cmap.New[*MysqlClient]()
		conf := config2.GetConf()
		confList = conf.Mysql
		//本地文件中获取
		for _, v := range confList {
			if v.Name != "" {
				MysqlEngine.MysqlConf.Set(v.Name, v)
			}
		}
		//nacos获取
		if conf.MysqlNacos.Name != "" {
			var yaml []byte
			mysqlList := new(config2.MysqlConf)
			ns, e := nacos.GetEngine(ctx, conf.MysqlNacos.Name)
			if e == nil {
				yaml, e = ns.GetConfig(ctx, conf.MysqlNacos.Did, conf.MysqlNacos.Group, conf.MysqlNacos.Ns)
				if e == nil {
					e = yaml2.DecodeByBytes(yaml, mysqlList)
					if e == nil {
						for _, v := range mysqlList.List {
							MysqlEngine.MysqlConf.Set(v.Name, v)
						}
					} else {
						logger.AddError(ctx, zap.Error(errors.New("yaml conver error")))
					}
				}
			}
		}
		_ = app.RegisterFunc("mysql", "mysql close", func() {
			MysqlEngine.Reset()
		})
	}

	e, ok := MysqlEngine.SqlContainer.Get(name)
	if ok {
		return e, nil
	}
	o, okC := MysqlEngine.MysqlConf.Get(name)
	if okC {
		hook := new(Hooks)
		if o.SqlLog == "yes" {
			hook.IsPrintSQLDuration = true
		}
		hook.DbName = o.Name
		//_ = mysql.SetLogger(&MysqlLog{})
		sql.Register("mysqlWithHooks_"+o.Name, sqlhooks.Wrap(&mysql.MySQLDriver{}, hook))
		objMysql := newClient(ctx, o)
		go func() {
			objMysql.GetDb()
		}()
		MysqlEngine.SqlContainer.Set(name, objMysql)
		return objMysql, nil
	} else {
		logger.AddError(ctx, zap.Error(errors.New("no find mysql config "+name)))
		panic(errors.New(name + " db config no find"))
	}
}

// https://github.com/golang-migrate/migrate/blob/master/database/mysql/README.md
func newClient(ctx context.Context, o config2.MidMysqlConf) *MysqlClient {
	c := &sync.Pool{
		New: func() any {
			start := time.Now()
			var dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?allowCleartextPasswords=true&checkConnLiveness=false&parseTime=true&interpolateParams=true&loc=Local", o.User, o.Pwd, o.Address, o.Port, o.Db) //"user=" + o.User + " host=" + o.Address + " port=" + o.Port + " dbname=" + o.Db
			if o.CharSet != "" {
				dsn = dsn + "&charset=" + o.CharSet
			}
			if o.ReadTimeout > 0 {
				dsn = dsn + "&readTimeout=" + strconv.Itoa(o.ReadTimeout) + "ms"
			}
			if o.ConnTimeout > 0 {
				dsn = dsn + "&timeout=" + strconv.Itoa(o.ConnTimeout) + "ms"
			}
			if o.WriteTimeout > 0 {
				dsn = dsn + "&writeTimeout=" + strconv.Itoa(o.WriteTimeout) + "ms"
			}
			if o.Collation != "" {
				dsn = dsn + "&collation=" + o.Collation
			}
			n, e := sqlx.Open("mysqlWithHooks_"+o.Name, dsn)
			go func() {
				if n.Ping() != nil {
					logger.AddError(ctx, zap.Error(errors.New("dsn link fail:"+o.Address)))
				}
			}()

			logger.AddMysqlConnTime(ctx, int(time.Since(start).Microseconds()))
			if e != nil {
				logger.AddError(ctx, zap.String("dsn link fail ", o.Name+"|"+o.Address), zap.Error(e))
				panic(e.Error())
			}
			if o.MaxIdleConns > 0 {
				n.SetMaxIdleConns(o.MaxIdleConns)
			}
			if o.MaxOpenConns > 0 {
				n.SetMaxOpenConns(o.MaxOpenConns)
			}
			if o.MaxLifetime > 0 {
				n.SetConnMaxLifetime(time.Duration(o.MaxLifetime) * time.Second)
			}
			if o.MaxIdleTime > 0 {
				n.SetConnMaxIdleTime(time.Duration(o.MaxIdleTime) * time.Second)
			}
			return n
		},
	}

	return &MysqlClient{c, nil}
}
func (m *MysqlClient) GetDb() *sqlx.DB {
	if m.CurrDb == nil {
		m.CurrDb = m.Poll.Get().(*sqlx.DB)
	}
	return m.CurrDb
}
func (m *MysqlClient) PutDb(a *sqlx.DB) {
	m.Poll.Put(a)
}
func (m *MysqlClient) CloseDb() {
	m.Poll.Put(m.CurrDb)
	m.CurrDb = nil
}
func (m *SqlContainer) Reset() {
	if MysqlEngine != nil {
		for _, v := range MysqlEngine.SqlContainer.Items() {
			if v.CurrDb != nil {
				if v.CurrDb != nil {
					_ = v.CurrDb.Close()
				}
			}
		}
		MysqlEngine = nil
	}
}
func GetUpdateSql(table string, fields []string) string {
	upFields := make([]string, len(fields))
	for i := range fields {
		upFields[i] = "`" + fields[i] + "`=:" + fields[i]
	}
	return fmt.Sprintf("update `%s` set %s where ", table, strings.Join(upFields, ","))
}

type fieldExt struct {
	F string `json:"f"`
	V string `json:"v"`
}

func GetInsertSql(table string, fields []string) string {
	iFields := make([]string, len(fields))
	iValues := make([]string, len(fields))
	for i := range fields {
		if fields[i][0:6] == "{\"f\":\"" && strings.Index(fields[i], "\"v\":\"") > 0 {
			tTmp := fieldExt{}
			e := json.Decode([]byte(fields[i]), &tTmp)
			if e == nil {
				iFields[i] = "`" + tTmp.F + "`"
				iValues[i] = fmt.Sprintf("%s", tTmp.V)
			}
		} else {
			iFields[i] = "`" + fields[i] + "`"
			iValues[i] = fmt.Sprintf(":%s", fields[i])
		}
	}
	return fmt.Sprintf("insert into `%s` (%s) values(%s)", table, strings.Join(iFields, ","), strings.Join(iValues, ","))
}
func IsNoFindRowErr(err error) bool {
	return err.Error() == "sql: no rows in result set"
}
