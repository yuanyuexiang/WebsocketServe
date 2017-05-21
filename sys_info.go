package main

import (
	// "fmt"
	// "math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var sysInfoChan chan string

func init() {
	sysInfoChan = make(chan string)
	tick := time.NewTicker(2 * time.Second)
	tickSave := time.NewTicker(30 * time.Second)
	var m runtime.MemStats
	file, err := os.OpenFile("sys_info", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			<-tick.C

			if len(sysController.clients) == 0 {
				continue
			}

			runtime.ReadMemStats(&m)
			// alloc := math.Float64frombits(m.Alloc) / 1024 / 1024
			// sys := math.Float64frombits(m.Sys) / 1024 / 1024
			info := "time:" + strconv.FormatInt(time.Now().Unix(), 10)
			// info += "alloc:" + strconv.FormatFloat(alloc, 'f', 3, 32)
			// info += " sys:" + strconv.FormatFloat(sys, 'f', 3, 32)
			info += " alloc:" + strconv.FormatUint(m.Alloc, 10)
			info += " sys:" + strconv.FormatUint(m.Sys, 10)
			info += " numGoroutine:" + strconv.Itoa(runtime.NumGoroutine())
			info += " numGc:" + strconv.FormatUint(uint64(m.NumGC), 10)
			// info += " pauseTotalNs:" + FormatUint(m.PauseTotalNs, 10)

			// users data
			info += " numClient:" + strconv.Itoa(len(Controller.Clients))
			info += " numUser:" + strconv.Itoa(len(Controller.Users))
			// redis pool
			poolUseNum := make([]string, len(RConn.connPool.conns))
			for i, conn := range RConn.connPool.conns {
				poolUseNum[i] = strconv.FormatInt(conn.usedNum, 10)
			}
			str := "\"[" + strings.Join(poolUseNum, ",") + "]\""
			info += " redisPool:" + str + "\n"
			sysInfoChan <- info
			// fmt.Print(info)
			// runtime.GC()
		}
	}()

	go func() {
		for {
			<-tickSave.C
			// time, alloc, sys, numGoroutine, numGc, pauseTotalNs, numClient, numUsers, redisPool
			runtime.ReadMemStats(&m)
			// alloc := math.Float64frombits(m.Alloc) / 1024 / 1024
			// sys := math.Float64frombits(m.Sys) / 1024 / 1024
			info := strconv.FormatInt(time.Now().Unix(), 10)
			// info += "," + strconv.FormatFloat(alloc, 'f', 3, 32)
			// info += "," + strconv.FormatFloat(sys, 'f', 3, 32)
			info += "," + strconv.FormatUint(m.Alloc, 10)
			info += "," + strconv.FormatUint(m.Sys, 10)
			info += "," + strconv.Itoa(runtime.NumGoroutine())
			info += "," + strconv.FormatUint(uint64(m.NumGC), 10)
			// info += "," + strconv.FormatUint(m.PauseTotalNs, 10)
			// users data
			info += "," + strconv.Itoa(len(Controller.Clients))
			info += "," + strconv.Itoa(len(Controller.Users))
			// redis pool
			poolUseNum := make([]string, len(RConn.connPool.conns))
			for i, conn := range RConn.connPool.conns {
				poolUseNum[i] = strconv.FormatInt(conn.usedNum, 10)
			}
			str := strings.Join(poolUseNum, " ")
			info += "," + str + "\n"

			file.Write([]byte(info))
		}
	}()

}
