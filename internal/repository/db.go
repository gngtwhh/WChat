package repository

import (
    "fmt"
    "wchat/internal/config"
    "wchat/internal/model"
    "wchat/pkg/zlog"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

var gormDB *gorm.DB

func InitDB() (*gorm.DB, error) {
    conf := config.GetConfig()

    user := conf.MysqlConfig.User
    password := conf.MysqlConfig.Password
    host := conf.MysqlConfig.Host
    port := conf.MysqlConfig.Port
    appName := conf.AppName

    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, appName,
    )
    var err error
    gormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        zlog.Fatal(err.Error())
        return nil, err
    }

    err = gormDB.AutoMigrate(
        &model.User{},
        &model.Message{},
        &model.Contact{},
        &model.ContactApply{},
        &model.Group{},
        &model.Session{},
    )
    if err != nil {
        zlog.Fatal(err.Error())
        return nil, err
    }
    return gormDB, nil
}
