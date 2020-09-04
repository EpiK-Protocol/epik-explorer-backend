package api

import (
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/epik"
	"github.com/gin-gonic/gin"
)

func tokenConfirm(c *gin.Context) {
	parseToken(c)
	if !c.GetBool("auth") {
		// c.AbortWithStatus(http.StatusUnauthorized)
		c.AbortWithStatusJSON(200, map[string]interface{}{"code": map[string]interface{}{"code": 401, "message": "unauthorized"}})
		return
	}
	c.Next()
}

func tokenParse(c *gin.Context) {
	parseToken(c)
	c.Next()
}

//ParseToken 解析用户信息
func parseToken(c *gin.Context) {
	tokenStr := c.GetHeader("token")
	valid := false
	var userID int64
	var expired int64
	var platform string
	if tokenStr != "" {
		userID, expired, platform, valid = epik.ParseToken(tokenStr)
		if expired < time.Now().Unix() {
			valid = false
		}
	}
	if valid {
		c.Set("auth", true)
		c.Set("uid", userID)
		if expired < time.Now().Add(time.Hour*(-24)).Unix() { //token续期
			tokenStr = epik.CreateToken(userID, platform, time.Hour*24*30)
			c.Header("token", tokenStr)
		}
	}
}
