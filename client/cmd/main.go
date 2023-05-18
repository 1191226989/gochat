package main

import (
	"flag"
	"fmt"
	"gochat/client/internal/service"
)

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "服务器IP地址(默认127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "服务器端口(默认8888)")

	flag.Parse()
}

func main() {
	client, err := service.NewClient(serverIp, serverPort)
	if err != nil {
		fmt.Println(">>>>>> 服务器连接失败")
	}
	fmt.Println(">>>>>> 服务器连接成功")

	go client.DealResponse()

	client.Run()
}
