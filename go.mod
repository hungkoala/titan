module gitlab.com/silenteer-oss/titan

go 1.15

require (
	cloud.google.com/go v0.54.0 // indirect
	emperror.dev/errors v0.7.0
	github.com/HdrHistogram/hdrhistogram-go v1.0.0 // indirect
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/go-playground/validator/v10 v10.2.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/uuid v1.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.5 // indirect
	github.com/hashicorp/consul/api v1.1.0
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/nats-io/nats.go v1.10.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/common v0.10.0
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	gitlab.com/silenteer-oss/hestia v1.0.30
	go.etcd.io/bbolt v1.3.5 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0 // indirect
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	golang.org/x/text v0.3.4 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	google.golang.org/grpc v1.33.2 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	logur.dev/adapter/logrus v0.4.1
	logur.dev/logur v0.16.2

)

replace google.golang.org/grpc => google.golang.org/grpc v1.29.0
