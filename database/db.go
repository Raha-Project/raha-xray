package database

import (
	"os"
	"path"
	"raha-xray/config"
	"raha-xray/database/model"
	"raha-xray/util/random"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func InitDB() error {
	var err error
	var gormLogger logger.Interface

	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	settings := config.GetSettings()

	c := &gorm.Config{
		Logger: gormLogger,
	}

	if settings.DbType == "mysql" {
		db, err = gorm.Open(mysql.Open(settings.GetMysqlDsn()), c)
	} else {
		dbPath := settings.GetDBPath()
		err = os.MkdirAll(path.Dir(dbPath), 0750)
		if err != nil {
			return err
		}
		db, err = gorm.Open(sqlite.Open(dbPath), c)
	}
	if err != nil {
		return err
	}

	err = db.AutoMigrate(
		&model.Config{},
		&model.Inbound{},
		&model.Client{},
		&model.ClientInbound{},
		&model.Traffic{},
		&model.Outbound{},
		&model.Rule{},
		&model.User{})
	if err != nil {
		return err
	}

	// Init user
	var count int64
	err = db.Model(&model.User{}).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		user := &model.User{
			Key: random.Seq(32),
		}
		return db.Create(user).Error
	}

	return nil
}

func GetDB() *gorm.DB {
	return db
}

func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
