package accessor

import (
	"context"
	"practice/internal/config"
	"practice/internal/storage/rdb"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type shutdownHandler func(context.Context)

type accessor struct {
	shutdownOnce     sync.Once
	shutdownHandlers []shutdownHandler

	Config *config.Config // configuration management
	RDB    *rdb.Rdb       // relational database instance
}

func BuildAccessor() *accessor {
	return &accessor{
		Config: config.NewFromViper(),
	}
}

func (a *accessor) Close(ctx context.Context) {
	a.shutdownOnce.Do(func() {
		logrus.Infoln("start to close accessors.")
		for _, handler := range a.shutdownHandlers {
			handler(ctx)
		}
	})

	logrus.Infoln("all accessors closed.")
}

func (a *accessor) InitRDB(ctx context.Context) {
	switch a.Config.RDB.Driver {
	case "mysql":
		a.RDB = rdb.NewMysqlClient(ctx,
			a.Config.RDB.MysqlOpts.UserName,
			a.Config.RDB.MysqlOpts.Password,
			a.Config.RDB.MysqlOpts.Address,
			a.Config.RDB.MysqlOpts.DBName,
			time.Duration(a.Config.RDB.MysqlOpts.ConnMaxLifetime)*time.Minute,
			a.Config.RDB.MysqlOpts.MaxOpenConns,
			a.Config.RDB.MysqlOpts.MaxIdleConns,
		)
	default:
		logrus.Panicf("RDB driver undifined: %v", a.Config.RDB.Driver)
	}

	a.shutdownHandlers = append(a.shutdownHandlers, func(c context.Context) {
		a.RDB.Shutdown(c)
		logrus.Infoln("relational database accessor closed.")
	})

	logrus.Infoln("initial relational database accessor successful.")
}
