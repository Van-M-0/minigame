package main

import (
	"net/http"
	"fmt"
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
		http.HandleFunc("/wechatLogin", hp.wechatLogin)

		if err := http.ListenAndServe(":11447", nil); err != nil {
			fmt.Println("http serve error ", err)
		}
	}()
}

func (hp *httpServer) wechatLogin(w http.ResponseWriter, r *http.Request) {

}




