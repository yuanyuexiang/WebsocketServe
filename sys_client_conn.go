package main

import (
	"github.com/gorilla/websocket"
	// "log"
	"net/http"
	"time"
)

type SysControl struct {
	clients map[string]*SysClientConn
}

type SysClientConn struct {
	id        string
	loadTime  int64
	ws        *websocket.Conn
	writeChan chan []byte
	writed    chan int
	quiter    chan int
	isClose   bool
}

var sysController *SysControl

func init() {
	sysController = &SysControl{}
	sysController.clients = make(map[string]*SysClientConn)
	go func() {
		for {
			sysmsg := <-sysInfoChan
			for _, cli := range sysController.clients {
				cli.writeChan <- []byte(sysmsg)
			}
		}
	}()
}

func NewSysClient(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	id := randStr()
	c := SysClientConn{}
	c.id = id
	c.writeChan = make(chan []byte, 100)
	c.writed = make(chan int)
	c.quiter = make(chan int)
	c.ws = ws
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	c.loadTime = time.Now().Unix()
	// add self to controller
	sysController.clients[c.id] = &c
	go c.run()
}

func (this *SysClientConn) Write(msg []byte) bool {
	if this.isClose {
		return false
	}
	this.writeChan <- msg
	r := <-this.writed
	if r == 1 {
		return true
	} else {
		return false
	}
}

func (this *SysClientConn) Close() {
	this.isClose = true
	this.ws.SetWriteDeadline(time.Now().Add(writeWait))
	this.ws.WriteMessage(websocket.CloseMessage, []byte{})
	this.quiter <- 0
}

func (this *SysClientConn) run() {
	go this.readMsg()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		this.isClose = true
		ticker.Stop()
		this.ws.Close()
	}()
	for {
		select {
		case msg, ok := <-this.writeChan:
			if !ok {
				return
			}
			this.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := this.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				// this.writed <- 0
				return
			}
			// this.writed <- 1
		case <-ticker.C:
			// log.Println("ping send msg id:", this.id)
			this.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := this.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-this.quiter:
			return
		}
	}
}

func (this *SysClientConn) readMsg() {
	defer func() {
		this.close()
	}()

	for {
		_, _, err := this.ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func (this *SysClientConn) close() {
	this.isClose = true
	close(this.writeChan)
	close(this.writed)
	close(this.quiter)
	delete(sysController.clients, this.id)
}
