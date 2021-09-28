package main

import (
	_ "chatroom/global"
	"chatroom/server"
	"fmt"
	"log"
	"net/http"
)

var (
	addr   = ":2022"
	banner = `
    ____                 _____
   |     |    |    /\      |
   |     |____|   /  \     |
   |     |    |  /----\    |
   |____ |    | /      \   |

Go 语言编程之旅 —— 一起用 Go 做项目：ChatRoom，start on：%s
`
)

func main() {
	fmt.Printf(banner, addr)

	server.RegisterHandle()

	log.Fatal(http.ListenAndServe(addr, nil))
}
