package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type httpServer struct {
	w 		*watchdog
}


func newHttpServer() *httpServer {
	hp := &httpServer{}
	return hp
}

func (hp *httpServer) start() {
	go func() {
		if err := http.ListenAndServe(":11447", nil); err != nil {
			fmt.Println("http serve error ", err)
		}
	}()
}

var (
	appId = "wxd93455aa46721ce7"
	secret = "eb1004e1f10358e2934b1a80a8de8032"
)

func (hp *httpServer) wechatLogin(code string) (string, string, string, string, string, int){
	type access struct {
		Appid	string		`json:"appid"`
		Secret 	string		`json:"secret"`
		Code 	string		`json:"code"`
		GrantType string	`json:"grant_type"`
	}

	fmt.Println("client wechat login ", code)

	request := "appid=" + appId + "&"+
		"secret=" + secret + "&" +
		"code=" + code + "&" +
		"grant_type=authorization_code"

	type response struct {
		AccToken 		string 	`json:"access_token"`
		ExpiresIn		int 	`json:"expires_in"`
		RefToken 		string  `json:"refresh_token"`
		OpenId 			string 	`json:"openid"`
		Scope 			string  `json:"scope"`
		SnsapiUserInfo 	string 	`json:"snsapi_userinfo"`
		Unionid 		string 	`json:"unionid"`
	}

	var errCode, AccToken, OpenId string
	errCode = "err"
	hp.get2("https://api.weixin.qq.com/sns/oauth2/access_token", request, true, func(suc bool, d interface{}) {
		//hp.get2("https://api.weixin.qq.com/sns/userinfo", string(d), true, func(suc bool, data interface{}) {
		//})
		if suc {
			data := d.([]byte)
			var r response
			err := json.Unmarshal(data, &r)
			if err == nil {
				errCode = "ok"
				AccToken = r.AccToken
				OpenId = r.OpenId
			} else {
				errCode = "openid"
				fmt.Println("wechatlogin error")
				fmt.Println(request)
				fmt.Println(d)
				fmt.Println(err)
			}
		}
	})

	fmt.Println("wechat client access retcode ", errCode)
	if errCode != "ok" {
		return errCode, "", "", "", "", -1
	}

	type responseUser struct {
		OpenId 			string 		`json:"openid"`
		NickName 		string		`json:"nickname"`
		Sex 			int			`json:"sex"`
		HeadImg 		string		`json:"headimgurl"`
	}

	request = "access_token=" + AccToken + "&"+
		"openid=" + OpenId

	var nickName, headImg string
	var sex int

	hp.get2("https://api.weixin.qq.com/sns/userinfo", request, true, func(suc bool, d interface{}) {
	//hp.get2("https://api.weixin.qq.com/cgi-bin/user/info", request, true, func(suc bool, d interface{}) {
			//hp.get2("https://api.weixin.qq.com/sns/userinfo", string(d), true, func(suc bool, data interface{}) {
			//})
			if suc {
				data := d.([]byte)
				var r responseUser
				err := json.Unmarshal(data, &r)
				if err == nil {
					errCode = "ok"
					nickName = r.NickName
					headImg = r.HeadImg
					sex = r.Sex
				} else {
					errCode = "userinfo"
					fmt.Println("wechat login get user info")
					fmt.Println(request)
					fmt.Println(d)
					fmt.Println(err)
				}
			}
		})

	fmt.Println("wechat client userinfo err code", errCode)
	return errCode, AccToken, OpenId, nickName, headImg, sex
}

func (hp *httpServer) get2(url string, content string, bHttps bool, cb func(bool, interface{})) {
	request := url + "?" + content
	fmt.Println("http request ", request)
	res, err := http.Get(request)

	if res.StatusCode == 200 {
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println("get ", request, res, err, string(body))
		cb(true, body)
	} else {
		fmt.Println("http status errcode ", res.StatusCode)
		cb(false, nil)
	}
}


