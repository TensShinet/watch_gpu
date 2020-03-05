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
	KillList  []uint
	AutoKill  bool
}

type Command struct {
	KillList []uint
	AutoKill bool
}

// 维护两个内存 map
// HostName 必须独一无二
type machines struct {
	MachineMap map[string]*Machine
	mu         sync.Mutex
}

type valueTypeError struct {
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}

func (m *machines) deleteOneMachine(hostname string) {
	m.mu.Lock()
	_, ok := m.MachineMap[hostname]
	if ok {
		delete(m.MachineMap, hostname)
	}
	m.mu.Unlock()
}

func (m *machines) updateOneMachine(new *Machine) {
	m.mu.Lock()
	machine, ok := m.MachineMap[new.HostName]
	if ok {
		machine.Processes = new.Processes
	} else {
		m.MachineMap[new.HostName] = new
	}

	m.mu.Unlock()
}

func (m *machines) getKillList(hostname string) []uint {
	_, ok := m.MachineMap[hostname]
	if ok {
		return m.MachineMap[hostname].KillList
	}
	return nil
}

func (m *machines) deleteKillList(hostname string) {
	m.mu.Lock()
	machine, ok := m.MachineMap[hostname]
	if ok {
		machine.KillList = nil
	}
	m.mu.Unlock()
}

func (m *machines) checkAutoKill(hostname string) bool {
	_, ok := m.MachineMap[hostname]
	if ok {
		return m.MachineMap[hostname].AutoKill
	}
	return false
}

func (m *machines) updateOneMachineProcess(hostname string, pids []uint) {
	m.mu.Lock()
	machine, ok := m.MachineMap[hostname]
	if ok {
		machine.KillList = pids
	}
	m.mu.Unlock()
}
var allMachine machines

func (m *machines) setAutokill(hostname string, value bool) {
	m.mu.Lock()
	machine, ok := m.MachineMap[hostname]
	if ok {
		machine.AutoKill = value
	}
	m.mu.Unlock()
}

func init() {
	allMachine = machines{
		MachineMap: map[string]*Machine{},
		mu:         sync.Mutex{},
	}
}

func (gpuController *GpuController) Post() {
	var m Machine
	var err error
	if err = json.Unmarshal(gpuController.Ctx.Input.RequestBody, &m); err != nil {
		gpuController.Data["json"] = err.Error()
	} else {
		// 更新所有
		allMachine.updateOneMachine(&m)
		// 发送要删除的 pids
		command := Command{
			KillList: allMachine.getKillList(m.HostName),
			AutoKill: allMachine.checkAutoKill(m.HostName),
		}

		gpuController.Data["json"] = command
		// 删除 pids
		allMachine.deleteKillList(m.HostName)
	}
	gpuController.ServeJSON()
}

func (gpuController *GpuController) Get() {
	gpuController.Data["json"] = allMachine.MachineMap
	gpuController.ServeJSON()
}

func (gpuController *GpuController) Delete() {
	t := gpuController.GetString("type")
	if t == "KILLONE" {
		hostname := gpuController.GetString("hostname")
		pid, err := gpuController.GetInt("PID")

		if err != nil {
			gpuController.Data["json"] = valueTypeError{
				"valueError",
				400,
			}
			gpuController.ServeJSON()
			return
		}

		pids := []uint{uint(pid)}
		allMachine.updateOneMachineProcess(hostname, pids)
	} else if t == "AUTOKILL" {
		hostname := gpuController.GetString("hostname")
		value, err := gpuController.GetBool("value")
		if err != nil {
			gpuController.Data["json"] = valueTypeError{
				"valueError",
				400,
			}
			gpuController.ServeJSON()
			return
		}

		allMachine.setAutokill(hostname, value)
	} else {

	}
}
