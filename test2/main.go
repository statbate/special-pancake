package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
)

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("./test room server")
		return
	}

	room := os.Args[1]
	server := os.Args[2]

	u := url.URL{Scheme: "wss", Host: server + ".highwebmedia.com", Path: "/ws/555/kmdqiune/websocket"}

	statRoom(room, server, u)
}
