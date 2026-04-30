// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package cpu

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Policy struct {
	ID    string
	Cores []int
}

func GetPolicies() []Policy {
	dirs, _ := filepath.Glob("/sys/devices/system/cpu/cpufreq/policy*")

	var policies []Policy

	for _, d := range dirs {
		data, err := os.ReadFile(d + "/related_cpus")
		if err != nil {
			continue
		}

		fields := strings.Fields(string(data))
		var cores []int

		for _, f := range fields {
			val := 0
			fmt.Sscanf(f, "%d", &val)
			cores = append(cores, val)
		}

		policies = append(policies, Policy{
			ID:    d,
			Cores: cores,
		})
	}

	return policies
}
