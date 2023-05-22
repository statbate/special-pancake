package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

var uptime = time.Now().Unix()

func mapRooms() {

	data := make(map[string]*Info)

	for {
		select {
		case m := <-rooms.Add:
			data[m.room] = &Info{Id: m.Id, Auth: m.Auth, Proxy: m.Proxy, Rid: m.Rid, Start: m.Start, Last: m.Last, Online: m.Online, Income: m.Income, Dons: m.Dons, Tips: m.Tips, ch: m.ch}

		case s := <-rooms.Json:
			j, err := json.Marshal(data)
			if err == nil {
				s = string(j)
			}
			rooms.Json <- s

		case <-rooms.Count:
			rooms.Count <- len(data)

		case key := <-rooms.Del:
			delete(data, key)

		case room := <-rooms.Check:
			if _, ok := data[room]; !ok {
				room = ""
			}
			rooms.Check <- room

		case room := <-rooms.Stop:
			if _, ok := data[room]; ok {
				close(data[room].ch)
			}
		}
	}
}

func announceCount() {
	for {
		time.Sleep(30 * time.Second)
		rooms.Count <- 0
		l := <-rooms.Count
		msg, err := json.Marshal(struct {
			Chanel string `json:"chanel"`
			Count  int    `json:"count"`
		}{
			Chanel: "chaturbate",
			Count:  l,
		})
		if err == nil {
			socketServer <- msg
		}
	}
}

func reconnectRoom(workerData Info) {
	time.Sleep(5 * time.Second)
	fmt.Println("reconnect:", workerData.room, workerData.Id, workerData.Auth, workerData.Proxy)
	workerData.Last = time.Now().Unix()
	startRoom(workerData)
}

