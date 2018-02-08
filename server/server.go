package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func main() {
	InitHandler()
}

var (
	neighbors_map = make(map[string]*NeighborsPair)
	locker        sync.RWMutex
	port          = flag.String("port", ":8888", "listen port")
	heart_timeout = flag.Int64("heart_timeout", 30, "heart_timeout")
)

type NeighborsPair struct {
	id         string
	regionInfo []RegionInfo
}

type RegionInfo struct {
	info      string
	neighbors []string
	webWs     *websocket.Conn
}

func InitHandler() {
	log.Println("listen:", *port)
	http.Handle("/", websocket.Handler(phoneHandler))
	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Println("ListenAndServe err,", err)
		return
	}

}

func GetParam(ws *websocket.Conn, param string) string {
	return ws.Config().Location.Query().Get(param)
}
func phoneHandler(ws *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("phone handler fatal error: ", err)
		}
	}()
	info := GetParam(ws, "info")
	if info == "" {
		return
	}

	latitudeStr := GetParam(ws, "latitude")
	if latitudeStr == "" {
		return
	}
	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {

	}
	longitudeStr := GetParam(ws, "longitude")
	if longitudeStr == "" {
		return
	}
	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {

	}

	//获取geohash
	id, _ := Encode(latitude, longitude, 8)
	neighbors := GetNeighbors(latitude, longitude, 8)
	msg := "add =>" + info
	for _, neighborsId := range neighbors {
		neighborsWsSlice, ok := neighbors_map[neighborsId]
		if !ok {
			continue
		}
		for _, conns := range neighborsWsSlice.regionInfo {
			//发送给附近的设备
			conns.webWs.Write([]byte(msg))
			//发送给自己附近的设备信息
			ws.Write([]byte("add=>" + conns.info))
		}
	}

	locker.RLock()
	neighbors_pair, ok := neighbors_map[id]
	locker.RUnlock()
	if ok {
		//存在
		log.Printf("device id (%v) map exits", id)
		neighbors_pair.setWs(ws, info, neighbors)
	} else {
		//不存在
		log.Println("device id map not exist")
		locker.Lock()
		neighbors_map[id] = NewPhoneWebPair(id)
		locker.Unlock()
		neighbors_map[id].setWs(ws, info, neighbors)
	}
}

func NewPhoneWebPair(deviceId string) (pwpair *NeighborsPair) {
	return &NeighborsPair{id: deviceId}
}

//设置手机端websocket
func (pwpair *NeighborsPair) setWs(ws *websocket.Conn, info string, neighbors []string) {
	defer func() {
		pwpair.clearMap()
		if err := recover(); err != nil {
			log.Println("set websocket fatal error:", err)
		}
	}()
	log.Printf("id:(%s) side connected", pwpair.id)

	var regionInfo RegionInfo

	regionInfo.info = info
	regionInfo.webWs = ws
	regionInfo.neighbors = append(regionInfo.neighbors, neighbors...)
	pwpair.regionInfo = append(pwpair.regionInfo, regionInfo)
	var msg string
	for {
		//接收web  端发来的信息
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			if err != io.EOF {
				// 一般都是超时
				log.Printf("web id: %v read heartbeat timeout error(%v)", pwpair.id, err)
			} else {
				log.Printf("web id: %v client connection close error(%v)", pwpair.id, err)
			}
			pwpair.closeWs(ws)
			return
		}
		// 判断是否是心跳
		if msg == "h" {
			log.Println("receive web heartbeat:", pwpair.id)
			if !pwpair.sendWsHeartbeatBack(ws) {
				return
			}
		}
	}
}

//清理map
func (pwpair *NeighborsPair) clearMap() {
	if len(pwpair.regionInfo) == 0 {
		locker.Lock()
		delete(neighbors_map, pwpair.id)
		locker.Unlock()
		pwpair = nil
	}
}

// 关闭 ws
func (pwpair *NeighborsPair) closeWs(ws *websocket.Conn) {
	for k, v := range pwpair.regionInfo {
		if v.webWs == ws {
			v.webWs = nil
			pwpair.regionInfo = append(pwpair.regionInfo[:k], pwpair.regionInfo[k+1:]...)
			//并且通知附近的人该设备离开
			for _, neighborsId := range v.neighbors {
				neighborsWsSlice, ok := neighbors_map[neighborsId]
				if !ok {
					continue
				}
				for _, conns := range neighborsWsSlice.regionInfo {
					//发送给附近的设备
					msg := "del=>" + v.info
					conns.webWs.Write([]byte(msg))

				}
			}

		}
	}
	log.Println("close web socket,", pwpair.id)
	ws.Close()
	pwpair.clearMap()
}

// 发送心跳
func (pwpair *NeighborsPair) sendWsHeartbeatBack(ws *websocket.Conn) bool {
	webHeartbeatData := "h"
	if _, err := ws.Write([]byte(webHeartbeatData)); err != nil {
		log.Printf("web ws.Write() write heartbeat to client error(%v), id(%s)", err, pwpair.id)
		pwpair.closeWs(ws)
		return false
	}
	log.Println("send web heartbeat =>", pwpair.id)
	// 重新设置超时时间
	ws.SetReadDeadline(time.Now().Add(time.Second * time.Duration(*heart_timeout)))
	return true
}
