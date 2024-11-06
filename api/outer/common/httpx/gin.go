package httpx

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/api/outer/common/loggerx"
	"rxcsoft.cn/pit3/lib/msg"
)

// log出力
const (
	MsgProcessError = "Process %s Error:%v"
)

// GinHTTPError 返回错误
func GinHTTPError(c *gin.Context, action string, err error) {
	er := errors.Parse(err.Error())
	if er.GetDetail() == mongo.ErrNoDocuments.Error() {
		loggerx.InfoLog(c, action, fmt.Sprintf(MsgProcessError, action, err))
		c.JSON(200, Response{
			Status:  2,
			Message: msg.GetMsg("ja-JP", msg.Error, msg.E002, fmt.Sprintf(Temp, "GIN", action)),
			Data:    gin.H{},
		})
		return
	}

	if er.GetDetail() == "has no change" {
		loggerx.InfoLog(c, action, fmt.Sprintf(MsgProcessError, action, err))
		c.JSON(200, Response{
			Status:  2,
			Message: "has no change",
			Data:    gin.H{},
		})
		return
	}

	loggerx.FailureLog(c, action, fmt.Sprintf(MsgProcessError, action, er.GetDetail()))
	c.JSON(500, ErrorResponse{
		Message: er.GetDetail(),
	})
	c.Abort()
}

// GinTokenError 返回token刷新出错的error
func GinTokenError(c *gin.Context, action string, err error) {
	loggerx.FailureLog(c, action, fmt.Sprintf(MsgProcessError, action, err))
	c.JSON(600, ErrorResponse{
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I001),
	})
	c.Abort()
}
