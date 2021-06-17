package model

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"unsafe"
)

type MusicData struct {
	Code      int    `json:"code"`
	Curtime   int64  `json:"curTime"`
	Data      Data   `json:"data"`
	Msg       string `json:"msg"`
	Profileid string `json:"profileId"`
	Reqid     string `json:"reqId"`
	Tid       string `json:"tId"`
}
type Mvpayinfo struct {
	Play int `json:"play"`
	Vid  int `json:"vid"`
	Down int `json:"down"`
}
type Feetype struct {
	Song string `json:"song"`
	Vip  string `json:"vip"`
}
type Payinfo struct {
	Play             string  `json:"play"`
	Download         string  `json:"download"`
	LocalEncrypt     string  `json:"local_encrypt"`
	Limitfree        int     `json:"limitfree"`
	Cannotdownload   int     `json:"cannotDownload"`
	ListenFragment   string  `json:"listen_fragment"`
	Cannotonlineplay int     `json:"cannotOnlinePlay"`
	Feetype          Feetype `json:"feeType"`
	Down             string  `json:"down"`
}
type List struct {
	Musicrid         string    `json:"musicrid"`
	Barrage          string    `json:"barrage"`
	Artist           string    `json:"artist"`
	Pic              string    `json:"pic"`
	Isstar           int       `json:"isstar"`
	Rid              int       `json:"rid"`
	Duration         int       `json:"duration"`
	Score100         string    `json:"score100"`
	ContentType      string    `json:"content_type"`
	Track            int       `json:"track"`
	Haslossless      bool      `json:"hasLossless"`
	Hasmv            int       `json:"hasmv"`
	Releasedate      string    `json:"releaseDate"`
	Album            string    `json:"album"`
	Pay              string    `json:"pay"`
	Artistid         int       `json:"artistid"`
	Albumpic         string    `json:"albumpic"`
	Originalsongtype int       `json:"originalsongtype"`
	Songtimeminutes  string    `json:"songTimeMinutes"`
	Islistenfee      bool      `json:"isListenFee"`
	Pic120           string    `json:"pic120"`
	Name             string    `json:"name"`
	Online           int       `json:"online"`
}
type Data struct {
	Total int `json:"total"`
	List  []List `json:"list"`
}

func (k *MusicData) UnmarshalJSON(b []byte) error {
	regStr := "\"total\":([\\w+\"]+)"
	reg, err := regexp.Compile(regStr)
	if err != nil {
		return  err
	}
	str := reg.ReplaceAllStringFunc(string(b), func(s string) string {
		if len(s) == 0 {
			return s
		}

		var newStr string
		strArr := strings.Split(s, ":")
		switch reflect.ValueOf(strArr[1]).Kind() {
		case reflect.String:
			reg, _ := regexp.Compile(":\"(\\d+)\"")
			newStr = reg.ReplaceAllString(s, ":${1}")
		default:
			newStr = s
		}
		return newStr
	})

	type ms MusicData
	var kugouMusic ms
	if err := json.Unmarshal([]byte(str), &kugouMusic); err != nil {
		return err
	}
	*k = *(*MusicData)(unsafe.Pointer(&kugouMusic))
	return nil
}