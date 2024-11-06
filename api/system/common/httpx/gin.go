package httpx

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/errors"
	"rxcsoft.cn/pit3/lib/msg"
	myLogger "rxcsoft.cn/utils/logger"
)

// log出力
const (
	MsgProcessError = "Process %s Error:%v"
)

var log = myLogger.New()

// GinHTTPError 返回错误
func GinHTTPError(c *gin.Context, action string, err error) {
	er := errors.Parse(err.Error())
	log.Error(c, action, fmt.Sprintf(MsgProcessError, action, er.GetDetail()))
	c.JSON(500, ErrorResponse{
		Message: er.GetDetail(),
	})
	c.Abort()
}

// GinTokenError 返回token刷新出错的error
func GinTokenError(c *gin.Context, action string, err error) {
	log.Error(c, action, fmt.Sprintf(MsgProcessError, action, err))
	c.JSON(401, ErrorResponse{
		Message: msg.GetMsg("ja-JP", msg.Error, msg.E005),
	})
	c.Abort()
}
