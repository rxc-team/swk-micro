module rxcsoft.cn/pit3/api/outer

go 1.13

require (
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.3.2
	github.com/antonfisher/nested-logrus-formatter v1.3.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.7.7
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins/logger/logrus/v2 v2.9.1
	github.com/micro/go-plugins/registry/consul/v2 v2.9.1
	github.com/micro/go-plugins/registry/kubernetes/v2 v2.9.1
	github.com/micro/go-plugins/transport/tcp/v2 v2.9.1
	github.com/panjf2000/ants/v2 v2.4.7
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.5.0
	github.com/ugorji/go v1.1.13 // indirect
	go.mongodb.org/mongo-driver v1.5.2
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781
	golang.org/x/text v0.3.6
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	rxcsoft.cn/pit3/lib/logger v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/lib/msg v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/database v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/global v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/import v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/manage v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/task v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/workflow v0.0.0-00010101000000-000000000000
	rxcsoft.cn/utils v0.0.0-00010101000000-000000000000
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.53.0
	cloud.google.com/go/storage => cloud.google.com/go/storage v1.9.0
	google.golang.org/api => google.golang.org/api v0.14.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20191216164720-4f79533eabd1
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	rxcsoft.cn/k8s/go/web => ../../k8s/go/web
	rxcsoft.cn/pit3/lib/logger => ../../lib/logger
	rxcsoft.cn/pit3/lib/msg => ../../lib/msg
	rxcsoft.cn/pit3/srv/database => ../../srv/database
	rxcsoft.cn/pit3/srv/global => ../../srv/global
	rxcsoft.cn/pit3/srv/import => ../../srv/import
	rxcsoft.cn/pit3/srv/manage => ../../srv/manage
	rxcsoft.cn/pit3/srv/report => ../../srv/report
	rxcsoft.cn/pit3/srv/storage => ../../srv/storage
	rxcsoft.cn/pit3/srv/task => ../../srv/task
	rxcsoft.cn/pit3/srv/workflow => ../../srv/workflow
	rxcsoft.cn/utils => ../../../utils
)
