module rxcsoft.cn/pit3/lib/logger

go 1.12

require (
	github.com/antonfisher/nested-logrus-formatter v1.3.0
	github.com/keepeye/logrus-filename v0.0.0-20190711075016-ce01a4391dd1
	github.com/micro/go-micro/v2 v2.9.1
	github.com/sirupsen/logrus v1.7.0
	go.mongodb.org/mongo-driver v1.4.4
	google.golang.org/appengine v1.6.6
	rxcsoft.cn/utils v0.0.0-00010101000000-000000000000
)

replace (
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	rxcsoft.cn/pit3/lib/msg => ../msg
	rxcsoft.cn/utils => ../../../utils
)
