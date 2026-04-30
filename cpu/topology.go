// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package cpu

import (
	"path/filepath"
)

func GetCPUCount() int {
	files, _ := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	return len(files)
}
