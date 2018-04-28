package main

import (
	"fmt"
	"os/exec"
)

func main() {

	//var nvidia_smi_cmd_args []string

	nvidia_smi_cmd_args := []string{"C:\\Program Files\\NVIDIA Corporation\\NVSMI\\nvidia-smi", "--query-gpu=power.draw,power.limit", "--format=csv,noheader"}
	//"cmd", "/C",
	//fmt.Println("C:\\Program Files\\NVIDIA Corporation\\NVSMI\\nvidia-smi --query-gpu=power.draw,power.limit --format=csv,noheader\"")

	cmd := exec.Command(nvidia_smi_cmd_args[0], nvidia_smi_cmd_args[1:]...)

	// prog, arg0,arg1
	//cmd := exec.Command("cmd", "/C", "C:\\Program Files\\NVIDIA Corporation\\NVSMI\\nvidia-smi")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr

	//var outbuf, errbuf bytes.Buffer
	//cmd.Stdout = &outbuf
	//cmd.Stderr = &errbuf
	//err := cmd.Run()

	//stdout := outbuf.String()
	//stderr := errbuf.String()
	stdout, err := cmd.Output()
	stderr := ""

	fmt.Println("Hello, 世界")

	if err == nil {
		fmt.Println("Command run successfull")
	} else {
		fmt.Println("Command run failed")
		fmt.Println(err)
	}
	fmt.Println("StdOut:")
	fmt.Println(string(stdout))
	fmt.Println("StdErr:")
	fmt.Println(stderr)
}
