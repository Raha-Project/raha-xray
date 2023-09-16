package handlers

import (
	"net/http"
	"raha-xray/api/entity"
	"raha-xray/logger"

	"github.com/gin-gonic/gin"
)

// func getRemoteIp(c *gin.Context) string {
// 	value := c.GetHeader("X-Forwarded-For")
// 	if value != "" {
// 		ips := strings.Split(value, ",")
// 		return ips[0]
// 	} else {
// 		addr := c.Request.RemoteAddr
// 		ip, _, _ := net.SplitHostPort(addr)
// 		return ip
// 	}
// }

func jsonMsg(c *gin.Context, msg string, err error) {
	jsonMsgObj(c, msg, nil, err)
}

func jsonObj(c *gin.Context, obj interface{}, err error) {
	jsonMsgObj(c, "", obj, err)
}

func jsonMsgObj(c *gin.Context, msg string, obj interface{}, err error) {
	m := entity.Msg{
		Obj: obj,
	}
	if err == nil {
		m.Success = true
		if msg != "" {
			m.Msg = msg + ": success"
		}
	} else {
		m.Success = false
		m.Msg = msg + " failed: " + err.Error()
		logger.Warning("failure: ", err)
	}
	c.JSON(http.StatusOK, m)
}

func pureJsonMsg(c *gin.Context, success bool, msg string) {
	if success {
		c.JSON(http.StatusOK, entity.Msg{
			Success: true,
			Msg:     msg,
		})
	} else {
		c.JSON(http.StatusOK, entity.Msg{
			Success: false,
			Msg:     msg,
		})
	}
}
