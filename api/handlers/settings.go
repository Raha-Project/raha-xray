package handlers

import (
	"raha-xray/api/services"
	"raha-xray/config"
	"time"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	BaseHandlers
	services.SettingService
}

func NewSettingsHandler(g *gin.RouterGroup) *SettingsHandler {
	a := &SettingsHandler{}
	a.initRouter(g)
	return a
}

func (a *SettingsHandler) initRouter(g *gin.RouterGroup) {
	g = g.Group("/settings")
	g.Use(a.checkLogin)

	g.POST("/getXrayDefault", a.getXrayDefault)
	g.POST("/setXrayDefault", a.setXrayDefault)
	g.POST("/getSettings", a.getSettings)
	g.POST("/setSettings", a.setSettings)
	g.POST("/restartApp", a.restartApp)
}

func (a *SettingsHandler) getXrayDefault(c *gin.Context) {
	xrayDefault, err := a.SettingService.GetXrayDefault()
	if err != nil {
		jsonMsg(c, "Get xray default config", err)
		return
	}
	jsonObj(c, xrayDefault, nil)
}

func (a *SettingsHandler) setXrayDefault(c *gin.Context) {
	xrayDefault, err := c.GetRawData()
	if err != nil {
		jsonMsg(c, "Receive xray default config", err)
		return
	}

	err = a.SettingService.SaveXrayDefault(string(xrayDefault))
	if err != nil {
		jsonMsg(c, "Save xray default config file", err)
		return
	}
	a.SettingService.RestartApp(time.Second * 3)
	jsonMsg(c, "Default xray configuration saved", nil)
}

func (a *SettingsHandler) getSettings(c *gin.Context) {
	jsonObj(c, a.SettingService.GetSettings(), nil)
}

func (a *SettingsHandler) setSettings(c *gin.Context) {
	settingJSON := &config.Setting{}
	err := c.ShouldBind(settingJSON)
	if err != nil {
		jsonMsg(c, "Receive app settings", err)
		return
	}

	err = a.SettingService.SaveSettings(settingJSON)
	if err != nil {
		jsonMsg(c, "Save app settings file", err)
		return
	}
	a.SettingService.RestartApp(time.Second * 3)
	jsonMsg(c, "App settings saved", nil)
}

func (a *SettingsHandler) restartApp(c *gin.Context) {
	err := a.SettingService.RestartApp(time.Second * 3)
	jsonMsg(c, "Restart ordered", err)
}
