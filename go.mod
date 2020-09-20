module github.com/miuer/miu-netdisc

go 1.14

require (
	github.com/aliyun/aliyun-oss-go-sdk v0.0.0-20190307165228-86c17b95fcd5
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.2
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-micro/v3 v3.0.0-beta.2
	github.com/micro/go-plugins/registry/consul v0.0.0-20200119172437-4fe21aa238fd
	github.com/micro/go-plugins/registry/consul/v2 v2.9.1
	github.com/streadway/amqp v1.0.0
	golang.org/x/net v0.0.0-20200904194848-62affa334b73
	google.golang.org/grpc v1.31.1
	gopkg.in/amz.v1 v1.0.0-20150111123259-ad23e96a31d2
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
