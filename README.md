
## How to use GoLang with a private Gitlab repo? ##
Run command below
~~~~
git config --global --add url."git@gitlab.com:".insteadOf "https://gitlab.com/"

export GOPRIVATE="gitlab.com/silenteer,git.tutum.dev/medi/tutum"
~~~~

 
## Consideration: ##
 
 - architecture
 - package structure
 - building the application
 - testing
 - configuration
 - running the application (eg. in Docker)
 - developer environment/experience
 
 ## Go libraries
 - project layout (follow [Standard Go Project Layout](https://github.com/golang-standards/project-layout))
 - cli (using [spf13/Cobra](https://github.com/spf13/cobra))
 - configuration (using [spf13/viper](https://github.com/spf13/viper))
 - logging (using [logur.dev/logur](https://logur.dev/logur) and [sirupsen/logrus](https://github.com/sirupsen/logrus))
 - error handling (using [github.com/pkg/errors](github.com/pkg/errors))
 - messaging (using [NATS](https://github.com/nats-io))
 - router (using [chi](https://github.com/go-chi/chi))
 - metrics and tracing using [Prometheus](https://prometheus.io/) and [Jaeger](https://www.jaegertracing.io/) (via [OpenCensus](https://opencensus.io/))
 - health checks (using [AppsFlyer/go-sundheit](https://github.com/AppsFlyer/go-sundheit))
 - test (using [testify](https://github.com/stretchr/testify))
 - load test (using [vegeta](https://github.com/tsenart/vegeta)
 - file system ([spf13/afero](https://github.com/spf13/afero)) 
 - code gen  https://clipperhouse.com/gen/overview/, https://github.com/awalterschulze/goderive, https://github.com/hexdigest/gowrap
 - golang jq like https://github.com/tidwall/gjson

## Dev tools ##
- go format  https://golang.org/cmd/gofmt/
- go imports https://godoc.org/golang.org/x/tools/cmd/goimports
- go lint https://github.com/golangci/golangci-lint

## Must read Articles ##
Golang module (https://github.com/golang/go/wiki/Modules)

Don’t just check errors, handle them gracefully 
 - https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully
 - https://banzaicloud.com/blog/error-handling-go/
 
Golang common mistakes
  - http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/
  
Twelve Go Best Practices
  - https://talks.golang.org/2013/bestpractices.slide#1

Go awesome
 - https://oxozle.com/awetop/avelino-awesome-go/
 
 Table Driven Tests
- https://github.com/golang/go/wiki/TableDrivenTests