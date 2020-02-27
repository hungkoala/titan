package main

import (
	"gitlab.com/silenteer-oss/titan/examples/companyservice/internal/app"
)

func main() {
	app.NewServer().Start()
}
