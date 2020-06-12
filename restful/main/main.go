package main

import "gitlab.com/silenteer-oss/titan/restful"

func main() {
	port := "6968"

	server := restful.NewServer(port)

	server.Start()

}
