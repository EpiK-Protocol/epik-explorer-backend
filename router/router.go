package router

import (
	"github.com/EpiK-Protocol/epik-explorer-backend/api"
	"github.com/EpiK-Protocol/epik-explorer-backend/utils"

	"github.com/gin-gonic/gin"
)

//StartRouter 配置路由
func StartRouter(e *gin.Engine, port int) {
	api.SetEpikExplorerAPI(e)
	api.SetTestNetAPI(e)
	api.SetWalletAPI(e)
	api.SetAdminAPI(e)
	e.Run(":" + utils.ParseString(port))
}
