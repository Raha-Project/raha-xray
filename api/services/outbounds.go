package services

import (
	"encoding/json"
	"raha-xray/database"
	"raha-xray/database/model"
	"raha-xray/logger"
	"raha-xray/util/common"
	"raha-xray/util/json_util"
	"raha-xray/xray"

	"gorm.io/gorm"
)

type OutboundService struct {
	xray.XrayAPI
}

func (s *OutboundService) GetAll() ([]*model.Outbound, error) {
	db := database.GetDB()
	var outbounds []*model.Outbound
	err := db.Model(model.Outbound{}).Find(&outbounds).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return outbounds, nil
}

func (s *OutboundService) Get(id int) (*model.Outbound, error) {
	db := database.GetDB()
	var outbound *model.Outbound
	err := db.Model(model.Outbound{}).Where("id = ?", id).Find(&outbound).Error
	if err != nil {
		return nil, err
	}
	if outbound.Id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return outbound, nil
}

func (s *OutboundService) Save(outbound *model.Outbound) (error, bool) {
	var err error
	db := database.GetDB()

	needRestart := false
	err = s.XrayAPI.Init(p.GetAPIServer())
	if err != nil {
		needRestart = true
	}
	defer s.XrayAPI.Close()

	// Remove old outbound with API
	if outbound.Id > 0 && !needRestart {
		var outboundTag string
		err = db.Model(model.Outbound{}).Select("tag").Where("id = ?", outbound.Id).Find(&outboundTag).Error
		if err != nil {
			logger.Debug("Failed to find outbound data to removing by API:", err)
			needRestart = true
		} else {
			err = s.XrayAPI.DelOutbound(outboundTag)
			if err == nil {
				logger.Debug("Outbound deleted by api:", outboundTag)
			} else {
				logger.Debug("Unable to delete outbound by api:", err)
				needRestart = true
			}
		}
	}

	if !needRestart {
		outboundConfig, err := s.GetOutboundConfig(outbound)
		if err != nil {
			needRestart = true
		}
		data, err := json.Marshal(outboundConfig)
		if err != nil {
			needRestart = true
		}
		err = s.XrayAPI.AddOutbound(data)
		if err == nil {
			logger.Debug("Outbound added by api:", outbound.Tag)
		} else {
			needRestart = true
		}
	}

	return db.Save(outbound).Error, needRestart
}

func (s *OutboundService) GetOutboundConfig(outbound *model.Outbound) (*json_util.RawMessage, error) {
	outboundConfig := make(map[string]interface{})
	outboundConfig["protocol"] = outbound.Protocol
	outboundConfig["tag"] = outbound.Tag
	if common.NonEmptyValue(outbound.SendThrough) {
		outboundConfig["sendThrough"] = outbound.SendThrough
	}
	if common.NonEmptyValue(outbound.Settings) {
		var settings interface{}
		json.Unmarshal([]byte(outbound.Settings), &settings)
		outboundConfig["settings"] = settings
	}
	if common.NonEmptyValue(outbound.StreamSettings) {
		var streamSettings interface{}
		json.Unmarshal([]byte(outbound.StreamSettings), &streamSettings)
		outboundConfig["streamSettings"] = streamSettings
	}
	if common.NonEmptyValue(outbound.ProxySettings) {
		var proxySettings interface{}
		json.Unmarshal([]byte(outbound.ProxySettings), &proxySettings)
		outboundConfig["proxySettings"] = proxySettings
	}
	if common.NonEmptyValue(outbound.Mux) {
		var mux interface{}
		json.Unmarshal([]byte(outbound.Mux), &mux)
		outboundConfig["mux"] = mux
	}

	outboundJSON, err := json.Marshal(outboundConfig)
	if err != nil {
		return nil, err
	}
	outboundRaw := json_util.RawMessage(outboundJSON)
	return &outboundRaw, nil
}

func (s *OutboundService) Del(id uint) (error, bool) {
	var err error
	db := database.GetDB()

	needRestart := false
	err = s.XrayAPI.Init(p.GetAPIServer())
	defer s.XrayAPI.Close()
	if err != nil {
		needRestart = true
	} else {
		var outboundTag string
		err = db.Model(model.Outbound{}).Select("tag").Where("id = ?", id).Scan(&outboundTag).Error
		if err != nil {
			needRestart = true
		} else {
			err = s.XrayAPI.DelOutbound(outboundTag)
			if err == nil {
				logger.Debug("Outbound deleted by api:", outboundTag)
			} else {
				logger.Debug("Unable to delete outboundTag by api:", err)
				needRestart = true
			}
		}
	}

	return db.Delete(model.Outbound{}, id).Error, needRestart
}