func xWorker(workerData Info, u url.URL) {

	fmt.Println("Start", workerData.room, "room_id", workerData.Id, "auth", workerData.Auth, "proxy", workerData.Proxy)

	rooms.Add <- workerData

	defer func() {
		rooms.Del <- workerData.room
	}()

	Dialer := *websocket.DefaultDialer

	if _, ok := conf.Proxy[workerData.Proxy]; ok {
		Dialer = websocket.Dialer{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http", // or "https" depending on your proxy
				Host:   conf.Proxy[workerData.Proxy],
				Path:   "/",
			}),
			HandshakeTimeout: 45 * time.Second, // https://pkg.go.dev/github.com/gorilla/websocket
		}
	}

	c, _, err := Dialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err.Error(), workerData.room)
		return
	}
	defer c.Close()

	dons := make(map[string]struct{})

	initMessages := []string{
		`{"action":10,"flags":327680,"channel":"room:tip_alert:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:purchase:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:fanclub:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:message:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"global:push_service","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room_anon:presence:` + workerData.Id + `:0","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:quality_update:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:notice:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:enter_leave:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:password_protected:` + workerData.Id + `:13","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:mod_promoted:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:mod_revoked:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:status:` + workerData.Id + `:13","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:title_change:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:silence:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:kick:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:update:` + workerData.Id + `","params":{}}`,
		`{"action":10,"flags":327680,"channel":"room:settings:` + workerData.Id + `","params":{}}`,
	}

	c.SetReadDeadline(time.Now().Add(1 * time.Minute))
	_, message, err := c.ReadMessage()
	if err != nil {
		fmt.Println(err.Error(), workerData.room)
		return
	}

	slog <- saveLog{Rid: workerData.Rid, Now: time.Now().Unix(), Mes: string(message)}

	input := struct {
		Action   int                 `json:"action"`
		Key      string              `json:"connectionkey"`
		Error    jsoniter.RawMessage `json:"error"`
		Channel  string              `json:"channel"`
		Messages jsoniter.RawMessage `json:"messages"`
	}{}

	if err := json.Unmarshal(message, &input); err != nil {
		fmt.Println(err.Error(), workerData.room)
		return
	}

	if err = c.WriteMessage(websocket.TextMessage, []byte(`{"action":17, "auth":{"accessToken":"`+workerData.Auth+`"}}`)); err != nil {
		fmt.Println(err.Error(), workerData.room)
		return
	}

	_, message, err = c.ReadMessage()
	if err != nil {
		fmt.Println(err.Error(), workerData.room)
		return
	}

	for _, im := range initMessages {
		if err = c.WriteMessage(websocket.TextMessage, []byte(im)); err != nil {
			fmt.Println(err.Error(), workerData.room)
			return
		}
	}

	ticker := time.NewTicker(60 * 60 * 8 * time.Second)
	defer ticker.Stop()

	leave := time.NewTicker(60 * 60 * 8 * time.Second)
	defer leave.Stop()

	for {

		select {
		case <-ticker.C:
			fmt.Println("too_long exit:", workerData.room)
			return
		case <-leave.C:
			fmt.Println("leave exit:", workerData.room)
			return
		case <-workerData.ch:
			fmt.Println("Exit room:", workerData.room)
			return
		default:
		}

		c.SetReadDeadline(time.Now().Add(30 * time.Minute))
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err.Error(), workerData.room)
			if workerData.Income > 1 && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				go reconnectRoom(workerData)
			}
			return
		}

		now := time.Now().Unix()

		m := string(message)
		slog <- saveLog{Rid: workerData.Rid, Now: now, Mes: m}

		//fmt.Println(m)

		if now > workerData.Last+60*20 {
			fmt.Println("no_mes exit:", workerData.room)
			return
		}

		if err := json.Unmarshal(message, &input); err != nil {
			fmt.Println(err.Error(), workerData.room)
			break
		}

		if input.Action == 15 {

			workerData.Last = now
			rooms.Add <- workerData

			if input.Channel == "room:enter_leave:"+workerData.Id {
				type Message struct {
					Data string `json:"data"`
				}
				type Data struct {
					User struct {
						Username    string `json:"username"`
						Gender      string `json:"gender"`
						IsBroadcast bool   `json:"is_broadcast"`
					} `json:"user"`
					Action string `json:"action"`
				}
				var topmsg []*Message
				if err := json.Unmarshal(input.Messages, &topmsg); err != nil {
					fmt.Println(err.Error(), workerData.room)
					continue
				}
				for _, msg := range topmsg.Messages {
					inmsg := &Data{}

					if err := json.Unmarshal([]byte(msg.Data), inmsg); err != nil {
						fmt.Println(err.Error(), workerData.room)
						continue
					}

					// inmsg.User.Username, inmsg.User.Gender, inmsg.User.IsBroadcast, inmsg.Action

					if inmsg.User.Username == workerData.room && inmsg.User.IsBroadcast && inmsg.Action == "leave" {
						fmt.Println("leave, start ticker", workerData.room)
						leave.Reset(60 * 5 * time.Second)
					}

					if inmsg.User.Username == workerData.room && inmsg.User.IsBroadcast && inmsg.Action == "enter" {
						fmt.Println("enter, stop ticker", workerData.room)
						leave.Reset(60 * 60 * 8 * time.Second)
					}
				}
			}

			if input.Channel == "room:tip_alert:"+workerData.Id {
				tips := []struct {
					Data string `json:"data"`
				}{}

				if err := json.Unmarshal(input.Messages, &tips); err != nil {
					fmt.Println(err.Error(), workerData.room)
					continue
				}

				donate := struct {
					Name   string `json:"to_username"`
					From   string `json:"from_username"`
					Amount int64  `json:"amount"`
				}{}

				for _, tip := range tips {
					if err := json.Unmarshal([]byte(tip.Data), &donate); err != nil {
						fmt.Println(err.Error(), workerData.room)
						continue
					}

					if donate.Amount < 1 {
						fmt.Println("empty amount", workerData.room)
						continue
					}

					if len(donate.From) < 4 {
						donate.From = "anon_tips"
					}

					workerData.Tips++
					if _, ok := dons[donate.From]; !ok {
						dons[donate.From] = struct{}{}
						workerData.Dons++
					}

					save <- saveData{Room: workerData.room, From: donate.From, Rid: workerData.Rid, Amount: donate.Amount, Now: now}
					workerData.Income += donate.Amount
					rooms.Add <- workerData

					fmt.Println(donate.From, "send", donate.Amount, "tokens to", workerData.room)
				}
			}
		}
	}
}
