module rxcsoft.cn/pit3/api/internal

go 1.13

require (
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.3.2
	github.com/alexellis/go-execute v0.0.0-20210616110041-528c0bba4494
	github.com/antonfisher/nested-logrus-formatter v1.3.0
	github.com/casbin/casbin/v2 v2.37.0
	github.com/casbin/mongodb-adapter/v3 v3.2.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/sessions v0.0.4
	github.com/gin-gonic/gin v1.7.7
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-resty/resty/v2 v2.7.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/websocket v1.4.2
	github.com/jung-kurt/gofpdf v1.16.2
	github.com/kataras/i18n v0.0.6
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins/broker/rabbitmq/v2 v2.9.1
	github.com/micro/go-plugins/logger/logrus/v2 v2.9.1
	github.com/micro/go-plugins/registry/consul/v2 v2.9.1
	github.com/micro/go-plugins/registry/kubernetes/v2 v2.9.1
	github.com/micro/go-plugins/transport/tcp/v2 v2.9.1
	github.com/mojocn/base64Captcha v1.2.2
	github.com/panjf2000/ants/v2 v2.4.7
	github.com/robfig/cron/v3 v3.0.1
	github.com/rxc-team/dcron v0.2.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1
	github.com/ugorji/go v1.1.13 // indirect
	github.com/utrack/gin-csrf v0.0.0-20190424104817-40fb8d2c8fca
	github.com/yidane/formula v0.0.0-20200220154705-ec0e6bc4831b
	go.mongodb.org/mongo-driver v1.5.3
	golang.org/x/crypto v0.0.0-20201216223049-8b5274cf687f
	golang.org/x/net v0.0.0-20211029224645-99673261e6eb
	golang.org/x/text v0.3.6
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	rxcsoft.cn/pit3/lib/logger v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/lib/msg v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/database v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/global v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/import v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/journal v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/manage v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/report v0.0.0-00010101000000-000000000000
	rxcsoft.cn/pit3/srv/storage v0.0.0-00010101000000-000000000000
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
	rxcsoft.cn/pit3/srv/journal => ../../srv/journal
	rxcsoft.cn/pit3/srv/manage => ../../srv/manage
	rxcsoft.cn/pit3/srv/report => ../../srv/report
	rxcsoft.cn/pit3/srv/storage => ../../srv/storage
	rxcsoft.cn/pit3/srv/task => ../../srv/task
	rxcsoft.cn/pit3/srv/workflow => ../../srv/workflow
	rxcsoft.cn/utils => ../../../utils
)
