module rxcsoft.cn/pit3/api/system

go 1.13

replace (
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	rxcsoft.cn/pit3/lib/logger => ../../lib/logger
	rxcsoft.cn/pit3/lib/msg => ../../lib/msg
	rxcsoft.cn/pit3/srv/global => ../../srv/global
	rxcsoft.cn/pit3/srv/manage => ../../srv/manage
	rxcsoft.cn/utils => ../../../utils
)

require (
	github.com/antonfisher/nested-logrus-formatter v1.3.0
	github.com/gin-gonic/gin v1.7.7
	github.com/go-redis/redis/v8 v8.11.5
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins/broker/rabbitmq/v2 v2.9.1
	github.com/micro/go-plugins/logger/logrus/v2 v2.9.1
	github.com/micro/go-plugins/registry/consul/v2 v2.9.1
	github.com/micro/go-plugins/registry/kubernetes/v2 v2.9.1
	github.com/micro/go-plugins/transport/tcp/v2 v2.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0 // indirect
	go.mongodb.org/mongo-driver v1.9.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/sys v0.0.0-20220408201424-a24fb2fb8a0f // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	rxcsoft.cn/pit3/lib/msg v0.0.0-00010101000000-000000000000
	rxcsoft.cn/utils v0.0.0-00010101000000-000000000000
)
