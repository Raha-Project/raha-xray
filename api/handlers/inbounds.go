package handlers

import (
	"raha-xray/api/services"
	"raha-xray/database/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InboundHandler struct {
	BaseHandlers
	services.InboundService
	services.XrayService
	services.TrafficService
}

func NewInboundHandler(g *gin.RouterGroup) *InboundHandler {
	a := &InboundHandler{}
	a.initRouter(g)
	return a
}

func (a *InboundHandler) initRouter(gr *gin.RouterGroup) {
	g := gr.Group("/inbounds")
	g.Use(a.checkLogin)

	g.GET("/", a.getAll)
	g.GET("/get/:id", a.get)
	g.POST("/save", a.save)
	g.POST("/del/:id", a.del)
	g.GET("/traffics/:tag", a.traffics)
}

func (a *InboundHandler) getAll(c *gin.Context) {
	inbounds, err := a.InboundService.GetAll()
	if err != nil {
		jsonMsg(c, "Error in getting all inboudnds:", err)
		return
	}
	jsonObj(c, inbounds, nil)
}

func (a *InboundHandler) get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in getting inbound:", err)
		return
	}
	inbound, err := a.InboundService.Get(id)
	if err != nil {
		jsonMsg(c, "Error in finding inbound:", err)
		return
	}
	jsonObj(c, inbound, nil)
}

func (a *InboundHandler) save(c *gin.Context) {
	inbound := &model.Inbound{}
	err := c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, "Error in saving inbound:", err)
		return
	}
	err, needRestart := a.InboundService.Save(inbound)
	if err != nil {
		jsonMsg(c, "Error in saving inbound:", err)
		return
	}
	a.XrayService.WriteConfigFile(needRestart)
}

func (a *InboundHandler) del(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in deleting inbound:", err)
		return
	}
	err, needRestart := a.InboundService.Del(uint(id))
	if err != nil {
		jsonMsg(c, "Error in deleting inbound:", err)
		return
	}
	a.XrayService.WriteConfigFile(needRestart)
}

func (a *InboundHandler) traffics(c *gin.Context) {
	traffics, err := a.TrafficService.GetTraffics(c.Param("tag"))
	if err != nil {
		jsonMsg(c, "Error in grtting inbound traffics:", err)
		return
	}
	jsonObj(c, traffics, nil)
}
