// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package cpu

import (
	"path/filepath"
)

func GetCurrentFreqs() []float64 {
	files, _ := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq/scaling_cur_freq")

	var result []float64

	for _, f := range files {
		val, err := readInt(f)
		if err != nil {
			result = append(result, 0)
			continue
		}
		result = append(result, float64(val)/1e6)
	}

	return result
}
