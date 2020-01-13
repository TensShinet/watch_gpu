package controllers

import (
	"encoding/json"
	"sync"

	"github.com/astaxie/beego"
)

type GpuController struct {
	beego.Controller
}

type GpuProcess struct {
	GPU         int
	PID         uint
	Name        string
	MemoryUsed  uint64
	Type        uint
	MemoryUsage int
}

type Machine struct {
	HostName  string
	Processes []GpuProcess
}

type Command struct {
	KillList []uint
	AutoKill bool
}

// 维护两个内存 map
// HostName 必须独一无二

type MachineList struct {
	allMachineList    map[string]Machine
	deleteProcessList map[string][]uint
	autoKill          map[string]bool
	mutexMachineList  sync.RWMutex
	mutexProcessList  sync.RWMutex
	mutexAutoKill     sync.RWMutex
}
type valueTypeError struct {
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}

func (this *MachineList) deleteOneMachine(hostname string) {
	this.mutexMachineList.Lock()
	_, ok := this.allMachineList[hostname]
	if ok {
		delete(this.allMachineList, hostname)
	}
	this.mutexMachineList.Unlock()
}

func (this *MachineList) updateOneMachine(hostname string, Processes []GpuProcess) {
	this.mutexMachineList.Lock()
	this.allMachineList[hostname] = Machine{hostname, Processes}
	this.mutexMachineList.Unlock()
}

func (this *MachineList) deleteOneMachineProcess(hostname string) {
	this.mutexProcessList.Lock()
	_, ok := this.deleteProcessList[hostname]
	if ok {
		delete(this.deleteProcessList, hostname)
	}
	this.mutexProcessList.Unlock()
}

func (this *MachineList) updateOneMachineProcess(hostname string, pids []uint) {
	this.mutexProcessList.Lock()
	this.deleteProcessList[hostname] = pids
	this.mutexProcessList.Unlock()
}
func (this *MachineList) getProcessList(hostname string, GC *GpuController, command *Command) {
	this.mutexProcessList.RLock()
	command.KillList = this.deleteProcessList[hostname]
	this.mutexProcessList.RUnlock()
}

func (this *MachineList) getMachineList(GC *GpuController) {
	this.mutexMachineList.RLock()
	GC.Data["json"] = ML.allMachineList
	this.mutexMachineList.RUnlock()
}

var ML MachineList = MachineList{}

func (this *MachineList) setAutokill(hostname string, value bool) {
	this.mutexAutoKill.Lock()
	this.autoKill[hostname] = value
	this.mutexAutoKill.Unlock()
}

func (this *MachineList) deleteAutokill(hostname string) {
	this.mutexAutoKill.Lock()
	_, ok := this.autoKill[hostname]
	if ok {
		delete(this.autoKill, hostname)
	}
	this.mutexAutoKill.Unlock()
}

func (this *MachineList) getAutokill(hostname string, command *Command) {
	this.mutexAutoKill.Lock()
	_, ok := this.autoKill[hostname]
	if ok {
		command.AutoKill = true
	} else {
		command.AutoKill = false
	}
	this.mutexAutoKill.Unlock()
}

func init() {

	ML.allMachineList = map[string]Machine{}
	ML.deleteProcessList = map[string][]uint{}
	ML.autoKill = map[string]bool{}
}

func (this *GpuController) Post() {
	var m Machine
	var err error
	if err = json.Unmarshal(this.Ctx.Input.RequestBody, &m); err != nil {
		this.Data["json"] = err.Error()
	} else {
		// 更新所有
		ML.updateOneMachine(m.HostName, m.Processes)
		// 发送要删除的 pids
		command := Command{}
		ML.getProcessList(m.HostName, this, &command)
		ML.getAutokill(m.HostName, &command)
		this.Data["json"] = command
		// 删除 pids
		ML.deleteOneMachineProcess(m.HostName)

	}
	this.ServeJSON()
}

func (this *GpuController) Get() {
	ML.getMachineList(this)
	this.ServeJSON()
}

func (this *GpuController) Delete() {
	t := this.GetString("type")
	if t == "KILLONE" {
		hostname := this.GetString("hostname")
		pid, err := this.GetInt("PID")

		if err != nil {
			this.Data["json"] = valueTypeError{
				"valueError",
				400,
			}
			this.ServeJSON()
			return
		}

		pids := []uint{uint(pid)}
		ML.updateOneMachineProcess(hostname, pids)
	} else if t == "AUTOKILL" {
		hostname := this.GetString("hostname")
		value, err := this.GetBool("value")
		if err != nil {
			this.Data["json"] = valueTypeError{
				"valueError",
				400,
			}
			this.ServeJSON()
			return
		}

		ML.setAutokill(hostname, value)
	} else {

	}
}
