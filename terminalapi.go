/**
 * 
 * @author inori
 * @create 2019-03-18 15:42
 */
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

type TerminalResponse struct {
    Id string `json:"id"`
}
type Host struct {
    Ip       string `json:"ip"`
    Username string `json:"username"`
    Password string `json:"password"`
    Port     int    `json:"port"`
}

/**
 * 获取服务器 shell的sessionId
 * @param :
 * @return:
 * @author: inori
 * @time  : 2019/3/21 14:53
 */
func HandleExecNodeShell(context *gin.Context) {
    var host Host
    context.BindJSON(&host)
    if host.Ip == "" || host.Username == "" || host.Password == "" {
        context.JSON(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
        return
    }
    if host.Port == 0 {
        host.Port = 22
    }
    sessionId, err := genTerminalSessionId()
    if err != nil {
        Fail(err.Error(), context)
        return
    }
    terminalSessions.Set(sessionId, TerminalSession{
        id:       sessionId,
        bound:    make(chan error),
        sizeChan: make(chan TerminalSize),
    })

    go WaitForNodeTerminal(host.Ip, host.Username, host.Port, host.Password, sessionId)
    SuccessWithData(TerminalResponse{Id: sessionId}, context)
}
