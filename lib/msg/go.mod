module rxcsoft.cn/pit3/lib/msg

go 1.12

require (
	github.com/micro/go-micro/v2 v2.9.1
	github.com/sirupsen/logrus v1.7.0
	rxcsoft.cn/pit3/srv/global v0.0.0-00010101000000-000000000000
)

replace (
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	rxcsoft.cn/pit3/lib/logger => ../logger
	rxcsoft.cn/pit3/srv/global => ../../srv/global
	rxcsoft.cn/utils => ../../../utils
)
