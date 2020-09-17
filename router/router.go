package router

import (
	"fmt"

	"github.com/EpiK-Protocol/epik-explorer-backend/api"

	"github.com/gin-gonic/gin"
)

//StartRouter 配置路由
func StartRouter(e *gin.Engine, port int64) {
	api.SetEpikExplorerAPI(e)
	e.Run(fmt.Sprintf(":%d", port))
}
