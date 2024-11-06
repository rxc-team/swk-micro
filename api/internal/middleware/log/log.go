package log

import (
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger gin日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logrus.New()

		formatter := &nested.Formatter{
			HideKeys:        true,
			NoFieldsColors:  false,
			NoColors:        false,
			FieldsOrder:     []string{"log_type", "client_ip", "app_name", "req_uri", "req_method", "status_code", "latency_time"},
			TimestampFormat: "2006-01-02 15:04:05",
		}

		log.SetFormatter(formatter)

		// 开始时间
		startTime := time.Now()
		// 处理请求
		c.Next()
		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqURI := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()
		log.WithFields(logrus.Fields{
			"log_type":     "GIN",
			"client_ip":    clientIP,
			"app_name":     "internal",
			"req_uri":      reqURI,
			"req_method":   reqMethod,
			"status_code":  statusCode,
			"latency_time": latencyTime,
		}).Info()
	}
}
