package ipx

import (
	"bytes"
	"net"

	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
)

func CheckIP(clientIP string, segments []*role.IPSegment) bool {
	for _, segment := range segments {
		cip := net.ParseIP(clientIP)
		// 判断IP地址是否合法
		if cip.To4() == nil {
			loggerx.ErrorLog("CheckIP", "client IP address is invalid")
			return false
		}

		sip := net.ParseIP(segment.Start)
		eip := net.ParseIP(segment.End)

		if bytes.Compare(cip, sip) >= 0 && bytes.Compare(cip, eip) <= 0 {
			return true
		}
	}
	return false
}
