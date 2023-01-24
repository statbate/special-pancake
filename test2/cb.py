#!/usr/bin/env python3

import cloudscraper, json, re, codecs, sys

def lowercase_escape(s):
    unicode_escape = codecs.getdecoder('unicode_escape')
    return re.sub(
        r'\\u[0-9a-fA-F]{4}',
        lambda m: unicode_escape(m.group(0))[0],
        s)

if len(sys.argv) < 2:
    print('./cb.py name'); 
    exit(0)


room = sys.argv[1]

scraper = cloudscraper.create_scraper(
  interpreter='nodejs',
  captcha={
    'provider': '2captcha',
    'api_key': ''
  }
)

# Get cookies + set agreeterms 1
html = scraper.get("https://chaturbate.com/"+room).text
cookies = scraper.cookies.get_dict()
cookies["agreeterms"] = "1"
#print(cookies)

# Get room_uid
result = re.search('window.initialRoomDossier = "(.*)";', html)
arr = json.loads(lowercase_escape(result.group(1)))
#print(arr["room_uid"])

# Send post to auth
topics = '{"RoomTipAlertTopic#RoomTipAlertTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomPurchaseTopic#RoomPurchaseTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomFanClubJoinedTopic#RoomFanClubJoinedTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomMessageTopic#RoomMessageTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"GlobalPushServiceBackendChangeTopic#GlobalPushServiceBackendChangeTopic":{},"RoomAnonPresenceTopic#RoomAnonPresenceTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"QualityUpdateTopic#QualityUpdateTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomNoticeTopic#RoomNoticeTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomEnterLeaveTopic#RoomEnterLeaveTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomPasswordProtectedTopic#RoomPasswordProtectedTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomModeratorPromotedTopic#RoomModeratorPromotedTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomModeratorRevokedTopic#RoomModeratorRevokedTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomStatusTopic#RoomStatusTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomTitleChangeTopic#RoomTitleChangeTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomSilenceTopic#RoomSilenceTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomKickTopic#RoomKickTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomUpdateTopic#RoomUpdateTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"},"RoomSettingsTopic#RoomSettingsTopic:ROOM_UID":{"broadcaster_uid":"ROOM_UID"}}'.replace("ROOM_UID", arr["room_uid"])
data = [('topics', topics), ('csrfmiddlewaretoken', cookies["csrftoken"])]
headers = {
    'Origin': 'https://chaturbate.com',
	'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.2 Safari/605.1.15',
	'X-Requested-With': 'XMLHttpRequest',
	'Referrer': 'https://chaturbate.com/'+room+'/',
	'X-CSRFToken': cookies["csrftoken"],
}
auth = json.loads(scraper.post("https://chaturbate.com/push_service/auth/", headers=headers, cookies=cookies, data=data).text)
#print(auth["token_request"])

headers = {
	'accept': 'application/json',
	'content-type': 'application/json',
	'origin': 'https://chaturbate.com',
	'referer': 'https://chaturbate.com/',
}
keys = json.loads(scraper.post('https://realtime.pa.highwebmedia.com/keys/'+auth["token_request"]["keyName"]+'/requestToken?rnd=9705437583116864', headers=headers, cookies=cookies, json=auth["token_request"]).text)
print(json.dumps({'id': arr["room_uid"], 'auth': keys["token"]}))
