package dal

import (
	"gorm.io/gorm"
	"supernova/pkg/conf"
	"supernova/pkg/session/mysql"
)

type MysqlClient struct {
	db *gorm.DB
}

func NewMysqlClient(mysqlConf *conf.MysqlConf) (*MysqlClient, error) {
	if db, err := mysql.InitMysql(mysqlConf); err != nil {
		return nil, err
	} else {
		return &MysqlClient{db: db}, nil
	}
}

func (c *MysqlClient) DB() *gorm.DB {
	return c.db
}
