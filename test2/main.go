package main

import (
	"fmt"
	"math/rand"
	"os"
)

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("./test room room_uid authToken")
		return
	}

	room := os.Args[1]
	room_uid := os.Args[2]
	authToken := os.Args[3]

	statRoom(room, room_uid, authToken)
}
