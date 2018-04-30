package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var lastGpuInfos []GPUinfo

//regexStrWatt := "(\\d+\\.\\d+) W, (\\d+\\.\\d+) W"
var regExAllGPUValues *regexp.Regexp
var regExHandleLoss *regexp.Regexp

func getLastGpuInfo(idx_from int, idx_to int) []GPUinfo {

	ret := make([]GPUinfo, idx_to-idx_from+1)

	// Copy data if exists
	for i := idx_from; i <= int(math.Min(float64(len(lastGpuInfos)-1), float64(idx_to))); i++ {

		ret[i-idx_from] = lastGpuInfos[i]
	}
	return ret
}

func parseInit() {

	regExAllGPUValues, _ = regexp.Compile("(\\d+), (0x[0-9A-za-z]+), (P\\d), " +
		"(\\d+) MHz, (\\d+) MHz, (\\d+\\.\\d+) W, (\\d+\\.\\d+) W, " +
		"(\\d+), (\\d+) %, (\\d+) MiB, (\\d+) MiB")
	regExHandleLoss, _ = regexp.Compile("Unable to determine the device handle for " +
		"GPU ([a-zA-z0-9:\\.]+): GPU is lost. Reboot the system to recover this GPU")
}

func parseNvOut(out string) []error {

	errors := []error{}
	var matches []string
	newGPUinfos := []GPUinfo{}

	// Look for "handlo loss" error in nvidia-smi output
	matches = regExHandleLoss.FindStringSubmatch(out)
	if matches != nil {
		err := NewErrGpu(-1, lastGpuInfos, strings.TrimSpace(out))
		errors = append(errors, err)
		return errors
	}
	// Split into lines
	outLines := strings.Split(out, "\n") //Last line is empty

	// "Index\tPci.Bus\tpState\tCore.Frequ\tMem.Freq\tPower.Draw\t""Power.Limit\tFan.Speed\tTemp\tMem.Total\tMem.Used

	for i := 0; i < len(outLines)-1; i++ {

		if strings.Contains(outLines[i], "error") {
			msg := "GPU" + "i" + "ecountered an error.\n" +
				"The follow data was read:\n" +
				outLines[i] + "\n"
			err := NewErrGpu(i, getLastGpuInfo(i, i), msg)
			errors = append(errors, err)
			continue
		}

		matches = regExAllGPUValues.FindStringSubmatch(outLines[i])

		tmp := GPUinfo{}

		if matches == nil {
			fmt.Println("Warning: Unable to parse line of nvidia-smi output: ", outLines[i])
		} else {
			tmp.ID, _ = strconv.Atoi(matches[1])
			tmp.PciBus = matches[2]
			tmp.PowerState = matches[3]
			tmp.ClockCore, _ = strconv.Atoi(matches[4])
			tmp.ClockMem, _ = strconv.Atoi(matches[5])
			tmp.PowerDraw, _ = strconv.ParseFloat(matches[6], 64)
			tmp.PowerLimit, _ = strconv.ParseFloat(matches[7], 64)
			tmp.Temp, _ = strconv.Atoi(matches[8])
			tmp.FanSpeed, _ = strconv.Atoi(matches[9])
			tmp.MemTotal, _ = strconv.Atoi(matches[10])
			tmp.MemUsed, _ = strconv.Atoi(matches[11])
		}
		newGPUinfos = append(newGPUinfos, tmp)

		if tmp.PowerDraw/tmp.PowerLimit < cfg.minPowerDrawPC/100 {
			msg := fmt.Sprintf("GPU %d is Idle.\n"+
				"Power consumpiton is %f out of %f", i, tmp.PowerDraw, tmp.PowerLimit)
			err := NewErrGpu(i, getLastGpuInfo(i, i), msg)
			errors = append(errors, err)
			continue
		}
	}
	if len(errors) == 0 {
		lastGpuInfos = newGPUinfos
	}
	return errors

}
