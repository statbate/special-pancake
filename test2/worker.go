package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
	"github.com/gorilla/websocket"
)

func statRoom(room, room_uid, authToken string) {
	u, err := url.Parse("wss://realtime.pa.highwebmedia.com/?access_token=" + authToken + "&format=json&heartbeats=true&v=1.2&agent=ably-js%2F1.2.13%20browser&remainPresentFor=0")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	timeout := time.Now().Unix() + 60*60
	init := false

	initMessages := []string{
		`{"action":16, "connectionKey":{"accessToken":"` + authToken + `"}}`,
		`{"action":10,"flags":327680,"channel":"room:tip_alert:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:purchase:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:fanclub:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:message:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"global:push_service","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room_anon:presence:` + room_uid + `:0","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:quality_update:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:notice:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:enter_leave:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:password_protected:` + room_uid + `:13","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:mod_promoted:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:mod_revoked:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:status:` + room_uid + `:13","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:title_change:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:silence:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:kick:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:update:` + room_uid + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:settings:` + room_uid + `","params":{}}`,
	}

	for {

		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		if time.Now().Unix() > timeout {
			fmt.Println("Timeout room:", room)
			break
		}

		m := string(message)

		input := struct {
			Action   int    `json:"action"`
			Key      string `json:"connectionkey"`
			Error    json.RawMessage `json:"error"`
			Channel  string `json:"channel"`
			Messages json.RawMessage `json:"messages"`
		}{}
		
		if err := json.Unmarshal(message, &input); err != nil {
			fmt.Println(err.Error())
			break
		}

		if !init {
			c.WriteMessage(websocket.TextMessage, []byte(`{"action":17, "auth":{"accessToken":"`+authToken+`"}}`))
			c.WriteMessage(websocket.TextMessage, []byte(`{"action":16, "connectionKey":"`+input.Key+`","connectionSerial": -1}`))
			for _, im := range initMessages {
				c.WriteMessage(websocket.TextMessage, []byte(im))
			}
			init = true
		}
		
		

		if input.Action == 15 {
			
			timeout = time.Now().Unix() + 60*60
			
			tips := []struct {
					Data string `json:"data"`
			}{}

			if err := json.Unmarshal([]byte(m), &tips); err != nil {
				fmt.Println(err.Error())
				continue
			}
			
			if input.Channel == "room:tip_alert:"+room_uid {
				
				donate := struct {
					Name   string `json:"to_username"`
					From   string `json:"from_username"`
					Amount int64  `json:"amount"`
				}{}
					
				for _, tip := range tips {
					if err := json.Unmarshal([]byte(tip.Data), donate); err != nil {
						fmt.Println(err.Error())
						continue
					}
					if len(donate.From) > 3 {
						fmt.Println(donate.From, "send ", donate.Amount, "tokens")
					}
				}
			}
		}
	}
	c.Close()
}
