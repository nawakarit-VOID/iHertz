// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package cpu

type Info struct {
	Cores   int
	Freqs   []float64
	HasFreq bool
}

func GetInfo() Info {
	freqs := GetCurrentFreqs()

	return Info{
		Cores:   GetCPUCount(),
		Freqs:   freqs,
		HasFreq: len(freqs) > 0,
	}
}
