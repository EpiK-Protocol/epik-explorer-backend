package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"epik-explorer-backend/epik"
	"epik-explorer-backend/router"
	"epik-explorer-backend/storage"
	"epik-explorer-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

func main() {
	StartUp()
	r := gin.New()
	if gin.Mode() == gin.DebugMode {
		r.Use(gin.Logger())
	}
	r.Use(gin.Recovery())
	if gin.Mode() == "debug" {
		r.Use(ginBodyLogMiddleware)
	}
	httpport := utils.ReadConfig("app.http.port")
	router.StartRouter(r, utils.ParseInt(httpport))
}

//StartUp 启动
func StartUp() {
	//初始化系统配置
	utils.LoadConfig()
	fmt.Println("init config done!")
	fmt.Printf("set system mode to %s.\n", utils.ReadConfig("app.mode"))
	utils.Log.AddHook(utils.NewLfsHook(utils.ReadConfig("log.dir")))
	fmt.Println("init log engine done!")
	gin.SetMode(utils.ReadConfig("app.mode"))
	//初始化数据库
	storage.InitDatabase()
	fmt.Println("init database engine done!")
	//初始化业务模块
	epik.LoadData()
	epik.StartFetch()
	//订时任务
	InitCron()
	fmt.Println("init pay model done!")
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		select {
		case <-c:
			fmt.Println("quit system")
			epik.SaveData()
			storage.CloseDatabase()
			os.Exit(0)
		}
	}()
}

//InitCron 订时任务
func InitCron() {
	crontab := cron.New()
	// crontab.AddFunc("0/1 * * * * *", func() { CallbackRetry(1) })
	// crontab.AddFunc("0/1 * * * * *", func() { wallet.RefreshICXOPool() })

	crontab.Start()
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func ginBodyLogMiddleware(c *gin.Context) {
	// if c.Request.Method == "GET" {
	// 	c.Next()
	// 	return
	// }
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()
	// fmt.Printf("token:%s\n", c.GetHeader("token"))
	// fmt.Printf("channel:%s\n", c.GetHeader("channel"))
	// fmt.Printf("Response body: \n%s\n", blw.body.String())

}
