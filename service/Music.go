package service

import (
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"music/model"
	"music/utils"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type KugouMusicService struct {
	ChanMusic chan *model.ChanMusic
	ChanMusicId chan *model.ChanMusicId
	ProductThreadNum int
	ConsumeThreadNum int
	MusicTotal int
}

func NewKugouMusicService(chanMusic chan *model.ChanMusic, chanMusicId chan *model.ChanMusicId, productThreadNum, consumeThreadNum int) *KugouMusicService {
	service := KugouMusicService{ChanMusic: chanMusic, ChanMusicId: chanMusicId, ProductThreadNum: productThreadNum, ConsumeThreadNum: consumeThreadNum }
	return &service
}

func(m *KugouMusicService) GetMusiciansId(username string)  int {
	var artistId int
	var webUrl string
	var pageSize = 30

	baseUrl := "https://www.kuwo.cn/api/www/search/searchMusicBykeyWord"
	req := utils.NewBrowser()
	paramsMap := map[string]string{
		"key" : username,
		"httpsStatus" : "1",
		"pn" : "1",
		"rn" : strconv.Itoa(pageSize),
		"reqId" : fmt.Sprintf("%s", uuid.NewV4()),
	}
	webUrl = baseUrl + "?" +  req.EncodeParams(paramsMap)
	str, code := req.Get(webUrl)

	if len(str) == 0  || code != 200 {
		fmt.Println(string(str))
		return  artistId
	}

	var kugouMusic model.MusicData
	if err := json.Unmarshal(str, &kugouMusic); err != nil {
		fmt.Println(err)
		return artistId
	}

	for _, v := range kugouMusic.Data.List {
		v.Artist = utils.GetRealUsername(v.Artist)
		if strings.ToLower(v.Artist) ==  strings.ToLower(username) {
			artistId = v.Artistid
			break
		}
	}
	return  artistId
}

func(m *KugouMusicService) GetMusicIdList(username string) int {
	var musicTotal int
	defer close(m.ChanMusicId)
	artistId := m.GetMusiciansId(username)
	if artistId == 0 {
		fmt.Println("获取音乐人ID失败")
		return musicTotal
	}

	var webUrl string
	var pageSize = 30
	baseUrl := "https://www.kuwo.cn/api/www/artist/artistMusic"
	req := utils.NewBrowser()
	paramsMap := map[string]string{
		"artistid" : strconv.Itoa(artistId),
		"httpsStatus" : "1",
		"pn" : "1",
		"rn" : strconv.Itoa(pageSize),
		"reqId" : fmt.Sprintf("%s", uuid.NewV4()),
	}
	webUrl = baseUrl + "?" +  req.EncodeParams(paramsMap)
	str, code := req.Get(webUrl)
	if len(str) == 0  || code != 200 {
		fmt.Println(string(str))
		return  musicTotal
	}

	var kugouMusic model.MusicData
	if err := json.Unmarshal(str, &kugouMusic); err != nil {
		fmt.Println(err)
		fmt.Println(string(str))
		return  musicTotal
	}

	musicTotal = kugouMusic.Data.Total
	if  musicTotal == 0 {
		return  musicTotal
	}

	if len(kugouMusic.Data.List) > 0 {
		for _, v := range kugouMusic.Data.List {
			m.ChanMusicId <- &model.ChanMusicId{Name: v.Name , Id: v.Rid}
		}
	}

	f1 := decimal.NewFromFloat(float64(musicTotal))
	f2 := decimal.NewFromFloat(float64(pageSize))
	maxPage := int(f1.Div(f2).Ceil().IntPart())

	for page := 2; page <= maxPage; page++ {
		paramsMap["pn"] = strconv.Itoa(page)
		paramsMap["reqId"] = utils.CreateRandomString(32)
		webUrl = baseUrl + "?" +  req.EncodeParams(paramsMap)
		var kugouMusic model.MusicData

		for tryNum := 0; tryNum<=2; tryNum++ {
			str, code := req.Get(webUrl)
			if len(str) == 0  || code != 200 {
				time.Sleep(time.Second*time.Duration(1))
				continue
			}

			if err := json.Unmarshal(str, &kugouMusic); err != nil {
				time.Sleep(time.Second*time.Duration(1))
				continue
			}

			if len(kugouMusic.Data.List) > 0 {
				for _, v := range kugouMusic.Data.List {
					m.ChanMusicId <- &model.ChanMusicId{Name: v.Name , Id: v.Rid}
				}
				break
			}

		}
		time.Sleep(time.Second*time.Duration(1))
	}
	return musicTotal
}

func(m *KugouMusicService) ProductMusicUrl(username string, musicDir string, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		 m.GetMusicIdList(username)
		 wg.Done()
	}()

	isStopChan := make(chan struct{})
	var once sync.Once
	wg.Add(1)
	go func() {
		defer  wg.Done()
		t1 := time.NewTimer(time.Second * 3)
		for {
			select {
				case <- isStopChan:
					close(m.ChanMusic)
					close(isStopChan)
					return
				case <- t1.C:
			}
		}
	}()

	for i := 0; i< m.ProductThreadNum; i++ {
		wg.Add(1)
		go func() {
			for {
				if music,  ok :=  <- m.ChanMusicId; ok {
					if len(music.Name) > 30 {
						continue
					}

					music.Name = utils.FilterStr(music.Name, username)
					sep := string(os.PathSeparator)
					filePath := musicDir + sep + music.Name +".mp3"
					isExist, _ := utils.FileExists(filePath)
					if isExist {
						fmt.Printf("音乐名称:%s，已经下载过了\n", music.Name)
						continue
					}

					s := strings.ToLower(music.Name)
					var isFilter bool
					filterStrList := []string{"live", "cover", "+"}
					for _, filterStr := range filterStrList {
						if strings.Contains(s,filterStr) {
							isFilter = true
							break
						}
					}
					if isFilter {
						continue
					}

					url := m.GetMusicDownUrl(music.Id)
					if len(url) > 0 {
						m.ChanMusic <- &model.ChanMusic{Name: music.Name, Url: url}
					}

					time.Sleep(time.Second*time.Duration(1))

				} else {
					once.Do(func() {
						isStopChan <- struct{}{}
					})
					break
				}
			}
			wg.Done()
		}()
	}
}

