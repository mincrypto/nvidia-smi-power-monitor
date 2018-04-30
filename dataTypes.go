package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"
)

type Config struct {
	nvCommand           string
	nvidia_smi_cmd_args []string
	nvTimeout           time.Duration
	nvPeriod            time.Duration
	onErrorExec         string
	minPowerDrawPC      float64
}

func (c *Config) init() error {

	//see https://gobyexample.com/command-line-flags for flag package use
	// Mandatory Arguments
	flag.StringVar(&c.nvCommand, "nvCommand", "", "Command to execute nvidia-smi without arguments. type: string")

	//Optional Arguments
	flag.DurationVar(&c.nvTimeout, "nvTimeout", 60*time.Second, "nvidia-smi timeout as Duration, e.g. 3s for 3 seconds. Type: string. Default: 60s")
	flag.DurationVar(&c.nvPeriod, "nvPeriod", 60*time.Second, "Time between nvidia-smi calls. Type: string. Default: 60s")
	flag.StringVar(&c.onErrorExec, "onErrorExec", "", "Execute command on error. Seperate comamnd and arguments by \";\" ! Type: string. Default: Do nothing, only report")
	flag.Float64Var(&c.minPowerDrawPC, "minPowerUse", 50, "Idle Treshold for GPU power-draw readings in percent. Type: integer. Default: 50")

	flag.Parse()

	return c.validate()

}

func (c *Config) validate() error {

	// Validate config
	if c.nvCommand == "" {
		return errors.New("Error: Argument nvCommand missing")
	}

	c.nvidia_smi_cmd_args = []string{c.nvCommand, "--query-gpu=index,pci.bus,pstate,clocks.sm,clocks.mem,power.draw,power.limit,temperature.gpu,fan.speed,memory.total,memory.used", "--format=csv,noheader"}
	fmt.Println("nvidia-smi command is: ", strings.Join(c.nvidia_smi_cmd_args, " "))

	fmt.Printf("nvidia-smi timeout set to %f seconds \n", c.nvTimeout.Seconds())
	fmt.Printf("nvidia-smi period set to  %f seconds \n", c.nvPeriod.Seconds())

	if c.minPowerDrawPC < 0 || c.minPowerDrawPC > 100 {
		return errors.New("Error:GPU power-draw idle threshold must be between 0 and 100")
	}
	fmt.Printf("Power draw idle threshold set to  %f%% \n", c.minPowerDrawPC)

	return nil
}

type GPUinfo struct {
	ID         int
	PciBus     string
	PowerState string
	ClockCore  int
	ClockMem   int
	PowerDraw  float64
	PowerLimit float64
	Temp       int
	FanSpeed   int
	MemTotal   int
	MemUsed    int
}

type ErrGpu struct {
	IdGPU     int
	Infos     []GPUinfo
	Timestamp time.Time
	message   string
}

func NewErrGpu(idGPU int, infos []GPUinfo, message string) *ErrGpu {
	return &ErrGpu{
		IdGPU:     idGPU,
		Infos:     infos,
		Timestamp: time.Now(),
		message:   message,
	}
}

func (e *ErrGpu) Error() string {
	return e.message + "\n" + e.printGpuInfo()
}

func (e *ErrGpu) printGpuInfo() string {

	buf := new(bytes.Buffer)

	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(buf, 0, 8, 0, '\t', 0)

	fmt.Fprintln(w, "Index\tPci.Bus\tpState\tCore.Frequ\tMem.Freq\tPower.Draw\t"+
		"Power.Limit\tFan.Speed\tTemp\tMem.Total\tMem.Used")
	for _, info := range e.Infos {

		fmt.Fprintf(w, "%d\t%s\t%s\t", info.ID, info.PciBus, info.PowerState)
		fmt.Fprintf(w, "%d\t%d\t%f\t", info.ClockCore, info.ClockMem, info.PowerDraw)
		fmt.Fprintf(w, "%f\t%d\t%d\t", info.PowerLimit, info.FanSpeed, info.Temp)
		fmt.Fprintf(w, "%d\t%d", info.MemTotal, info.MemUsed)
		fmt.Fprintln(w, "")
	}
	w.Flush()
	return fmt.Sprintln("Last known GPU state:") + buf.String()
}
