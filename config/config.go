package config

import (
	_ "embed"
	"os"
	"strings"
)

//go:embed version
var version string

//go:embed name
var name string

//go:embed config.json
var xrayTemplateConfig string

type LogLevel string

const (
	Debug LogLevel = "debug"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
)

func GetVersion() string {
	return strings.TrimSpace(version)
}

func GetName() string {
	return strings.TrimSpace(name)
}

func GetLogLevel() LogLevel {
	if IsDebug() {
		return Debug
	}
	logLevel := os.Getenv("RAHA_LOG_LEVEL")
	if logLevel == "" {
		return Info
	}
	return LogLevel(logLevel)
}

func IsDebug() bool {
	return os.Getenv("RAHA_DEBUG") == "true"
}

func GetXrayFolderPath() string {
	binFolderPath := os.Getenv("RAHA_XRAY_FOLDER")
	if binFolderPath == "" {
		binFolderPath = "xray-core"
	}
	return binFolderPath
}

func GetDefaultXrayConfig() string {
	return xrayTemplateConfig
}

func GetEnvApi() string {
	return os.Getenv("RAHA_XRAY_API")
}
