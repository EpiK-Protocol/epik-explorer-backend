package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/EpiK-Protocol/epik-explorer-backend/api"
	"github.com/EpiK-Protocol/epik-explorer-backend/epik"
	"github.com/EpiK-Protocol/epik-explorer-backend/etc"
	"github.com/EpiK-Protocol/epik-explorer-backend/router"
	"github.com/EpiK-Protocol/epik-explorer-backend/storage"

	"github.com/gin-gonic/gin"
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
	router.StartRouter(r, etc.Config.Server.HTTPPort)
}

//StartUp 启动
func StartUp() {
	//初始化系统配置
	err := etc.Load("./conf/config.yml")
	if err != nil {
		panic(err)
	}
	fmt.Println(etc.Config)
	fmt.Println("init config done!")
	fmt.Printf("set system mode to %s.\n", etc.Config.Server.Mode)
	fmt.Println("init log engine done!")
	gin.SetMode(etc.Config.Server.Mode)

	//初始化数据库
	storage.InitDatabase()
	fmt.Println("init database engine done!")
	//初始化业务模块
	epik.StartFetch()
	//订时任务
	api.Start()
	fmt.Println("init pay model done!")
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		select {
		case <-c:
			fmt.Println("quit system")
			storage.CloseDatabase()
			os.Exit(0)
		}
	}()
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
	var buf bytes.Buffer
	tee := io.TeeReader(c.Request.Body, &buf)
	body, _ := ioutil.ReadAll(tee)
	c.Request.Body = ioutil.NopCloser(&buf)
	fmt.Println(c.Request.Header)
	fmt.Println(string(body))
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()
	// fmt.Printf("token:%s\n", c.GetHeader("token"))
	// fmt.Printf("channel:%s\n", c.GetHeader("channel"))
	// fmt.Printf("Response body: \n%s\n", blw.body.String())

}
