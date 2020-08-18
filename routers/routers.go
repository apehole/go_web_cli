/**
 *@Author: IronHuang
 *@Date: 2020/8/18 9:58 下午
**/

package routers

import (
	"github.com/gin-gonic/gin"
	"go_web_cli/logger"
	"net/http"
)

func SetUp() *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecover(true))
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}
