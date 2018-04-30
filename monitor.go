package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var cfg *Config

// Needs Golang 1.7 for context

func main() {

	fmt.Println("\nStarting Nvidia GPU monitor")

	cfg = &Config{}
	err := cfg.init()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		terminate()
	}

	parseInit()

	//see https://medium.com/@vCabbage/go-timeout-commands-with-os-exec-commandcontext-ba0c861ed738 for timeout

	out, nvErr := queryNV()

	// We want to check the context error to see if the timeout was executed.
	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.

	switch nvErr {
	case context.DeadlineExceeded:
		fmt.Println("GPU-Error: nvidia-smi timed out after ", cfg.nvTimeout.Seconds(), " seconds.")
	case nil: // No error
	default:
		fmt.Println("nvidia-smi could not be started.")
		fmt.Println(err.Error())
		terminate()
	}

	//Parse output

	errors := parseNvOut(out)

	if len(errors) > 0 {

		fmt.Println("Errors found:")

		for _, el := range errors {
			fmt.Println(el.Error())
		}
		onError()

	}

}

func searchErrOut(nvOut string) {

}

func queryNV() (string, error) {

	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), cfg.nvTimeout)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	fmt.Println(time.Now()," Exec nv smi command")
	cmd := exec.Command(cfg.nvidia_smi_cmd_args[0], cfg.nvidia_smi_cmd_args[1:]...)
	fmt.Println(time.Now()," After exec nv smi command")

go func() {
		select {
		case <-ctx.Done():
		fmt.Println(time.Now(),"Timeout:", string(time.Now().Format("05.00")))
		err := ctx.Err()
		if err != nil {
			fmt.Println(time.Now(),"in <-ctx.Done(): ", err)
		}
fmt.Println(time.Now(),"Before wait")	

fmt.Println(time.Now(),"After wait")
	break 
	}
	}()

	outByte, errOut := cmd.CombinedOutput()
fmt.Println(time.Now()," After read output")
	if errOut != nil {
		fmt.Println("running nvidia-smi failed")
		return "", errOut
	} else {
		fmt.Println("nvidia-smi run successfull")
	}
	// Convert to UNIX-style EOL
	out := strings.Replace(string(outByte), "\r", "", -1)
	fmt.Println("nvidia-smi raw cmbined output:")
	fmt.Println(out)

	return out, ctx.Err()

}

func terminate() {

	fmt.Println("Terminating due to internal error.")
	os.Exit(1)
}

func onError() {

	execCmd := strings.Split(cfg.onErrorExec, ";")

	cmd := exec.Command(execCmd[0], execCmd[1:]...)
	err := cmd.Run()

	if err != nil {

		if _, ok := err.(*exec.ExitError); ok {
			os.Stderr.WriteString("'" + cfg.onErrorExec + "'" + " executed not successfull.")
		} else {
			os.Stderr.WriteString("'" + cfg.onErrorExec + "'" + " could not be started.")
		}

	}

	os.Exit(0)
}
