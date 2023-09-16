package services

import (
	"encoding/json"
	"raha-xray/database"
	"raha-xray/database/model"
	"raha-xray/logger"
	"raha-xray/xray"

	"gorm.io/gorm"
)

type ClientService struct {
	xray.XrayAPI
	InboundService
}

func (s *ClientService) GetAll() ([]*model.Client, error) {
	db := database.GetDB()
	var clients []*model.Client
	err := db.Model(model.Client{}).Preload("ClientInbounds").Find(&clients).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return clients, nil
}

func (s *ClientService) Get(id uint) (*model.Client, error) {
	db := database.GetDB()
	client := &model.Client{}
	err := db.Model(model.Client{}).Preload("ClientInbounds").Where("id = ?", id).Find(client).Error
	if err != nil {
		return nil, err
	}
	if client.Id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return client, nil
}

func (s *ClientService) Add(clients []*model.Client) (error, bool) {
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
	}

	// Add clients to inbounds by API
	for _, client := range clients {
		for index, clientInbound := range client.ClientInbounds {
			var inbound model.Inbound
			err1 = tx.Model(model.Inbound{}).Preload("Config").Where("id = ?", clientInbound.InboundId).Find(&inbound).Error
			if err1 != nil {
				logger.Debug("Failed to find inbound data for addinf client by API:", err1)
				needRestart = true
			} else {
				// Push email to config
				var clientConfig map[string]interface{}
				json.Unmarshal([]byte(clientInbound.Config), &clientConfig)
				clientConfig["email"] = client.Name
				newClientConfig, _ := json.MarshalIndent(clientConfig, "", "  ")
				client.ClientInbounds[index].Config = string(newClientConfig)

				if !needRestart && client.Enable {
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

	err = tx.Create(clients).Error
	if err != nil {
		return err, needRestart
	}

	return nil, needRestart
}

func (s *ClientService) Update(client *model.Client) (error, bool) {
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
	}

	var oldClient model.Client
	err = tx.Model(model.Client{}).Where("id = ?", client.Id).Preload("ClientInbounds").Find(&oldClient).Error
	if err != nil {
		return err, false
	}

	// Remove all clients from xray
	if !needRestart && oldClient.Enable {
		needRestart = s.apiRemoveClients(tx, client.Id)
	}

	for index, clientInbound := range oldClient.ClientInbounds {
		var inbound model.Inbound
		err1 = tx.Model(model.Inbound{}).Preload("Config").Where("id = ?", clientInbound.InboundId).Find(&inbound).Error
		if err1 != nil {
			logger.Debug("Failed to find inbound data for add client by API:", err1)
			needRestart = true
		} else {
			// Add client ID
			clientInbound.ClientId = client.Id

			// Push email to config
			var clientConfig map[string]interface{}
			json.Unmarshal([]byte(clientInbound.Config), &clientConfig)
			clientConfig["email"] = client.Name
			newClientConfig, _ := json.MarshalIndent(clientConfig, "", "  ")
			oldClient.ClientInbounds[index].Config = string(newClientConfig)

			if !needRestart && client.Enable {
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
	// Update ClientInbounds due to changes client
	err = tx.Save(oldClient.ClientInbounds).Error
	if err != nil {
		return err, true // Restart on unsuccessfull update
	}

	newClient := model.Client{
		Id:     client.Id,
		Name:   client.Name,
		Enable: client.Enable,
		Quota:  client.Quota,
		Expiry: client.Expiry,
		Reset:  client.Reset,
		Once:   client.Once,
		Up:     client.Up,
		Down:   client.Down,
		Remark: client.Remark,
	}

	// Update client
	err = tx.Save(newClient).Error
	if err != nil {
		return err, true // Restart on unsuccessfull update
	}

	return nil, needRestart
}

func (s *ClientService) Inbounds(id int, clientInbounds []*model.ClientInbound) (error, bool) {
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

	client := &model.Client{}
	err = db.Model(model.Client{}).Preload("ClientInbounds").Where("id = ?", id).Find(client).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err, false
	}

	needRestart := false
	err1 = s.XrayAPI.Init(p.GetAPIServer())
	if err1 != nil {
		needRestart = false
	}
	// Remove all clients from xray
	if !needRestart && client.Enable {
		needRestart = s.apiRemoveClients(tx, client.Id)
	}

	var clientInboundIds []uint

	for _, clientInbound := range clientInbounds {
		if clientInbound.Id > 0 {
			clientInboundIds = append(clientInboundIds, clientInbound.Id)
		}

		// Add client id
		clientInbound.ClientId = client.Id

		// Push email to config
		var clientConfig map[string]interface{}
		json.Unmarshal([]byte(clientInbound.Config), &clientConfig)
		clientConfig["email"] = client.Name
		newClientConfig, _ := json.MarshalIndent(clientConfig, "", "  ")
		clientInbound.Config = string(newClientConfig)

		if !needRestart && client.Enable {
			var inbound *model.Inbound
			err := tx.Preload("Config").Where("id = ?", clientInbound.InboundId).Find(&inbound).Error
			if err != nil {
				return err, false
			}

			// Call API to add new config
			err1 = s.XrayAPI.AddUser(string(inbound.Config.Protocol), inbound.Tag, clientConfig)
			if err1 == nil {
				logger.Debug("Client added by api:", client.Name)
			} else {
				logger.Debug("Failed to adding client by api:", err1)
				needRestart = true
			}
		}
	}

	// Remove ommited ClientInbounds
	err = tx.Where("id not in ? and client_id=?", clientInboundIds, client.Id).Delete(model.ClientInbound{}).Error
	if err != nil {
		return err, needRestart
	}

	// Save changes in database
	for _, clientInbound := range clientInbounds {
		if clientInbound.Id > 0 {
			err = tx.Save(clientInbound).Error
			if err != nil {
				return err, needRestart
			}
		} else {
			err = tx.Create(clientInbound).Error
			if err != nil {
				return err, needRestart
			}
		}
	}

	return nil, needRestart
}

func (s *ClientService) apiRemoveClients(tx *gorm.DB, client_id uint) bool {
	var err error
	var clientTags []struct {
		InboundTag string
		Email      string
	}
	err = tx.Raw(`
		SELECT inbounds.tag as inbound_tag, clientInbound.email as email
		FROM inbounds
		JOIN (
			SELECT client_inbounds.inbound_id, client.name AS email
			FROM client_inbounds
			JOIN (
				SELECT id, name
				FROM clients
				WHERE enable = ? AND id = ?
				) AS client ON client_inbounds.client_id = client.id
			) AS clientInbound ON inbounds.id = clientInbound.inbound_id;
		`, true, client_id).Scan(&clientTags).Error
	if err != nil {
		logger.Debug("Failed to find client data for removing by API:", err)
		return true
	} else {
		for _, clientTag := range clientTags {
			err := s.XrayAPI.DelUser(clientTag.InboundTag, clientTag.Email)
			if err == nil {
				logger.Debug("Client removed by api:", clientTag.Email)
			} else {
				logger.Debug("Failed to removing client by api:", err)
				return true
			}
		}
	}
	return false
}

func (s *ClientService) Del(id uint) (error, bool) {
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
		needRestart = false
	}
	// Remove all clients from xray
	if !needRestart {
		needRestart = s.apiRemoveClients(tx, id)
	}

	// Delete related ClientInbouds
	err = tx.Where("client_id = ?", id).Delete(model.ClientInbound{}).Error
	if err != nil {
		return err, needRestart
	}
	err = tx.Delete(model.Client{}, id).Error
	if err != nil {
		return err, needRestart
	}
	return nil, needRestart
}

func (s *ClientService) GetOnlineClinets() []string {
	return p.GetOnlineClients()
}
