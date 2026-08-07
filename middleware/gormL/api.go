package gormL

import (
	"context"
	"errors"
	"fmt"
	"github.com/flyerxp/lib/v2/app"
	config2 "github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"github.com/flyerxp/lib/v2/middleware/nacos"
	yaml2 "github.com/flyerxp/lib/v2/utils/yaml"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"time"
)

// GormContainer 容器
type GormContainer struct {
	GormClient cmap.ConcurrentMap[string, *gorm.DB]
	GormConf   cmap.ConcurrentMap[string, config2.MidMysqlConf]
	MyLock     *sync.Mutex
	IsEnd      bool //是否初始化完成
}

var GormEngine *GormContainer

func GetEngine(ctx context.Context, name string) (*gorm.DB, error) {
	if GormEngine == nil || !GormEngine.IsEnd {
		if GormEngine == nil {
			GormEngine = new(GormContainer)
			GormEngine.MyLock = new(sync.Mutex)
			GormEngine.GormConf = cmap.New[config2.MidMysqlConf]()
			GormEngine.GormClient = cmap.New[*gorm.DB]()
		}
		GormEngine.MyLock.Lock()
		defer func() {
			GormEngine.MyLock.Unlock()
		}()
		if GormEngine.GormConf.IsEmpty() {
			conf := config2.GetConf()
			confList := conf.Mysql
			//本地文件中获取
			for _, v := range confList {
				if v.Name != "" {
					GormEngine.GormConf.Set(v.Name, v)
				}
			}
			//从nacos的mysql配置中获取（GORM也走mysql）
			if conf.MysqlNacos.Name != "" {
				var yaml []byte
				gormList := new(config2.MysqlConf)
				ns, e := nacos.GetEngine(ctx, conf.MysqlNacos.Name)
				if e == nil {
					yaml, e = ns.GetConfig(ctx, conf.MysqlNacos.Did, conf.MysqlNacos.Group, conf.MysqlNacos.Ns)
					if e == nil {
						e = yaml2.DecodeByBytes(yaml, gormList)
						if e == nil {
							for _, v := range gormList.List {
								GormEngine.GormConf.Set(v.Name, v)
							}
						} else {
							logger.AddError(ctx, zap.Error(errors.New("yaml conver error")))
						}
						GormEngine.IsEnd = true
					}
				}
			}
			_ = app.RegisterFunc("gorm", "gorm close", func() {
				GormEngine.Reset()
			})
		}
	}
	e, ok := GormEngine.GormClient.Get(name)
	if ok {
		return e, nil
	}
	o, okC := GormEngine.GormConf.Get(name)
	if okC {
		db := newClient(ctx, o)
		GormEngine.GormClient.Set(name, db)
		return db, nil
	} else {
		logger.AddError(ctx, zap.Error(errors.New("no find gorm config "+name)))
		panic(errors.New(name + " db config no find"))
	}
}

func newClient(ctx context.Context, o config2.MidMysqlConf) *gorm.DB {
	start := time.Now()
	var dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?allowCleartextPasswords=true&checkConnLiveness=false&parseTime=true&interpolateParams=true&loc=Local",
		o.User, o.Pwd, o.Address, o.Port, o.Db)
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

	plugin := &GormPlugin{
		IsPrintSQLDuration: o.SqlLog == "yes",
		DbName:             o.Name,
	}

	var poolCtx context.Context
	if ctx.Value(logger.GetLogIdKey()) != nil {
		poolCtx = ctx
	} else {
		poolCtx = logger.GetContext(context.Background(), fmt.Sprintf("gorm_pool_%s", o.Name))
	}
	plugin.baseCtx = poolCtx

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: plugin,
	})
	if err != nil {
		logger.AddError(ctx, zap.String("dsn link fail ", o.Name+"|"+o.Address), zap.Error(err))
		panic(err.Error())
	}

	logger.AddGormConnTime(ctx, int(time.Since(start).Microseconds()))

	sqlDB, err := db.DB()
	if err != nil {
		logger.AddError(ctx, zap.String("get sql.DB fail ", o.Name+"|"+o.Address), zap.Error(err))
		panic(err.Error())
	}

	if o.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(o.MaxIdleConns)
	}
	if o.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(o.MaxOpenConns)
	}
	if o.MaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(o.MaxLifetime) * time.Second)
	}
	if o.MaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(o.MaxIdleTime) * time.Second)
	}

	return db
}

func (m *GormContainer) Reset() {
	if GormEngine != nil {
		for _, v := range GormEngine.GormClient.Items() {
			if sqlDB, err := v.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
		GormEngine = nil
	}
}
