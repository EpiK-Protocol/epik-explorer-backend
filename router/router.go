package router

import (
	"epik-explorer-backend/api"
	"epik-explorer-backend/utils"

	"github.com/gin-gonic/gin"
)

//StartRouter 配置路由
func StartRouter(e *gin.Engine, port int) {
	api.SetEpikExplorerAPI(e)
	e.Run(":" + utils.ParseString(port))
}
