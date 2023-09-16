package services

import (
	"encoding/json"
	"raha-xray/database"
	"raha-xray/database/model"
	"raha-xray/logger"
	"raha-xray/util/json_util"
	"raha-xray/xray"

	"gorm.io/gorm"
)

type InboundService struct {
	xray.XrayAPI
}

func (s *InboundService) GetAll() ([]*model.Inbound, error) {
	db := database.GetDB()
	var inbounds []*model.Inbound
	err := db.Preload("Config").Preload("ClientInbounds").Find(&inbounds).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return inbounds, nil
}

func (s *InboundService) Get(id int) (*model.Inbound, error) {
	db := database.GetDB()
	var inbound *model.Inbound
	err := db.Preload("Config").Preload("ClientInbounds").Where("id = ?", id).Find(&inbound).Error
	if err != nil {
		return nil, err
	}
	if inbound.Id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return inbound, nil
}

func (s *InboundService) Save(inbound *model.Inbound) (error, bool) {
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

	// Remove old inbound with API
	if inbound.Id > 0 {
		err1 = s.XrayAPI.Init(p.GetAPIServer())
		if err1 == nil {
			var inboundTag string
			err1 = tx.Model(model.Inbound{}).Select("tag").Where("id = ?", inbound.Id).Find(&inboundTag).Error
			if err1 != nil {
				logger.Debug("Failed to find inbound data to removing by API:", err1)
				needRestart = true
			} else {
				err1 = s.XrayAPI.DelInbound(inboundTag)
				if err1 == nil {
					logger.Debug("Inbound deleted by api:", inboundTag)
				} else {
					logger.Debug("Unable to delete inbound by api:", err1)
					needRestart = true
				}
			}
		}
		s.XrayAPI.Close()
	}

	err = tx.Save(inbound).Error
	if err != nil {
		return err, false
	}

	if !needRestart {
		err1 = s.RebuildByApi(tx, []uint{inbound.Id}, false)
		if err1 != nil {
			logger.Debug("Unable to rebuild inbound by api:", err1)
			needRestart = true
		}
	}

	return nil, needRestart
}

func (s *InboundService) RebuildByApi(tx *gorm.DB, ids []uint, delFirst bool) error {
	var err error
	var inbounds []*model.Inbound

	err = tx.Model(model.Inbound{}).
		Preload("Config").
		Preload("ClientInbounds", "client_id NOT IN (select Id from clients where enable=false)").
		Where("enable = true and id in ?", ids).Find(&inbounds).Error
	if err != nil {
		return err
	}
	err = s.XrayAPI.Init(p.GetAPIServer())
	defer s.XrayAPI.Close()
	if err != nil {
		return err
	}
	for _, inbound := range inbounds {
		if delFirst {
			err = s.XrayAPI.DelInbound(inbound.Tag)
			if err == nil {
				logger.Debug("Inbound deleted by api:", inbound.Tag)
			} else {
				return err
			}
		}
		inboundConfig, err := s.GetInboundConfig(inbound)
		if err != nil {
			return err
		}
		data, err := json.Marshal(inboundConfig)
		if err != nil {
			return err
		}
		err = s.XrayAPI.AddInbound(data)
		if err == nil {
			logger.Debug("Inbound rebuilded by api:", inboundConfig.Tag)
		} else {
			return err
		}
	}
	return nil
}

func (s *InboundService) GetInboundConfig(inbound *model.Inbound) (*xray.InboundConfig, error) {
	var err error

	// Add clients object to settings
	var settings map[string]interface{}
	err = json.Unmarshal([]byte(inbound.Config.Settings), &settings)
	if err != nil {
		return nil, err
	}
	var clientInbounds []interface{}
	for _, clientInbound := range inbound.ClientInbounds {
		var clientConfigs interface{}
		json.Unmarshal([]byte(clientInbound.Config), &clientConfigs)
		clientInbounds = append(clientInbounds, clientConfigs)
	}
	settings["clients"] = clientInbounds
	settingsConfig, _ := json.Marshal(settings)
	inbound.Config.Settings = string(settingsConfig)

	inboundConfig := xray.InboundConfig{
		Listen:         json_util.RawMessage(inbound.Listen),
		Port:           int(inbound.Port),
		Protocol:       string(inbound.Config.Protocol),
		Settings:       json_util.RawMessage(inbound.Config.Settings),
		StreamSettings: json_util.RawMessage(inbound.Config.StreamSettings),
		Tag:            inbound.Tag,
		Sniffing:       json_util.RawMessage(inbound.Config.Sniffing),
	}

	return &inboundConfig, nil
}

func (s *InboundService) GetXrayInboundConfigs() ([]xray.InboundConfig, error) {
	var err error
	var inbounds []*model.Inbound
	var inboundConfigs []xray.InboundConfig

	db := database.GetDB()
	err = db.Model(model.Inbound{}).
		Preload("Config").
		Preload("ClientInbounds", "client_id NOT IN (select Id from clients where enable=false)").
		Where("enable = true").Find(&inbounds).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	for _, inbound := range inbounds {
		inboundConfig, err := s.GetInboundConfig(inbound)
		if err != nil {
			return nil, err
		}
		inboundConfigs = append(inboundConfigs, *inboundConfig)
	}

	return inboundConfigs, nil
}

func (s *InboundService) Del(id uint) (error, bool) {
	var err, err1 error
	db := database.GetDB()
	tx := db.Begin()
	defer func() {
		s.XrayAPI.Close()
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	needRestart := false
	err1 = s.XrayAPI.Init(p.GetAPIServer())
	if err1 != nil {
		needRestart = true
	} else {
		var inboundTag string
		err1 = tx.Model(model.Inbound{}).Select("tag").Where("id = ?", id).Scan(&inboundTag).Error
		if err1 != nil {
			needRestart = true
		} else {
			err1 = s.XrayAPI.DelInbound(inboundTag)
			if err1 == nil {
				logger.Debug("Inbound deleted by api:", inboundTag)
			} else {
				logger.Debug("Unable to delete inbound by api:", err1)
				needRestart = true
			}
		}
	}

	// Delete related ClientInbouds
	err = tx.Where("inbound_id = ?", id).Delete(model.ClientInbound{}).Error
	if err != nil {
		return err, needRestart
	}

	err = tx.Delete(model.Inbound{}, id).Error
	if err != nil {
		return err, needRestart
	}

	return nil, needRestart
}
