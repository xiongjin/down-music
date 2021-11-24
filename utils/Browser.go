/**
1.可设置代理
2.可设置 cookie
3.自动保存并应用响应的 cookie
4.自动为重新向的请求添加 cookie
*/
package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	_ "strconv"
	"strings"
	"time"
)

type Browser struct {
	cookies []*http.Cookie
	headers map[string]string
	client *http.Client
}

//初始化
func NewBrowser() *Browser {
	hc := &Browser{}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	hc.client = &http.Client{Timeout : time.Duration(10)*time.Second, Transport:tr}
	randomStr :=  CreateRandomString(11)
	headers := map[string]string {
		"Accept-Language" : "zh-CN",
		"User-Agent" : "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36",
		"Referer" : " https://www.kuwo.cn/search/list",
		"Content-Type" : "application/x-www-form-urlencoded",
		"Host" : "www.769sc.com",
		"authority": "www.769sc.com",
		"origin": "https://www.769sc.com",
		"Connection": "keep-alive",
		"x-requested-with" : "XMLHttpRequest",
		"csrf" : randomStr,
	}
	hc.AddHeader(headers)

	cookies := map[string]string {
		"Hm_lvt_8bc816d5bc6d7326bf1502068c9ce6ec" : "1622770168",
		"Hm_lpvt_8bc816d5bc6d7326bf1502068c9ce6ec" : "1622771199",
		"kw_token" : randomStr,
	}
	hc.AddCookie(cookies)

	//hc.SetProxyUrl("socks5h://localhost:1080")

	//为所有重定向的请求增加cookie
	hc.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	//hc.client.Jar, _ = cookiejar.New(nil)
	return hc
}

//设置代理地址
func (self *Browser) SetProxyUrl(proxyUrl string)  {
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyUrl)
	}
	transport := &http.Transport{Proxy:proxy}
	self.client.Transport = transport
}

//获取当前所有的cookie
func (self *Browser) GetCookie() ([]*http.Cookie) {
	    return self.cookies
}

//设置请求cookie
func (self *Browser) AddCookie(cookies map[string]string)  {
	for k, v := range cookies {
		cookie := &http.Cookie{Name : k, Value : v}
		self.cookies = append(self.cookies, cookie)
	}
}

//设置请求header
func (self *Browser) AddHeader(headers map[string]string)  {
	self.headers = headers
}

//发送Get请求
func (self *Browser) Get(requestUrl string) ([]byte, int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	request, _ := http.NewRequest("GET", requestUrl, nil)
	self.setRequestCookie(request)
	self.setRequestHeader(request)
	response,err := self.client.Do(request)
	defer response.Body.Close()

	if err != nil {
		return []byte(""), response.StatusCode
	}

	data, _ := ioutil.ReadAll(response.Body)

	return data, response.StatusCode
}

//发送Get请求
func (self *Browser) GetLocationUrl(requestUrl string) (string, int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	var locationUrl string
	request, _ := http.NewRequest("GET", requestUrl, nil)
	self.setRequestCookie(request)
	self.setRequestHeader(request)
	response,err := self.client.Do(request)
	defer response.Body.Close()

	Redirect:
 	if response.StatusCode == 302 {
		locationUrl = response.Header.Get("location")
		if len(locationUrl) > 0 {
			request, _ := http.NewRequest("GET", locationUrl, nil)
			response,_ = self.client.Do(request)
			goto Redirect
		}

	}

	if err != nil {
		return "", response.StatusCode
	}

	return locationUrl, response.StatusCode
}

//发送Post请求
func (self *Browser) Post(requestUrl string, params map[string]string) ([]byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	postData := self.EncodeParams(params)
	request, _ := http.NewRequest("POST", requestUrl, strings.NewReader(postData))
	self.setRequestCookie(request)
	self.setRequestHeader(request)

	response, err := self.client.Do(request)
	defer response.Body.Close()

	if err != nil {
		return []byte("")
	}
	//保存响应的 cookie
	self.SetResponseCookie(response.Cookies())
	data, _ := ioutil.ReadAll(response.Body)
	return data
}

func (self *Browser) PostJson(requestUrl string, params map[string]interface{}) []byte {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	str, err := json.Marshal(params)
	if err != nil {
		return []byte("")
	}

	request, _ := http.NewRequest("POST", requestUrl, strings.NewReader(string(str)))
	self.headers["Content-Type"] = "application/json"
	self.setRequestCookie(request)
	self.setRequestHeader(request)

	response, _ := self.client.Do(request)
	defer response.Body.Close()

	//保存响应的 cookie
	self.SetResponseCookie(response.Cookies())
	data, _ := ioutil.ReadAll(response.Body)
	return data
}

//为请求设置 cookie
func (self *Browser) setRequestCookie(request *http.Request)  {
	for _,v := range self.cookies{
		request.AddCookie(v)
	}
}

func (self *Browser) SetResponseCookie(cookies []*http.Cookie)  {
	self.cookies = cookies
}

func (self *Browser) GetResponseCookie() []*http.Cookie {
	return self.cookies
}

func (self *Browser) setRequestHeader(request *http.Request)  {
	for k , v := range self.headers {
		request.Header.Set(k, v)
	}
}

//参数 encode
func (self *Browser) EncodeParams(params map[string]string) string {
	paramsData := url.Values{}
	for k,v := range params {
		paramsData.Set(k,v)
	}
	return paramsData.Encode()
}

