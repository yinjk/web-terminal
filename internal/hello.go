/**
 *
 * @author inori
 * @create 2019-05-30 20:19
 */
package internal

import "github.com/gin-gonic/gin"

func HelloWord(ctx *gin.Context) {
	SuccessWithData("Golang 大法好！", ctx)
}
