package xorm

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xormplus/xorm"
	"time"
)

type MysqlConfig struct {
	Hostname     []string
	Database     string
	Username     string
	Password     string
	Hostport     string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifeTime  int
}

// Connect 单库连接
func Connect(conf *MysqlConfig) (*xorm.Engine, error) {
	source := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8", conf.Username, conf.Password, conf.Hostname[0], conf.Hostport, conf.Database)
	db, err := xorm.NewEngine("mysql", source)
	if err != nil {
		return nil, err
	}
	//最大空闲
	db.SetMaxIdleConns(conf.MaxIdleConns)
	//最大打开
	db.SetMaxOpenConns(conf.MaxOpenConns)
	//最大保持时间
	db.SetConnMaxLifetime(time.Duration(conf.MaxLifeTime) * time.Second)
	// 打印sql
	db.ShowSQL(true)

	return db, err
}

// GroupConnect 读写分离
func GroupConnect(conf *MysqlConfig) (*xorm.EngineGroup, error) {
	var source []string
	for _, host := range conf.Hostname {
		s := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8", conf.Username, conf.Password, host, conf.Hostport, conf.Database)
		source = append(source, s)
	}

	db, err := xorm.NewEngineGroup("mysql", source)
	if err != nil {
		return nil, err
	}

	//最大空闲
	db.SetMaxIdleConns(conf.MaxIdleConns)
	//最大打开
	db.SetMaxOpenConns(conf.MaxOpenConns)
	//最大保持时间
	db.SetConnMaxLifetime(time.Duration(conf.MaxLifeTime) * time.Second)
	// 打印sql
	db.ShowSQL(true)

	return db, nil
}
