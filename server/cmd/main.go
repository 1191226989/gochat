package main

import "gochat/server/internal/service"

func main() {
	s := service.NewServer("127.0.0.1", 8888)
	s.Start()
}
