module rxcsoft.cn/pit3/srv/import

go 1.13

replace (
	cloud.google.com/go => cloud.google.com/go v0.53.0
	cloud.google.com/go/storage => cloud.google.com/go/storage v1.9.0
	google.golang.org/api => google.golang.org/api v0.14.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20191216164720-4f79533eabd1
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	rxcsoft.cn/k8s/go/web => ../../../k8s/go/web
	rxcsoft.cn/pit3/lib/logger => ../../lib/logger
	rxcsoft.cn/pit3/lib/msg => ../../lib/msg
	rxcsoft.cn/pit3/srv/database => ../database
	rxcsoft.cn/pit3/srv/global => ../global
	rxcsoft.cn/pit3/srv/manage => ../manage
	rxcsoft.cn/pit3/srv/task => ../task
	rxcsoft.cn/pit3/srv/workflow => ../workflow
	rxcsoft.cn/utils => ../../../utils
)

require (
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.3.2
	github.com/Andrew-M-C/go.timeconv v0.3.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.1.2
	github.com/kataras/i18n v0.0.6
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins/broker/rabbitmq/v2 v2.9.1
	github.com/micro/go-plugins/logger/logrus/v2 v2.9.1
	github.com/micro/go-plugins/registry/consul/v2 v2.9.1
	github.com/micro/go-plugins/registry/kubernetes/v2 v2.9.1
	github.com/micro/go-plugins/transport/tcp/v2 v2.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1
	go.mongodb.org/mongo-driver v1.5.2
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781
	golang.org/x/text v0.3.6
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	rxcsoft.cn/pit3/lib/logger v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/lib/msg v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/database v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/global v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/manage v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/task v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/workflow v0.0.0-00010101000000-000000000000
	rxcsoft.cn/utils v0.0.0-00010101000000-000000000000
)