func(m *KugouMusicService) DownMusic(musicDir string, wg *sync.WaitGroup  ) {
	for i := 0; i< m.ConsumeThreadNum; i++ {
		wg.Add(1)
		go func() {
			for {
				if music,  ok :=  <- m.ChanMusic; ok {
					var isSuccess = false
					var code = 0

					for tryNum := 0; tryNum <= 2; tryNum++ {
						isSuccess, code = utils.CreateFile(musicDir, music.Name, music.Url)
						if !isSuccess && code != 404 {
							time.Sleep(time.Second*time.Duration(1))
							continue
						}

						if isSuccess && code == 200 {
							m.MusicTotal++
							time.Sleep(time.Second*time.Duration(1))
							break
						}

						time.Sleep(time.Second*time.Duration(1))
					}

				} else {
					break
				}
			}
			defer  wg.Done()
		}()
	}
}

func(m *KugouMusicService) GetMusicDownUrl(musicId int) string {
	var url string

	for tryNum := 0; tryNum <= 2; tryNum++ {
		baseUrl := "http://api.4dn.net/kuwo/kw.php?id="+strconv.Itoa(musicId)+"&km=320"
		req := utils.NewBrowser()
		str, code := req.Get(baseUrl)
		if len(str) == 0  || code != 200 {
			continue
		}

		var musicUrl model.MusicUrl
		if err := json.Unmarshal(str, &musicUrl); err != nil {
			continue
		}

		if len(musicUrl.URL) > 0 {
			url = musicUrl.URL
		}

		break
	}
	return  url
}

