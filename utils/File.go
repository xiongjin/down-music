package utils

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func CreateDir(basePath, username string) (flag bool, dirPath string) {
	musicDir := filepath.Join(basePath, username)
	fileFlag , _ := FileExists(musicDir)
	if !fileFlag {
		err := os.MkdirAll(musicDir, 0777 )
		if err != nil {
			fmt.Println(err)
			return false, ""
		}
		err = os.Chmod(musicDir, 0777)
		if err != nil {
			fmt.Println(err)
			return false, ""
		}
	}

	return true, musicDir
}


func CreateFile(basePath, musicName, musicUrl string) (bool, int) {
	sep := string(os.PathSeparator)
	filePath := basePath + sep + musicName +".mp3"
	isExist, _ := FileExists(filePath)
	if isExist {
		fmt.Printf("音乐名称:%s，已经下载过了\n", musicName)
		return true, 202
	}

	req := NewBrowser()
	str, code := req.Get(musicUrl)
	if len(str) == 0 || code != 200 {
		log.Printf("音乐名称:%s，歌曲下载失败，code值:%d， url:%s\n", musicName, code, musicUrl)
		return false, code
	}

	fp, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		fmt.Printf("音乐名称:%s，创建歌曲文件失败\n", musicName)
		return false, 200
	}

	defer fp.Close()

	_, err = fp.WriteString(string(str))
	if err != nil {
		fmt.Printf("音乐名称:%s，生成mp3失败\n", musicName)
		return false, 200
	}

	fmt.Printf("音乐名称:%s，下载成功\n", musicName)
	return true, 200
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}


func CreateRandomString(len int) string  {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0;i < len ;i++  {
		randomInt,_ := rand.Int(rand.Reader,bigInt)
		container += string(str[randomInt.Int64()])
	}
	return strings.ToUpper(container)
}

func FilterStr(str, filterStr string)  string {
	index := strings.Index(str, "(")
	if index != -1 {
		str = string([]byte(str)[:index])
	}
	index = strings.Index(str, "【")
	if index != -1 {
		str = string([]byte(str)[:index])
	}

	str = strings.TrimSpace(str)
	if len(filterStr) > 0 {
		str = strings.Replace(str, filterStr, "", -1)
		str = strings.Replace(str, "-", "", -1)
	}

	re := regexp.MustCompile("[?|(\\\\)|(/)|(\\s)]")
	if re.MatchString(str) {
		str = re.ReplaceAllString(str, "")
	}

	return str
}


func GetRealUsername(username string) string {
	s := []rune(username)
	var str2 string
	var flag = false
	for i:= 0; i<len(s); i++ {
		if s[i] < 255 {
			continue
		}

		str2 += string(s[i])
		if flag == false {
			flag = true
		}
	}

	if flag == false {
		return username
	}
	return str2
}

