package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	ably "github.com/ably/ably-go/ably"
)

func main() {
	ctx := context.Background()
	c, err := ably.NewRealtime(
		ably.WithAutoConnect(true),
		ably.WithToken(os.Args[1]),
		// ably.WithLogLevel(ably.LogDebug),
		ably.WithAuthMethod(http.MethodGet),
		ably.WithUseTokenAuth(true),
		ably.WithRealtimeHost("realtime.pa.highwebmedia.com"),
	)
	if err != nil {
		panic(err)
	}

	rc := c.Channels.Get("room:tip_alert:" + os.Args[2])
	if err := rc.Attach(ctx); err != nil {
		panic(err)
	}
	defer func() {
		if err := rc.Detach(ctx); err != nil {
			panic(err)
		}
	}()

	type donateMsg struct {
		Name   string `json:"to_username"`
		From   string `json:"from_username"`
		Amount int64  `json:"amount"`
	}

	cancel, err := rc.SubscribeAll(ctx, func(msg *ably.Message) {
		donate := &donateMsg{}
		if err := json.Unmarshal([]byte(msg.Data.(string)), donate); err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(donate.From) > 3 {
			fmt.Println(donate.From, "send ", donate.Amount, "tokens")
		}
	})
	if err != nil {
		panic(err)
	}
	defer cancel()
	defer c.Close()
	select {}
}
