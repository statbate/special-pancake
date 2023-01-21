package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Input struct {
	Args   []string `json:"args"`
	Method string   `json:"method"`
}

type TipMsg struct {
	Channel  string `json:"channel"`
	Messages []struct {
		Data string `json:"data"`
	} `json:"messages"`
}

type Donate struct {
	Name   string `json:"name"`
	From   string `json:"from_username"`
	Amount int64  `json:"amount"`
}

func getAuth(room string) (string, string, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", "", err
	}

	client := &http.Client{
		Jar: jar,
	}

	var rsp *http.Response

	rsp, err = client.Get("https://chaturbate.com/" + room)
	if err != nil {
		return "", "", err
	}
	buf, err := io.ReadAll(rsp.Body)
	if err != nil {
		rsp.Body.Close()
		return "", "", err
	}
	if rsp.StatusCode != 200 {
		return "", "", fmt.Errorf("%s", buf)
	}

	var room_uid string
	// var wschat_host string
	idx := strings.Index(string(buf), `room_uid`)
	if idx < 0 {
		return "", "", fmt.Errorf("failed to find room_uid: %s", buf)
	}
	start := idx + len(`room_uid\u0022: \u0022`)
	for {
		idx++
		if string(buf)[idx] != ',' {
			continue
		} else {
			break
		}
	}
	room_uid = string(buf)[start : idx-len(`\u0022`)]
	fmt.Printf("room_uid: %s\n", room_uid)
	var csrftoken string
	var xcookie http.Cookie
	u, _ := url.Parse("https://chaturbate.com")
	cookies := jar.Cookies(u)
	for _, cookie := range cookies {
		if cookie.Name != "csrftoken" {
			continue
		}
		fmt.Printf("cookie with csrftoken %s\n", cookie.Value)
		csrftoken = cookie.Value
		xcookie = *cookie
		break
	}
	xcookie.Name = "agreeterms"
	xcookie.Value = "1"
	cookies = append(cookies, &xcookie)
	jar.SetCookies(u, cookies)

	body, writer := io.Pipe()

	req, err := http.NewRequest(http.MethodPost, "https://chaturbate.com/push_service/auth/", body)
	if err != nil {
		return "", "", err
	}
	mwriter := multipart.NewWriter(writer)
	req.Header.Add("Content-Type", mwriter.FormDataContentType())
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Accept", "*/*")
	//	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Referrer", "https://chaturbate.com/"+room+"/")
	req.Header.Add("Accept-Language", "ru")
	req.Header.Add("Origin", "https://chaturbate.com")
	req.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.2 Safari/605.1.15")
	req.Header.Add("X-NewRelic-ID", "VQIGWV9aDxACUFNVDgMEUw==")
	req.Header.Add("tracestate", "1418997@nr=0-1-1418997-24506750-2b71146eaa569098----1674161211828")
	req.Header.Add("traceparent", "00-01e7ed3b72a4b6787f0fadeb83557180-2b71146eaa569098-01")
	req.Header.Add("X-CSRFToken", csrftoken)
	req.Header.Add("newrelic", "eyJ2IjpbMCwxXSwiZCI6eyJ0eSI6IkJyb3dzZXIiLCJhYyI6IjE0MTg5OTciLCJhcCI6IjI0NTA2NzUwIiwiaWQiOiI4MDFiYTg1ZGI0NWQwMTUwIiwidHIiOiIzMWYxNmQ1ZDJlZGRlNWM3NzQyZTM0NWM5NmVkY2MxMCIsInRpIjoxNjc0MzAwNjA5NTY0fX0=")
	req.Header.Add("X-NewRelic-ID", "VQIGWV9aDxACUFNVDgMEUw==")

	errchan := make(chan error)

	go func() {
		defer close(errchan)
		defer writer.Close()
		defer mwriter.Close()

		wr, err := mwriter.CreateFormField("topics")
		if err != nil {
			errchan <- err
			return
		}
		if _, err = wr.Write([]byte(strings.ReplaceAll(`{"RoomTipAlertTopic#RoomTipAlertTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomPurchaseTopic#RoomPurchaseTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomFanClubJoinedTopic#RoomFanClubJoinedTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomMessageTopic#RoomMessageTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"GlobalPushServiceBackendChangeTopic#GlobalPushServiceBackendChangeTopic":{},"RoomAnonPresenceTopic#RoomAnonPresenceTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"QualityUpdateTopic#QualityUpdateTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomNoticeTopic#RoomNoticeTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomEnterLeaveTopic#RoomEnterLeaveTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomPasswordProtectedTopic#RoomPasswordProtectedTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomModeratorPromotedTopic#RoomModeratorPromotedTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomModeratorRevokedTopic#RoomModeratorRevokedTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomStatusTopic#RoomStatusTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomTitleChangeTopic#RoomTitleChangeTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomSilenceTopic#RoomSilenceTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomKickTopic#RoomKickTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomUpdateTopic#RoomUpdateTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomSettingsTopic#RoomSettingsTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"}}`, "ROOM_UID", room_uid))); err != nil {
			errchan <- err
			return
		}

		wr, err = mwriter.CreateFormField("csrfmiddlewaretoken")
		if err != nil {
			errchan <- err
			return
		}

		if _, err = wr.Write([]byte(csrftoken)); err != nil {
			errchan <- err
			return
		}
	}()

	rsp, err = client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer rsp.Body.Close()
	r := rsp.Body
	/*
		r, err := gzip.NewReader(rsp.Body)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		defer r.Close()
	*/

	buf, err = io.ReadAll(r)
	if err != nil {
		return "", "", err
	}
	if rsp.StatusCode != 200 {
		return "", "", fmt.Errorf("status != 200: %s", buf)
	}

	type keyNameRsp struct {
		TokenRequest struct {
			KeyName string `json:"keyName"`
		} `json:"token_request"`
	}
	keyName := &keyNameRsp{}
	type tokenRequestRsp struct {
		TokenRequest json.RawMessage `json:"token_request"`
	}
	tokenRequest := &tokenRequestRsp{}
	if err = json.Unmarshal(buf, keyName); err != nil {
		return "", "", err
	}
	if err = json.Unmarshal(buf, tokenRequest); err != nil {
		return "", "", err
	}

	rsp, err = client.Post("https://realtime.pa.highwebmedia.com/keys/"+keyName.TokenRequest.KeyName+"/requestToken?rnd=9705437583116864", "application/json", bytes.NewReader(tokenRequest.TokenRequest))
	if err != nil {
		return "", "", err
	}
	defer rsp.Body.Close()
	buf, err = io.ReadAll(rsp.Body)
	if err != nil {
		return "", "", err
	}
	if rsp.StatusCode != 201 {
		fmt.Printf("buf %s\n", tokenRequest.TokenRequest)
		return "", "", fmt.Errorf("status != 201: %s", buf)
	}

	type tokenRsp struct {
		Token string `json:"token"`
	}
	tokenResponse := &tokenRsp{}
	if err = json.Unmarshal(buf, tokenResponse); err != nil {
		return "", "", err
	}

	return tokenResponse.Token, room_uid, nil
}

