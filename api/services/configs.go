package services

import (
	"errors"
	"raha-xray/database"
	"raha-xray/database/model"
	"raha-xray/logger"

	"gorm.io/gorm"
)

type ConfigService struct {
	InboundService
}

func (s *ConfigService) GetAll() ([]*model.Config, error) {
	db := database.GetDB()
	var configs []*model.Config
	err := db.Model(model.Config{}).Find(&configs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return configs, nil
}

func (s *ConfigService) Get(id int) (*model.Config, error) {
	db := database.GetDB()
	var config *model.Config
	err := db.Model(model.Config{}).Where("id = ?", id).Find(&config).Error
	if err != nil {
		return nil, err
	}
	if config.Id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return config, nil
}

func (s *ConfigService) Save(config *model.Config) (error, bool) {
	var err, err1 error
	db := database.GetDB()
	tx := db.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	needRestart := false
	if config.Id > 0 {
		var inboundIds []uint
		err1 = tx.Table("inbounds").Select("id").Where("config_id = ?", config.Id).Scan(&inboundIds).Error
		if err1 != nil {
			logger.Debug("Failed to load inbounds after config change.", err1)
			needRestart = true
		}
		if len(inboundIds) > 0 {
			err1 = s.InboundService.RebuildByApi(tx, inboundIds, true)
			if err1 != nil {
				logger.Debug("Failed to rebuild inbounds after config change by API:", err1)
				needRestart = true
			}
		}
	}

	err = tx.Save(config).Error
	if err != nil {
		return err, needRestart
	}

	return nil, needRestart
}

func (s *ConfigService) Del(id uint) error {
	db := database.GetDB()

	// Check if inbound using this config
	var count int64
	result := db.Model(&model.Inbound{}).Where("config_id = ?", id).Count(&count)
	if result.Error != nil {
		return result.Error
	}
	if count > 0 {
		return errors.New("config is in use by some inbounds")
	}
	return db.Delete(model.Config{}, id).Error
}
