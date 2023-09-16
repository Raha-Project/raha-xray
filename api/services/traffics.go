package services

import (
	"encoding/json"
	"raha-xray/config"
	"raha-xray/database"
	"raha-xray/database/model"
	"raha-xray/logger"
	"raha-xray/xray"
	"time"

	"gorm.io/gorm"
)

type TrafficService struct {
	xray.XrayAPI
}

func (s *TrafficService) GetTraffics(inboundTag string) ([]*model.Traffic, error) {
	var traffics []*model.Traffic
	var err error

	db := database.GetDB()
	err = db.Model(model.Traffic{}).Where("tag = ?", inboundTag).Find(&traffics).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return traffics, nil
}

func (s *TrafficService) AddTraffic(traffics []*model.Traffic) (error, bool) {
	if len(traffics) == 0 {
		return nil, false
	}
	var err, err1 error
	db := database.GetDB()
	tx := db.Begin()

	defer func() {
		s.XrayAPI.Close()
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	needRestart := false
	err1 = s.XrayAPI.Init(p.GetAPIServer())
	if err1 != nil {
		needRestart = true
	}

	err, needRestart = s.resetClients(tx, needRestart)
	if err != nil {
		return err, needRestart
	}

	for _, traffic := range traffics {
		if traffic.Resource == "user" {
			if traffic.Direction {
				err = tx.Model(&model.Client{}).Where("name = ?", traffic.Tag).
					Updates(map[string]interface{}{
						"down": gorm.Expr("down + ?", traffic.Traffic),
					}).Error
			} else {
				err = tx.Model(&model.Client{}).Where("name = ?", traffic.Tag).
					Updates(map[string]interface{}{
						"up": gorm.Expr("up + ?", traffic.Traffic),
					}).Error
			}
			if err != nil {
				return err, false
			}
		}
	}

	now := time.Now().Unix() * 1000
	if !needRestart {
		// remove finished clients with API
		var clientTags []struct {
			InboundTag string
			Email      string
		}
		err1 = tx.Raw(`
			SELECT inbounds.tag as tag, clientInbound.email as email
			FROM inbounds
			JOIN (
				SELECT client_inbounds.inbound_id, client.name AS email
				FROM client_inbounds
				JOIN (
					SELECT id, name
					FROM clients
					WHERE enable = ?
					) AS client ON client_inbounds.client_id = client.id
				) AS clientInbound ON inbounds.id = clientInbound.inbound_id;
			`, false).Scan(&clientTags).Error
		if err1 != nil {
			logger.Debug("Failed to find finished clients:", err1)
		} else {
			for _, clientTag := range clientTags {
				err1 := s.XrayAPI.DelUser(clientTag.InboundTag, clientTag.Email)
				if err1 == nil {
					logger.Debug("Client removed by api:", clientTag.Email)
				} else {
					logger.Debug("Failed to removing client by api:", err1)
					needRestart = true
					break
				}
			}
		}
	}

	result := tx.Model(model.Client{}).
		Where("((quota > 0 and up + down >= quota) or (expiry > 0 and expiry <= ?)) and enable = ?", now, true).
		Update("enable", false)
	err = result.Error
	if err != nil {
		logger.Warning("Error in disabling invalid clients:", err)
	} else if result.RowsAffected > 0 {
		logger.Debugf("%v clients disabled", result.RowsAffected)
	}

	appConfig := config.GetSettings()

	// Store all traffics if it is enabled
	if appConfig.TrafficDays != 0 {
		err = tx.Save(traffics).Error
		if err != nil {
			return err, needRestart
		}
	}

	return nil, needRestart
}

func (s *TrafficService) resetClients(tx *gorm.DB, needRestart bool) (error, bool) {
	var clients []*model.Client
	var err, err1 error

	err = tx.Where("reset > 0 and expiry > 0 and expiry <= ?", time.Now().UnixNano()).Preload("client_inbounds").Scan(clients).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err, true
	}
	for _, client := range clients {
		if !client.Enable && !needRestart {
			client.Enable = true
			// Add client to API
			for _, clientInbound := range client.ClientInbounds {
				var inbound model.Inbound
				err1 = tx.Model(model.Inbound{}).Preload("Config").Where("id = ?", clientInbound.InboundId).Find(&inbound).Error
				if err1 != nil {
					logger.Debug("Failed to find inbound data for addinf client by API:", err1)
					needRestart = true
				} else {
					var clientConfig map[string]interface{}
					json.Unmarshal([]byte(clientInbound.Config), &clientConfig)
					clientConfig["email"] = client.Name

					if !needRestart {
						// Call API
						err1 = s.XrayAPI.AddUser(string(inbound.Config.Protocol), inbound.Tag, clientConfig)
						if err1 == nil {
							logger.Debug("Client added by api:", client.Name)
						} else {
							logger.Debug("Failed to adding client by api:", err1)
							needRestart = true
						}
					}
				}
			}
		}
		client.Up = 0
		client.Down = 0
		client.Expiry = uint64(time.Now().AddDate(0, 0, int(client.Reset)).UnixNano())
	}

	return tx.Save(clients).Error, needRestart
}

func (s *TrafficService) DelOldTraffics() int64 {
	appConfig := config.GetSettings()

	if appConfig.TrafficDays > 0 {
		db := database.GetDB()
		dateTimeThreshold := time.Now().AddDate(0, 0, -appConfig.TrafficDays).UnixNano()
		result := db.Where("date_time < ?", dateTimeThreshold).Delete(model.Traffic{})
		if result.Error != nil {
			logger.Debug("Unable to delete old traffics", result.Error)
		} else {
			return result.RowsAffected
		}
	}
	return 0
}