func statRoom(room, server string, _ url.URL) {
	authToken, room_uid, err := getAuth(room)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//	fmt.Printf("auth status: %v rsp: %s room_uid: %s\n", err, authToken, room_uid)

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
		// fmt.Printf("msg :%s\n", m)

		var v map[string]interface{}
		if err := json.Unmarshal(message, &v); err != nil {
			fmt.Println(err.Error())
			break
		}

		if !init {
			c.WriteMessage(websocket.TextMessage, []byte(`{"action":17, "auth":{"accessToken":"`+authToken+`"}}`))
			c.WriteMessage(websocket.TextMessage, []byte(`{"action":16, "connectionKey":"`+v["connectionKey"].(string)+`","connectionSerial": -1}`))
			for _, im := range initMessages {
				c.WriteMessage(websocket.TextMessage, []byte(im))
			}
			init = true
		}
		/*
		   {"action":15,"id":"tCcHlgv0cq:0","connectionSerial":36,"channel":"room:tip_alert:EVX8CQC","channelSerial":"e91SLOyowBKTaO39138171:370","timestamp":1674333061062,"messages":[{"encoding":"json","data":"{\"tid\": \"16743330610:75553\", \"ts\": 1674333061.040021, \"amount\": 1, \"message\": \"\", \"history\": true, \"is_anonymous_tip\": false, \"to_username\": \"chloewildd\", \"from_username\": \"dm___of_cw\", \"gender\": \"m\", \"is_broadcaster\": false, \"in_fanclub\": false, \"is_following\": true, \"is_mod\": false, \"has_tokens\": true, \"tipped_recently\": true, \"tipped_alot_recently\": true, \"tipped_tons_recently\": true, \"method\": \"lazy\", \"pub_ts\": 1674333061.0486424}","name":"room:tip_alert:EVX8CQC"}]}
		*/

		switch v["action"].(float64) {
		case 0:
			continue
		case 15:
			/*
				{"action":15, "channel":"room:tip_alert:EVX8CQC", "channelSerial":"e91SLOyowBKTaO39138171:415", "connectionSerial":5, "id":"0-UOFp1i5K:0", "messages":[]interface {}{map[string]interface {}{"data":"{\"tid\": \"16743339917:85230\", \"ts\": 1674333991.730678, \"amount\": 10, \"message\": \"\", \"history\": true, \"is_anonymous_tip\": false, \"to_username\": \"chloewildd\", \"from_username\": \"majikal666\", \"gender\": \"m\", \"is_broadcaster\": false, \"in_fanclub\": false, \"is_following\": true, \"is_mod\": false, \"has_tokens\": true, \"tipped_recently\": true, \"tipped_alot_recently\": true, \"tipped_tons_recently\": true, \"method\": \"lazy\", \"pub_ts\": 1674333991.7439415}", "encoding":"json", "name":"room:tip_alert:EVX8CQC"}}, "timestamp":1.674333991766e+12}

			*/

			tipMsg := &TipMsg{}

			timeout = time.Now().Unix() + 60*60

			if err := json.Unmarshal([]byte(m), tipMsg); err != nil {
				fmt.Println(err.Error())
				continue
			}
			// fmt.Printf("msg %s\n", m)

			if tipMsg.Channel != "room:tip_alert:"+room_uid {
				continue
			}
			for _, msg := range tipMsg.Messages {
				donate := &Donate{}
				if err := json.Unmarshal([]byte(msg.Data), donate); err != nil {
					fmt.Println(err.Error())
					continue
				}
				if len(donate.From) > 3 {
					fmt.Println(donate.From, " send ", donate.Amount, "tokens")
				}
			}

		}
	}
	c.Close()
}
