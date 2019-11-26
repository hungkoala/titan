package main

import (
	"gitlab.com/silenteer/titan/examples/companyservice/internal/app"
)

func main() {
	app.NewServer().Start()
}
