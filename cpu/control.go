// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package cpu

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// 🔥 เขียนค่า (ต้องใช้ root)
func writeFile(path string, value string) error {
	return os.WriteFile(path, []byte(value), 0644)
}

// 🧠 หา policy ทั้งหมด
func getPolicyPaths() []string {
	paths, _ := filepath.Glob("/sys/devices/system/cpu/cpufreq/policy*")
	return paths
}

//////////////////////////////////////////////////
// ⚡ SET MAX FREQ (GHz)
//////////////////////////////////////////////////

func SetMaxFreqGHz(ghz float64) error {
	khz := int(ghz * 1e6)
	val := strconv.Itoa(khz)

	policies := getPolicyPaths()
	if len(policies) == 0 {
		return fmt.Errorf("no cpufreq policy found")
	}

	for _, p := range policies {
		path := p + "/scaling_max_freq"

		if err := writeFile(path, val); err != nil {
			return fmt.Errorf("fail on %s: %v", path, err)
		}
	}

	return nil
}

//////////////////////////////////////////////////
// ⚡ SET MIN FREQ (GHz)
//////////////////////////////////////////////////

func SetMinFreqGHz(ghz float64) error {
	khz := int(ghz * 1e6)
	val := strconv.Itoa(khz)

	for _, p := range getPolicyPaths() {
		path := p + "/scaling_min_freq"

		if err := writeFile(path, val); err != nil {
			return err
		}
	}
	return nil
}

//////////////////////////////////////////////////
// 🧠 SET GOVERNOR
//////////////////////////////////////////////////

func SetGovernor(gov string) error {
	for _, p := range getPolicyPaths() {
		path := p + "/scaling_governor"

		if err := writeFile(path, gov); err != nil {
			return err
		}
	}
	return nil
}
