package accessor

import (
	"context"
	"fmt"
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
	RDB    rdb.RDB        // relational database instance
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		a.Config.MySQL.UserName,
		a.Config.MySQL.Password,
		a.Config.MySQL.Address,
		a.Config.MySQL.DBName,
	)

	a.RDB = rdb.NewMysqlClient(ctx, dsn,
		time.Duration(a.Config.MySQL.ConnMaxLifetime)*time.Minute,
		a.Config.MySQL.MaxOpenConns,
		a.Config.MySQL.MaxIdleConns,
	)

	a.shutdownHandlers = append(a.shutdownHandlers, func(c context.Context) {
		a.RDB.Shutdown(c)
		logrus.Infoln("relational database accessor closed.")
	})

	logrus.Infoln("initial relational database accessor successful.")
}
