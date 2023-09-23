package handlers

import (
	"raha-xray/api/services"
	"raha-xray/database/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	BaseHandlers
	services.ClientService
	services.XrayService
	services.TrafficService
}

func NewClientHandler(g *gin.RouterGroup) *ClientHandler {
	a := &ClientHandler{}
	a.initRouter(g)
	return a
}

func (a *ClientHandler) initRouter(g *gin.RouterGroup) {
	g = g.Group("/clients")
	g.Use(a.checkLogin)

	g.GET("/", a.getAll)
	g.GET("/get/:id", a.get)
	g.POST("/add", a.add)
	g.POST("/update", a.update)
	g.POST("/inbounds/:id", a.inbounds)
	g.POST("/del/:id", a.del)
	g.POST("/onlines", a.onlines)
	g.GET("/traffics/:tag", a.traffics)
}

func (a *ClientHandler) getAll(c *gin.Context) {
	clients, err := a.ClientService.GetAll()
	if err != nil {
		jsonMsg(c, "Error in getting all clients:", err)
		return
	}
	jsonObj(c, clients, nil)
}

func (a *ClientHandler) get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in getting client id:", err)
		return
	}
	client, err := a.ClientService.Get(uint(id))
	if err != nil {
		jsonMsg(c, "Error in finding client:", err)
		return
	}
	jsonObj(c, client, nil)
}

func (a *ClientHandler) add(c *gin.Context) {
	var clients []*model.Client
	err := c.ShouldBind(&clients)
	if err != nil {
		jsonMsg(c, "Error in create client:", err)
		return
	}

	err, needRestart := a.ClientService.Add(clients)
	if err != nil {
		jsonMsg(c, "Error in adding client:", err)
		return
	}
	a.XrayService.WriteConfigFile(needRestart)
}

func (a *ClientHandler) update(c *gin.Context) {
	var data map[string]interface{}
	err := c.ShouldBind(&data)
	if err != nil {
		jsonMsg(c, "Error in fetch client data:", err)
		return
	}

	err, needRestart := a.ClientService.Update(data)
	if err != nil {
		jsonMsg(c, "Error in updating client:", err)
		return
	}
	a.XrayService.WriteConfigFile(needRestart)
}

func (a *ClientHandler) inbounds(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in getting client id:", err)
		return
	}
	var clientInbounds []*model.ClientInbound
	err = c.ShouldBind(&clientInbounds)
	if err != nil {
		jsonMsg(c, "Error in updating client:", err)
		return
	}

	err, needRestart := a.ClientService.Inbounds(id, clientInbounds)
	if err != nil {
		jsonMsg(c, "Error in updating client:", err)
		return
	}
	a.XrayService.WriteConfigFile(needRestart)
}

func (a *ClientHandler) del(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in deleting client:", err)
		return
	}
	err, needRestart := a.ClientService.Del(uint(id))
	if err != nil {
		jsonMsg(c, "Error in deleting client:", err)
		return
	}
	a.XrayService.WriteConfigFile(needRestart)
}

func (a *ClientHandler) onlines(c *gin.Context) {
	jsonObj(c, a.ClientService.GetOnlineClinets(), nil)
}

func (a *ClientHandler) traffics(c *gin.Context) {
	traffics, err := a.TrafficService.GetTraffics("user", c.Param("tag"))
	if err != nil {
		jsonMsg(c, "Error in grtting client traffics:", err)
		return
	}
	jsonObj(c, traffics, nil)
}
