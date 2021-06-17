package main

import (
	"bufio"
	"fmt"
	"log"
	"music/model"
	"music/service"
	"music/utils"
	"os"
	"os/exec"
	"sync"
	"time"
)

func main()  {
	for {
		LOOP:
			var username string
			input := bufio.NewScanner(os.Stdin)
			log.Printf("请输入要下载歌曲人的名字:\n")
			for input.Scan() {
				username = input.Text()
				if len(username) > 0 {
					break
				}
				log.Printf("请输入要下载歌曲人的名字:\n")
			}

			basePath := "f:\\音乐"
			log.Printf("开始采集%s的歌曲,下载歌曲位置是%s\n", username, basePath)
			isSuccess, musicDir := utils.CreateDir(basePath, username)
			if !isSuccess {
				fmt.Println("创建下载目录失败")
				return
			}

		    var consumeThreadNum = 3
		    var productThreadNum = 5
			chanMusic := make(chan *model.ChanMusic, 30)
			chanMusicId := make(chan *model.ChanMusicId, 30)
			musicService := service.NewKugouMusicService(chanMusic, chanMusicId, productThreadNum, consumeThreadNum)
			var wg sync.WaitGroup

			musicService.ProductMusicUrl(username, musicDir, &wg)
			time.Sleep(time.Second*time.Duration(1))
			musicService.DownMusic(musicDir, &wg)
			wg.Wait()

			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			_ = cmd.Run()

			log.Printf("%s的歌曲已经采集完成,成功采集歌曲%d首,下载歌曲位置是%s\n", username, musicService.MusicTotal, basePath)
			goto LOOP
		}

	}
