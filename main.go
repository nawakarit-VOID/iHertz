// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package main

import (
	"embed"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
)

func setAllCPUMaxFreq(freqKHz uint64) error {
	entries, err := os.ReadDir("/sys/devices/system/cpu")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		// กรองเฉพาะ cpu0, cpu1, cpu2, ...
		if !strings.HasPrefix(name, "cpu") {
			continue
		}
		var idx int
		if _, err := fmt.Sscanf(name, "cpu%d", &idx); err != nil {
			continue
		}

		path := fmt.Sprintf("/sys/devices/system/cpu/%s/cpufreq/scaling_max_freq", name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue // บาง CPU ไม่มี cpufreq
		}

		if err := os.WriteFile(path, []byte(strconv.FormatUint(freqKHz, 10)), 0644); err != nil {
			return fmt.Errorf("cpu%d: %w", idx, err)
		}
	}
	return nil
}

func setGovernor(cpuIndex int, governor string) error {
	path := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor", cpuIndex)
	return os.WriteFile(path, []byte(governor), 0644)
}

// setCPUMaxFreq ตั้งความถี่สูงสุดของ CPU core ที่ระบุ (หน่วย: kHz)
func setCPUMaxFreq(cpuIndex int, freqKHz uint64) error {
	path := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_max_freq", cpuIndex)
	return os.WriteFile(path, []byte(strconv.FormatUint(freqKHz, 10)), 0644)
}

// setCPUMinFreq ตั้งความถี่ต่ำสุด
func setCPUMinFreq(cpuIndex int, freqKHz uint64) error {
	path := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_min_freq", cpuIndex)
	return os.WriteFile(path, []byte(strconv.FormatUint(freqKHz, 10)), 0644)
}

// getCPUFreqInfo อ่านข้อมูลความถี่ของ CPU
func getCPUFreqInfo(cpuIndex int) {
	base := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/", cpuIndex)
	files := map[string]string{
		"scaling_cur_freq": "ความถี่ปัจจุบัน",
		"scaling_max_freq": "ความถี่สูงสุด (เพดาน)",
		"scaling_min_freq": "ความถี่ต่ำสุด",
		"cpuinfo_max_freq": "ความถี่สูงสุดของ hardware",
		"scaling_governor": "governor ที่ใช้อยู่",
	}

	for file, label := range files {
		data, err := os.ReadFile(base + file)
		if err != nil {
			fmt.Printf("  %s: ไม่สามารถอ่านได้\n", label)
			continue
		}
		fmt.Printf("  %s: %s", label, strings.TrimSpace(string(data)))
		if strings.Contains(file, "freq") {
			val, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			fmt.Printf(" kHz (%.2f GHz)", val/1e6)
		}
		fmt.Println()
	}
}

func setCPUMaxFreqWithAuth(freqKHz uint64) error {
	script := fmt.Sprintf(
		"echo %d | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_max_freq",
		freqKHz,
	)

	cmd := exec.Command("pkexec", "bash", "-c", script)
	return cmd.Run()
}

func onButtonClick() {
	freq := uint64(2000000) // อ่านจาก input field

	go func() { // รันใน goroutine ไม่ให้ UI ค้าง
		script := fmt.Sprintf(
			"echo %d | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_max_freq",
			freq,
		)
		cmd := exec.Command("pkexec", "bash", "-c", script)
		err := cmd.Run()
		if err != nil {
			// แสดง error dialog
			fmt.Println("ล้มเหลว")

		}
		// แสดง success dialog
		fmt.Println("สำเร็จ 2GHz")
	}()
}

// โหลด icon
func loadIcon(size int) fyne.Resource {
	var file string

	switch {
	case size >= 512:
		file = "icons/icon-512.png" ///ที่อยู่
	case size >= 256:
		file = "icons/icon-256.png"
	case size >= 128:
		file = "icons/icon-128.png"
	default:
		file = "icons/icon-64.png"
	}

	data, _ := iconFS.ReadFile(file)
	return fyne.NewStaticResource(file, data)
}

//go:embed icons/*
var iconFS embed.FS

//go:embed assets/font/Itim-Regular.ttf
var fontItim []byte
var myFont = fyne.NewStaticResource("Itim-Regular.ttf", fontItim)

var overlayW = color.NRGBA{250, 0, 0, 80}
var overlayB = color.NRGBA{0, 0, 0, 80}

//go:embed assets/lang/English.json
var enJSON []byte

//go:embed assets/lang/THAI.json
var thJSON []byte

// ============================================================================
// MAIN
// ============================================================================

func main() {

	a := app.NewWithID("com.nawakarit.iHertz")
	icon := loadIcon(64)
	w := a.NewWindow("iHertz")
	w.SetIcon(icon)

	ProgressCpu0 := widget.NewProgressBar()
	fmt.Println("=== ข้อมูล CPU0 ===")

	getCPUFreqInfo(0)

	// ตัวอย่าง: ตั้งเพดานที่ 2.0 GHz = 2,000,000 kHz
	targetFreq := uint64(2_000_000)
	fmt.Printf("\nตั้งเพดานความถี่ CPU0 เป็น %.1f GHz...\n", float64(targetFreq)/1e6)

	if err := setCPUMaxFreq(0, targetFreq); err != nil {
		fmt.Printf("เกิดข้อผิดพลาด: %v (ต้องรันด้วย root)\n", err)
		return
	}

	// governor ที่ใช้บ่อย: "powersave", "performance", "schedutil", "ondemand"
	setGovernor(0, "powersave")
	fmt.Println("สำเร็จ!")

	bt1 := widget.NewButton("TTT", func() {
		onButtonClick()
	})

	w.SetContent(container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		bt1),
	)

	w.Resize(fyne.NewSize(200, 200))
	w.ShowAndRun()
}
