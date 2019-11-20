package main

import "gitlab.com/silenteer/go-nats/_example/companyservice/internal/app"

func main() {
	app.NewServerAndStart()
	select {}
}
