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
	if len(os.Args) < 3 {
		fmt.Println("./test room_uid authToken")
		return
	}

	room_uid := os.Args[1]
	authToken := os.Args[2]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := ably.NewRealtime(
		ably.WithAutoConnect(true),
		ably.WithToken(authToken),
		// ably.WithLogLevel(ably.LogDebug),
		ably.WithAuthMethod(http.MethodGet),
		ably.WithUseTokenAuth(true),
		ably.WithRealtimeHost("realtime.pa.highwebmedia.com"),
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	type donateMsg struct {
		Name   string `json:"to_username"`
		From   string `json:"from_username"`
		Amount int64  `json:"amount"`
	}

	type leaveMsg struct {
		Action        string `json:"action"`
		IsBroadcaster bool   `json:"is_broadcaster"`
	}

	tipChannel := "room:tip_alert:" + room_uid
	leaveChannel := "room:enter_leave:" + room_uid

	handler := func(msg *ably.Message) {
		switch msg.Name {
		case leaveChannel:
			leave := &leaveMsg{}
			if err := json.Unmarshal([]byte(msg.Data.(string)), leave); err != nil {
				fmt.Println(err.Error())
				cancel()
				return
			}
			if leave.Action == "leave" && leave.IsBroadcaster {
				fmt.Println("user finished and exit")
				cancel()
				return
			}
		case tipChannel:
			donate := &donateMsg{}
			if err := json.Unmarshal([]byte(msg.Data.(string)), donate); err != nil {
				fmt.Println(err.Error())
				cancel()
				return
			}
			if donate.From == "" {
				donate.From = "anonymous"
			}
			fmt.Println(donate.From, "send ", donate.Amount, "tokens")
		}
	}

	for _, channel := range []string{tipChannel, leaveChannel} {
		rc := c.Channels.Get(channel)
		defer func() {
			if err := c.Channels.Release(ctx, channel); err != nil {
				fmt.Println(err.Error())
				cancel()
			}
		}()

		if err := rc.Attach(ctx); err != nil {
			fmt.Println(err.Error())
			return
		}
		/* not needed as used Release before
		defer func() {
			if err := rc.Detach(ctx); err != nil {
				fmt.Println(err.Error())
			}
		}()
		*/

		rccancel, err := rc.SubscribeAll(ctx, handler)
		if err != nil {
			fmt.Println(err.Error())
			cancel()
		}
		defer rccancel()
		ecancel := rc.On(ably.ChannelEventDetached, func(state ably.ChannelStateChange) {
			fmt.Printf("received detached")
			cancel()
		})
		defer ecancel()
	}

	defer c.Close()
	<-ctx.Done()
}
