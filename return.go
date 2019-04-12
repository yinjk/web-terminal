/**
 * 
 * @author inori
 * @create 2019-04-12 15:47
 */
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

const (
    SUCCESS                    = 0
    FAIL                       = 500
)

/**
**快捷成功输出  {"state":200,"message":"执行成功","data":null}
 */
func Success(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "code":    SUCCESS,
        "message": "执行成功",
        "data":    nil,
    })
}

/**
*带内容成功输出比如显示的数据
 */
func SuccessWithData(data interface{}, c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "code":    SUCCESS,
        "message": "执行成功",
        "data":    data,
    })
}

/**
**失败输出 err为输出前端的提升错误信息      {"state":500,"message":"执行失败","data":nil}
 */
func Fail(err string, c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "code":    FAIL,
        "message": err,
        "data":    nil,
    })
}
