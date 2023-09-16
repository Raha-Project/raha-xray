package handlers

import (
	"raha-xray/api/services"
	"raha-xray/database/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RuleHandler struct {
	BaseHandlers
	services.RuleService
	services.XrayService
}

func NewRuleHandler(g *gin.RouterGroup) *RuleHandler {
	a := &RuleHandler{}
	a.initRouter(g)
	return a
}

func (a *RuleHandler) initRouter(gr *gin.RouterGroup) {
	g := gr.Group("/rules")
	g.Use(a.checkLogin)

	g.GET("/", a.getAll)
	g.GET("/get/:id", a.get)
	g.POST("/save", a.save)
	g.POST("/del/:id", a.del)
}

func (a *RuleHandler) getAll(c *gin.Context) {
	rules, err := a.RuleService.GetAll()
	if err != nil {
		jsonMsg(c, "Error in getting all inboudnds:", err)
		return
	}
	jsonObj(c, rules, nil)
}

func (a *RuleHandler) get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in getting rule:", err)
		return
	}
	rule, err := a.RuleService.Get(id)
	if err != nil {
		jsonMsg(c, "Error in finding rule:", err)
		return
	}
	jsonObj(c, rule, nil)
}

func (a *RuleHandler) save(c *gin.Context) {
	rule := &model.Rule{}
	err := c.ShouldBind(rule)
	if err != nil {
		jsonMsg(c, "Error in saving rule:", err)
		return
	}
	err = a.RuleService.Save(rule)
	if err != nil {
		jsonMsg(c, "Error in saving rule:", err)
		return
	}
	a.XrayService.WriteConfigFile(true)
}

func (a *RuleHandler) del(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Error in deleting rule:", err)
		return
	}
	err = a.RuleService.Del(uint(id))
	if err != nil {
		jsonMsg(c, "Error in deleting rule:", err)
		return
	}
	a.XrayService.WriteConfigFile(true)
}
