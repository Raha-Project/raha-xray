package handlers

import (
	"raha-xray/api/services"
	"raha-xray/database/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OutboundHandler struct {
	BaseHandlers
	services.OutboundService
	services.XrayService
	services.TrafficService
}

func NewOutboundHandler(g *gin.RouterGroup) *OutboundHandler {
	a := &OutboundHandler{}
	a.initRouter(g)
	return a
}

func (a *OutboundHandler) initRouter(gr *gin.RouterGroup) {
	g := gr.Group("/outbounds")
	g.Use(a.checkLogin)

	g.GET("/", a.getAll)
	g.GET("/get/:id", a.get)
	g.POST("/save", a.save)
	g.POST("/del/:id", a.del)
	g.GET("/traffics/:tag", a.traffics)

}

func (a *OutboundHandler) getAll(c *gin.Context) {
	outbounds, err := a.OutboundService.GetAll()
	if err != nil {
		jsonMsg(c, "Error in getting all inboudnds:", err)
		return
	}
	jsonObj(c, outbounds, nil)
}

func (a *OutboundHandler) get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in getting outbound:", err)
		return
	}
	outbound, err := a.OutboundService.Get(id)
	if err != nil {
		jsonMsg(c, "Error in finding outbound:", err)
		return
	}
	jsonObj(c, outbound, nil)
}

func (a *OutboundHandler) save(c *gin.Context) {
	outbound := &model.Outbound{}
	err := c.ShouldBind(outbound)
	if err != nil {
		jsonMsg(c, "Error in saving outbound:", err)
		return
	}
	err, needRestart := a.OutboundService.Save(outbound)
	if err != nil {
		jsonMsg(c, "Error in saving outbound:", err)
		return
	}
	a.XrayService.WriteConfigFile(needRestart)
}

func (a *OutboundHandler) del(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in deleting outbound:", err)
		return
	}
	err, needRestart := a.OutboundService.Del(uint(id))
	if err != nil {
		jsonMsg(c, "Error in deleting outbound:", err)
		return
	}
	a.XrayService.WriteConfigFile(needRestart)
}

func (a *OutboundHandler) traffics(c *gin.Context) {
	traffics, err := a.TrafficService.GetTraffics("outbound", c.Param("tag"))
	if err != nil {
		jsonMsg(c, "Error in grtting outbound traffics:", err)
		return
	}
	jsonObj(c, traffics, nil)
}
