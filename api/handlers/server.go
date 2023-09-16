package handlers

import (
	"raha-xray/api/services"
	"time"

	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	BaseHandlers

	services.ServerService
	services.XrayService

	lastStatus        *services.Status
	lastGetStatusTime time.Time

	lastVersions        []string
	lastGetVersionsTime time.Time
}

func NewServerHandler(g *gin.RouterGroup) *ServerHandler {
	a := &ServerHandler{
		lastGetStatusTime: time.Now(),
	}
	a.initRouter(g)
	return a
}

func (a *ServerHandler) initRouter(g *gin.RouterGroup) {
	g = g.Group("/server")
	g.Use(a.checkLogin)

	g.POST("/status", a.status)
	g.POST("/getXrayVersion", a.getXrayVersion)
	g.POST("/setXrayVersion/:version", a.setXrayVersion)
	g.POST("/stopXrayService", a.stopXrayService)
	g.POST("/restartXrayService", a.restartXrayService)
	g.POST("/getConfigJson", a.getConfigJson)
	g.POST("/logs/:app/:count", a.getLogs)
	g.POST("/getNewX25519Cert", a.getNewX25519Cert)
}

func (a *ServerHandler) status(c *gin.Context) {
	status := a.ServerService.GetStatus(a.lastStatus)

	jsonObj(c, status, nil)
}

func (a *ServerHandler) getXrayVersion(c *gin.Context) {
	now := time.Now()
	if now.Sub(a.lastGetVersionsTime) <= time.Minute {
		jsonObj(c, a.lastVersions, nil)
		return
	}

	versions, err := a.ServerService.GetXrayVersions()
	if err != nil {
		jsonMsg(c, "Get xray version", err)
		return
	}

	a.lastVersions = versions
	a.lastGetVersionsTime = time.Now()

	jsonObj(c, versions, nil)
}

func (a *ServerHandler) setXrayVersion(c *gin.Context) {
	version := c.Param("version")
	err := a.XrayService.SetXrayVersion(version)
	jsonMsg(c, "Install xray", err)
}

func (a *ServerHandler) stopXrayService(c *gin.Context) {
	err := a.XrayService.StopXray()
	if err != nil {
		jsonMsg(c, "Stoping xray", err)
		return
	}
	jsonMsg(c, "Xray stoped", err)
}

func (a *ServerHandler) restartXrayService(c *gin.Context) {
	err := a.XrayService.RestartXray()
	if err != nil {
		jsonMsg(c, "Restart xray", err)
		return
	}
	jsonMsg(c, "Xray restarted", err)
}

func (a *ServerHandler) getLogs(c *gin.Context) {
	count := c.Param("count")
	app := c.Param("app")
	logs := a.ServerService.GetLogs(count, app)
	jsonObj(c, logs, nil)
}

func (a *ServerHandler) getConfigJson(c *gin.Context) {
	configJson, err := a.ServerService.GetConfigJson()
	if err != nil {
		jsonMsg(c, "Get config.json", err)
		return
	}
	jsonObj(c, configJson, nil)
}

func (a *ServerHandler) getNewX25519Cert(c *gin.Context) {
	cert, err := a.ServerService.GetNewX25519Cert()
	if err != nil {
		jsonMsg(c, "Get x25519 certificate", err)
		return
	}
	jsonObj(c, cert, nil)
}
