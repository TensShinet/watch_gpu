package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TensShinet/watch_gpu/client/conf"
	"github.com/TensShinet/watch_gpu/client/watch"
	"github.com/TensShinet/watch_gpu/client/logging"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"
)

var logger = logging.GetLogger("client")
func init() {
	level := logging.GetLevel(conf.GetString("LogLevel"))
	logger.SetLevel(level)
}


type Machine struct {
	HostName  string
	Processes []watch.GpuProcess
}

func post(postData Machine, command *Command) (err error) {
	postUrl := "http://" + conf.GetString("addr") + "/gpu_information"
	jsonValue, _ := json.Marshal(postData)
	resp, err := http.Post(postUrl, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Println("Error post GPU information ", err)
		return err
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
			if p.MemoryUsage <= conf.GetInt("low") {
				pl[key] = val + 1

				if val+1 == conf.GetInt("times") {
					logger.WithField("pid", int(p.PID)).Info("Kill Process")
					err := syscall.Kill(int(p.PID), syscall.SIGKILL)
					if err != nil {
						logger.WithError(err)
					}
					_, ok := pl[key]
					if ok {
						delete(pl, key)
					}
					logger.WithField("pid", int(p.PID))
				}
			}

		} else {
			pl[key] = 0
		}
	}
	ProcessList = pl
}

type Command struct {
	KillList []int
	AutoKill bool
}

func killAll(killlist []int) {
	for _, pid := range killlist {
		kill(pid)
	}
}

func kill (pid int) {
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		logger.WithError(err).Error("failed to kill ", pid)
	}
}


func main() {

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
				return
			}
			kill(&command)
		}

		autoKill(processes)
		time.Sleep(time.Duration(*interval) * 1000 * time.Millisecond)
		// break
	}

}
