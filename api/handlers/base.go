package handlers

import (
	"raha-xray/database"
	"raha-xray/database/model"

	"github.com/gin-gonic/gin"
)

type BaseHandlers struct {
}

func (a *BaseHandlers) checkLogin(c *gin.Context) {
	apikey := c.GetHeader("X-Token")
	if apikey == "" {
		a.abort(c)
		return
	}
	db := database.GetDB()
	var count int64
	result := db.Model(&model.User{}).Where("`key` = ?", apikey).Count(&count)
	if result.Error != nil || count == 0 {
		a.abort(c)
		return
	}

	c.Next()
}

func (a *BaseHandlers) abort(c *gin.Context) {
	pureJsonMsg(c, false, "Invalid API key")
	c.Abort()
}
