// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package main

import (
	"fmt"
	"time"

	"iHertz/cpu"
)

func main() {
	for {
		info := cpu.GetInfo()

		fmt.Print("\033[H\033[2J")

		fmt.Println("Cores:", info.Cores)

		for i, f := range info.Freqs {
			fmt.Printf("CPU %d: %.2f GHz\n", i, f)
		}

		time.Sleep(300 * time.Millisecond)
	}
}
