
## How to use GoLang with a private Gitlab repo? ##
Run command below
~~~~
git config --global --add url."git@gitlab.com:".insteadOf "https://gitlab.com/"
~~~~

 
## Some of the areas Consideration: ##
 
 - architecture
 - package structure
 - building the application
 - testing
 - configuration
 - running the application (eg. in Docker)
 - developer environment/experience
 
 ## Features
 
 - CLI (using [spf13/Cobra](https://github.com/spf13/cobra))
 - configuration (using [spf13/viper](https://github.com/spf13/viper))
 - logging (using [logur.dev/logur](https://logur.dev/logur) and [sirupsen/logrus](https://github.com/sirupsen/logrus))
 - error handling (using [emperror.dev/emperror](https://emperror.dev/emperror))
 - messaging (using [NATS](https://github.com/nats-io))
 - metrics and tracing using [Prometheus](https://prometheus.io/) and [Jaeger](https://www.jaegertracing.io/) (via [OpenCensus](https://opencensus.io/))
 - health checks (using [AppsFlyer/go-sundheit](https://github.com/AppsFlyer/go-sundheit))
 - graceful restart (using [cloudflare/tableflip](https://github.com/cloudflare/tableflip)) and shutdown
 - support for multiple server/daemon instances (using [oklog/run](https://github.com/oklog/run))


## Must read Articles ##
Golang module (https://github.com/golang/go/wiki/Modules)

Don’t just check errors, handle them gracefully 
 - https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully
 - https://banzaicloud.com/blog/error-handling-go/
 
 