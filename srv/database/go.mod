module rxcsoft.cn/pit3/srv/database

go 1.13

replace (
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	rxcsoft.cn/k8s/go/web => ../../../k8s/go/web
	rxcsoft.cn/pit3/lib/logger => ../../lib/logger
	rxcsoft.cn/pit3/lib/msg => ../../lib/msg
	rxcsoft.cn/pit3/srv/global => ../global
	rxcsoft.cn/pit3/srv/journal => ../journal
	rxcsoft.cn/pit3/srv/manage => ../manage
	rxcsoft.cn/utils => ../../../utils
)

require (
	github.com/Andrew-M-C/go.timeconv v0.3.0
	github.com/golang/protobuf v1.5.2
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins/logger/logrus/v2 v2.9.1
	github.com/micro/go-plugins/registry/consul/v2 v2.9.1
	github.com/micro/go-plugins/registry/kubernetes/v2 v2.9.1
	github.com/micro/go-plugins/transport/tcp/v2 v2.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1
	go.mongodb.org/mongo-driver v1.5.2
	rxcsoft.cn/pit3/lib/logger v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/global v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/journal v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/manage v0.0.0-00010101000000-000000000000
	rxcsoft.cn/utils v0.0.0-00010101000000-000000000000
)
