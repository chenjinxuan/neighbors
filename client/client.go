package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	//"time"
	. "github.com/robfig/config"
)

var (
	c         *Conf
	wsAddr    string
	origin    string
	heartbeat string
	latitude  string
	longitude string
	info      string
)

func main() {
	initConfig()
	url := "%s/?latitude=%s&longitude=%s&info=%s"
	url = fmt.Sprintf(url, wsAddr, latitude, longitude, info)
	log.Println(url)
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	//如果客户端不发送心跳,,服务端也不会设置设置心跳机制
	//go func(ws *websocket.Conn) {
	//for {
	//   _, err = ws.Write([]byte("h"))
	//   if err != nil {
	//log.Fatal(err)
	//   }
	//   fmt.Printf("Send: %s\n", "h")
	//   time.Sleep(30 * time.Second)
	//}
	//}(ws)

	var msg = make([]byte, 512)
	for {
		m, err := ws.Read(msg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Receive: %s\n", msg[:m])
	}
}

func initConfig() {
	wsAddr, _ = GetCfg().String("Default", "wsAddr")
	origin, _ = GetCfg().String("Default", "origin")
	heartbeat, _ = GetCfg().String("Default", "heartbeat")
	latitude, _ = GetCfg().String("Default", "latitude")
	longitude, _ = GetCfg().String("Default", "longitude")
	info, _ = GetCfg().String("Default", "info")
}

type Conf struct {
	*Config
}



func GetCfg() *Conf {
	var cfg *Config
	if cfg == nil {
		_cfg, _ := ReadDefault("config.cfg")
		cfg = _cfg
	}
	c = &Conf{}
	c.Config = cfg
	return c
}
