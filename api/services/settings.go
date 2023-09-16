package services

import (
	"encoding/json"
	"io/fs"
	"os"
	"raha-xray/config"
	"raha-xray/logger"
	"raha-xray/xray"
	"syscall"
	"time"
)

var xrayDefault string

type SettingService struct {
}

func (s *SettingService) SaveSettings(data *config.Setting) error {
	err := data.CheckValid()
	if err != nil {
		return err
	}
	newSettings, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("raha-xray.json", newSettings, fs.ModePerm)
}

func (s *SettingService) GetSettings() *config.Setting {
	return config.GetSettings()
}

func (s *SettingService) LoadXrayDefaults() error {
	if _, err := os.Stat("xrayDefault.json"); err == nil {
		data, err := os.ReadFile("xrayDefault.json")
		if err != nil {
			return err
		}
		xrayDefault = string(data)
	} else {
		xrayDefault = config.GetDefaultXrayConfig()
		// listen to all port in docker
		if os.Getppid() == 0 {
			xrayConfig := xray.Config{}
			err = json.Unmarshal([]byte(xrayDefault), &xrayConfig)
			if err != nil {
				return err
			}
			for index, inbound := range xrayConfig.InboundConfigs {
				if inbound.Tag == "api" {
					xrayConfig.InboundConfigs[index].Listen = nil
					break
				}
			}
			newXrayDefault, err := json.MarshalIndent(xrayConfig, "", "  ")
			if err != nil {
				return err
			}
			xrayDefault = string(newXrayDefault)
		}
		s.SaveXrayDefault(xrayDefault)
	}
	return nil
}

func (s *SettingService) SaveXrayDefault(data string) error {
	return os.WriteFile("xrayDefault.json", []byte(data), fs.ModePerm)
}

func (s *SettingService) GetXrayDefault() (*xray.Config, error) {
	xrayConfig := &xray.Config{}
	err := json.Unmarshal([]byte(xrayDefault), xrayConfig)
	if err != nil {
		logger.Error("Error in loading config")
		return nil, err
	}
	return xrayConfig, nil
}

func (s *SettingService) RestartApp(delay time.Duration) error {
	p, err := os.FindProcess(syscall.Getpid())
	if err != nil {
		return err
	}
	go func() {
		time.Sleep(delay)
		err := p.Signal(syscall.SIGHUP)
		if err != nil {
			logger.Error("send signal SIGHUP failed:", err)
		}
	}()
	return nil
}
