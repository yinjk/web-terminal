/**
 * 
 * @author inori
 * @create 2019-04-12 15:03
 */
package main

import (
    "flag"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "time"
)

func main() {
    var port string
    flag.StringVar(&port, "port", "8080", "the server port!")
    fmt.Println(port)
    engine := gin.Default()
    //engine.StaticFS("/swagger", http.Dir("swagger"))
    engine.Static("/static", "./static")
    initRouter(engine)

    var readTimeOut time.Duration = 10000000000000
    var writeTimeOut time.Duration = 10000000000000
    s := &http.Server{
        Addr:           ":"+port,
        Handler:        engine,
        ReadTimeout:    readTimeOut * time.Second,
        WriteTimeout:   writeTimeOut * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    // service connections
    if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        panic(err)
    }
}

func initRouter(engine *gin.Engine)  {
    engine.POST("/v1/terminal", HandleExecNodeShell)
    engine.GET("/v1/sockjs/*any", func(context *gin.Context) {
        handler := CreateAttachHandler("/v1/sockjs")
        handler.ServeHTTP(context.Writer, context.Request)
    })
    engine.POST("/v1/sockjs/*any", func(context *gin.Context) {
        handler := CreateAttachHandler("/v1/sockjs")
        handler.ServeHTTP(context.Writer, context.Request)
    })
    engine.OPTIONS("/v1/sockjs/*any", func(context *gin.Context) {
        handler := CreateAttachHandler("/v1/sockjs")
        handler.ServeHTTP(context.Writer, context.Request)
    })
}