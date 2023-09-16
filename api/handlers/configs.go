package handlers

import (
	"raha-xray/api/services"
	"raha-xray/database/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	BaseHandlers
	services.ConfigService
	services.XrayService
}

func NewConfigHandler(g *gin.RouterGroup) *ConfigHandler {
	a := &ConfigHandler{}
	a.initRouter(g)
	return a
}

func (a *ConfigHandler) initRouter(gr *gin.RouterGroup) {
	g := gr.Group("/configs")
	g.Use(a.checkLogin)

	g.GET("/", a.getAll)
	g.GET("/get/:id", a.get)
	g.POST("/save", a.save)
	g.POST("/del/:id", a.del)
}

func (a *ConfigHandler) getAll(c *gin.Context) {
	configs, err := a.ConfigService.GetAll()
	if err != nil {
		jsonMsg(c, "Error in getting all inboudnds:", err)
		return
	}
	jsonObj(c, configs, nil)
}

func (a *ConfigHandler) get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in getting config:", err)
		return
	}
	config, err := a.ConfigService.Get(id)
	if err != nil {
		jsonMsg(c, "Error in finding config:", err)
		return
	}
	jsonObj(c, config, nil)
}

func (a *ConfigHandler) save(c *gin.Context) {
	config := &model.Config{}
	err := c.ShouldBind(config)
	if err != nil {
		jsonMsg(c, "Error in saving config:", err)
		return
	}
	err, needRestart := a.ConfigService.Save(config)
	if err != nil {
		jsonMsg(c, "Error in saving config:", err)
		return
	}
	a.XrayService.WriteConfigFile(needRestart)
}

func (a *ConfigHandler) del(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in deleting config:", err)
		return
	}
	err = a.ConfigService.Del(uint(id))
	if err != nil {
		jsonMsg(c, "Error in deleting config:", err)
		return
	}
}
