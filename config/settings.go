package config

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"os"
	"raha-xray/logger"
	"raha-xray/util/common"
	"strings"
	"time"
)

var settings *Setting

type Setting struct {
	Listen       string `json:"listen" form:"listen"`
	Domain       string `json:"domain" form:"domain"`
	Port         int    `json:"port" form:"port"`
	CertFile     string `json:"certFile" form:"certFile"`
	KeyFile      string `json:"keyFile" form:"keyFile"`
	BasePath     string `json:"basePath" form:"basePath"`
	TimeLocation string `json:"timeLocation" form:"timeLocation"`
	DbType       string `json:"dbType" form:"dbType"`
	DbAddr       string `json:"dbAddr" form:"dbAddr"`
	TrafficDays  int    `json:"trafficDays" form:"trafficDays"`
}

var defaultSettings = Setting{
	Listen:       "",
	Domain:       "",
	Port:         8080,
	CertFile:     "",
	KeyFile:      "",
	BasePath:     "/api",
	TimeLocation: "Asia/Tehran",
	DbType:       "sqlite",
	DbAddr:       "db",
	TrafficDays:  0,
}

func GetDefaultSettings() *Setting {
	return &defaultSettings
}

func (s *Setting) CheckValid() error {
	if s.Listen != "" {
		ip := net.ParseIP(s.Listen)
		if ip == nil {
			return common.NewError("Listen is not valid ip:", s.Listen)
		}
	}

	if s.Port <= 0 || s.Port > 65535 {
		return common.NewError("Port is not a valid port:", s.Port)
	}

	if s.CertFile != "" || s.KeyFile != "" {
		_, err := tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
		if err != nil {
			return common.NewErrorf("cert file <%v> or key file <%v> invalid: %v", s.CertFile, s.KeyFile, err)
		}
	}

	if !strings.HasPrefix(s.BasePath, "/") {
		s.BasePath = "/" + s.BasePath
	}
	s.BasePath = strings.TrimSuffix(s.BasePath, "/")

	_, err := time.LoadLocation(s.TimeLocation)
	if err != nil {
		return common.NewError("time location not exist:", s.TimeLocation)
	}

	return nil
}

func (s *Setting) GetTimeLocation() (*time.Location, error) {
	location, err := time.LoadLocation(s.TimeLocation)
	if err != nil {
		defaultLocation := defaultSettings.TimeLocation
		logger.Errorf("location <%v> not exist, using default location: %v", s.TimeLocation, defaultLocation)
		return time.LoadLocation(defaultLocation)
	}
	return location, nil
}

func (s *Setting) GetDBPath() string {
	return fmt.Sprintf("%s/%s.db", s.DbAddr, GetName())
}

func (s *Setting) GetMysqlDsn() string {
	mysqlServer := s.DbAddr
	if mysqlServer == "" {
		mysqlServer = "root@"
	}
	return fmt.Sprintf("%s/%s?charset=utf8mb4&loc=Local", mysqlServer, GetName())
}

func LoadSettings() error {
	if _, err := os.Stat("raha-xray.json"); err == nil {
		data, err := os.ReadFile("raha-xray.json")
		if err != nil {
			return err
		}
		var newConfig Setting
		err = json.Unmarshal(data, &newConfig)
		if err != nil {
			return err
		}
		settings = &newConfig
	} else {
		settings = &defaultSettings
		SaveSettings(settings)
	}
	return nil
}

func SaveSettings(data *Setting) error {
	newSettings, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("raha-xray.json", newSettings, fs.ModePerm)
}

func GetSettings() *Setting {
	return settings
}
