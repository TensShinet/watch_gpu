package main

import (
	"bytes"
	"encoding/json"
	"github.com/TensShinet/watch_gpu/client/conf"
	"github.com/TensShinet/watch_gpu/client/logging"
	"github.com/TensShinet/watch_gpu/client/watch"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var logger = logging.GetLogger("client")
var machine Machine

func init() {
	level := logging.GetLevel(conf.GetString("LogLevel"))
	logger.SetLevel(level)
	machine = Machine{
		HostName:  "",
		AutoKill:  conf.GetBool("AutoKill"),
		Processes: nil,
		KillList:  nil,
	}
}

type Machine struct {
	HostName  string
	AutoKill  bool
	Processes []watch.GpuProcess
	KillList  []int
	Quit      bool
}

func post(postData Machine, command *Command) error {
	postUrl := "http://" + conf.GetString("Addr") + "/gpu_information"
	jsonValue, _ := json.Marshal(postData)
	resp, err := http.Post(postUrl, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		logger.WithError(err).Error("failed to post data")
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, command); err != nil {
		logger.WithError(err).Error("failed to receive data")
		return err
	}
	return nil
}

var ProcessMap map[int]int

func autoKill(processes []watch.GpuProcess) {
	update := make(map[int]int)
	for _, p := range processes {
		key := p.PID
		if val, ok := ProcessMap[key]; ok {
			if p.MemoryUsage <= conf.GetInt("Low") {
				update[key] = val + 1

				if val+1 == conf.GetInt("Times") {
					logger.WithField("pid", p.PID).Info("Kill Process")
					if err := kill(p.PID); err != nil {
						continue
					}
					_, ok := update[key]
					if ok {
						delete(update, key)
					}
					logger.WithField("pid", p.PID).Info("Process Killed")
				}
			}

		} else {
			update[key] = 0
		}
	}
	ProcessMap = update
}

type Command struct {
	KillList []int
	AutoKill bool
}

func killAll() {
	for _, pid := range machine.KillList {
		_ = kill(pid)
	}
}

func kill(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		logger.WithError(err).Error("failed to kill ", pid)
		return err
	}
	logger.Info("Success to kill ", pid)
	return nil
}

func main() {
	Hostname := conf.GetString("Hostname")
	Addr := conf.GetString("Addr")
	if Hostname == "" || Addr == "" {
		logger.WithField("Hostname", Hostname).WithField("server Addr", Addr).Error("config error in Hostname or Addr")
		return
	}
	logger.WithField("Hostname", Hostname).WithField("server address", Addr).Info("connect...")
	machine.HostName = Hostname
	SetupCloseHandler()
	for true {
		// 得到 gpu 所有信息
		processes, err := watch.GetAllRunningProcesses()
		if err != nil {
			logger.WithError(err).Error("failed to get all process")
			return
		}
		machine.Processes = processes
		command := Command{}
		err = post(machine, &command)
		if err != nil {
			return
		}
		machine.KillList = command.KillList
		machine.AutoKill = command.AutoKill

		killAll()
		if machine.AutoKill == true {
			logger.Info("AutoKill...")
			autoKill(processes)
		}
		time.Sleep(time.Duration(conf.GetInt("Interval")) * 1000 * time.Millisecond)
	}

}

func quit() {
	command := Command{}
	machine.Quit = true
	if err := post(machine, &command); err != nil {
		logger.WithError(err).Error("quit send failed")
	}
	killAll()
}

func SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		logger.Info("Ctrl C pressed in Terminal")
		quit()
		os.Exit(0)
	}()
}
