package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"supernova/pkg/conf"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func initMysql(host, port, userName, password, dbName string, maxIdleConns, maxOpenConns int) (*gorm.DB, error) {
	var (
		err    error
		sqlDB  *sql.DB
		gormDB *gorm.DB
	)

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		userName, password, host, port, dbName)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Warn,
		},
	)

	gormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		klog.Errorf("Failed to connect to mysql:%v, with dsn:%v", err, dsn)
		return nil, err
	}

	sqlDB, err = gormDB.DB()
	if err != nil {
		klog.Errorf("Failed to connect get mysql DB:%v, with dsn:%v", err, dsn)
	}

	// 设置最大空闲连接数和最大打开连接数
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	return gormDB, nil
}

func InitMysql(mysqlConf *conf.MysqlConf) (*gorm.DB, error) {
	if db, err := initMysql(mysqlConf.Host, mysqlConf.Port,
		mysqlConf.UserName, mysqlConf.Password, mysqlConf.DbName,
		mysqlConf.MaxIdleConnections, mysqlConf.MaxOpenConnections); err != nil {
		return nil, fmt.Errorf("fail to connect mysql at config:%+v, error:%v", mysqlConf, err)
	} else {
		klog.Infof("mysql init success with config:%+v", mysqlConf)
		return db, nil
	}
}
