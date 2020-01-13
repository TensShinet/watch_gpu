package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"./watch"
)

func initFlag() {
	flag.Parse()
	log.SetFlags(0)
}

type Machine struct {
	HostName  string
	Processes []watch.GpuProcess
}

func post(postData Machine, command *Command) (err error) {
	postUrl := "http://" + *addr + "/gpu_information"
	jsonValue, _ := json.Marshal(postData)
	resp, err := http.Post(postUrl, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Panicln("Error post GPU information ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, command)
	return nil
}

// 维护全局进程
var ProcessList map[string]int

func init() {
	ProcessList = make(map[string]int)
}
func autoKill(processes []watch.GpuProcess) {
	pl := make(map[string]int)
	for _, p := range processes {
		key := strconv.Itoa(int(p.GPU)) + strconv.Itoa(int(p.PID))

		if val, ok := ProcessList[key]; ok {
			if p.MemoryUsage >= *upbound {
				pl[key] = val + 1

				if val+1 == *times {
					err := syscall.Kill(int(p.PID), syscall.SIGKILL)
					if err != nil {
						log.Panic("Kill failed ", err)
					}
					log.Println("Killed the Process " + strconv.Itoa(int(p.PID)) + "Name: " + p.Name + "Memory usage: " + strconv.Itoa(p.MemoryUsage) + "%")
					_, ok := pl[key]
					if ok {
						delete(pl, key)
					}
				}
			}

		} else {
			pl[key] = 0
		}
	}
	ProcessList = pl
}

type Command struct {
	KillList []uint
	AutoKill bool
}

func kill(command *Command) {
	killList := command.KillList
	for _, pid := range killList {
		syscall.Kill(int(pid), syscall.SIGKILL)
	}
}

var addr = flag.String("addr", "None", "server address, if not set this, it will not post data to server")
var hostname = flag.String("hostname", "ts_client", "current hostname something like `gpu0`")
var upbound = flag.Int("upbound", 80, "the upbound of process gpu memoory usage, default 80.")
var interval = flag.Uint("interval", 3, "default 3, which means scan all gpu process every 3s")
var times = flag.Int("Times", 60, "the maximum number of consecutive times, default 60")

func main() {
	initFlag()

	if *hostname == "ts_client" {
		log.Fatalln("You must set hostname something like `gpu0`")
		return
	}

	fmt.Println("Sucess! get hostname is " + *hostname)
	for true {
		// 得到 gpu 所有信息
		processes, err := watch.GetAllRunningProcesses()
		if err != nil {
			log.Panicln("Error getting GetAllRunningProcesses ", err)
		}
		if *addr != "None" {
			postData := Machine{*hostname, processes}
			command := Command{}
			err = post(postData, &command)

			if err != nil {
				log.Panicln("Error get command ", err)
			}
			kill(&command)
		}

		autoKill(processes)
		time.Sleep(time.Duration(*interval) * 1000 * time.Millisecond)
		// break
	}

}
