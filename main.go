package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readFreq(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return -1
	}
	val, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return val
}

func main() {
	min := readFreq("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_min_freq")
	max := readFreq("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq")

	fmt.Printf("Min: %.2f GHz\n", float64(min)/1e6)
	fmt.Printf("Max: %.2f GHz\n", float64(max)/1e6)
}
