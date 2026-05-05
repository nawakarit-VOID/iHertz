// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	//"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v3/cpu"
)

//-------------------------------------------------------------------------------------------

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

var iconFS embed.FS

var overlayW = color.NRGBA{250, 0, 0, 80}
var overlayB = color.NRGBA{0, 0, 0, 80}

//////////////////////////////////////////////////
// 🔥 MultiGraph (single buffer, multi-line)
//////////////////////////////////////////////////

type MultiGraph struct {
	img    *image.RGBA
	w, h   int
	maxVal float64
	colors []color.RGBA
}

func NewMultiGraph(w, h, cores int) *MultiGraph {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// สีแต่ละ core (วนใช้ได้)
	colors := []color.RGBA{
		{0, 255, 0, 255},
		{0, 128, 255, 255},
		{255, 0, 0, 255},
		{255, 255, 0, 255},
		{255, 0, 255, 255},
		{0, 255, 255, 255},
		{255, 128, 0, 255},
		{128, 255, 0, 255},
	}

	if cores > len(colors) {
		// ถ้า core เยอะ เอาสีวน
		tmp := make([]color.RGBA, cores)
		for i := 0; i < cores; i++ {
			tmp[i] = colors[i%len(colors)]
		}
		colors = tmp
	} else {
		colors = colors[:cores]
	}

	g := &MultiGraph{
		img:    img,
		w:      w,
		h:      h,
		maxVal: 100.0,
		colors: colors,
	}

	g.clear()
	return g
}

func (g *MultiGraph) clear() {
	for y := 0; y < g.h; y++ {
		row := y * g.img.Stride
		for x := 0; x < g.w; x++ {
			idx := row + x*4
			g.img.Pix[idx+0] = 0
			g.img.Pix[idx+1] = 0
			g.img.Pix[idx+2] = 0
			g.img.Pix[idx+3] = 255
		}
	}
}

// 🔥 shift pixel ไปซ้าย (เร็ว)
func (g *MultiGraph) shiftLeft() {
	for y := 0; y < g.h; y++ {
		row := y * g.img.Stride

		copy(
			g.img.Pix[row:row+(g.w-1)*4],
			g.img.Pix[row+4:row+g.w*4],
		)

		// เคลียร์คอลัมน์ขวาสุด
		idx := row + (g.w-1)*4
		g.img.Pix[idx+0] = 0
		g.img.Pix[idx+1] = 0
		g.img.Pix[idx+2] = 0
		g.img.Pix[idx+3] = 255
	}
}

// 🔥 วาดค่าของแต่ละ core เป็นเส้นแนวตั้ง
func (g *MultiGraph) draw(values []float64) {
	x := g.w - 1

	for i, v := range values {
		if i >= len(g.colors) {
			break
		}

		hVal := int((v / g.maxVal) * float64(g.h))
		col := g.colors[i]

		for y := g.h - 1; y >= g.h-hVal && y >= 0; y-- {
			idx := (y*g.w + x) * 4
			g.img.Pix[idx+0] = col.R
			g.img.Pix[idx+1] = col.G
			g.img.Pix[idx+2] = col.B
			g.img.Pix[idx+3] = 255
		}
	}
}

func (g *MultiGraph) Update(values []float64) {
	g.shiftLeft()
	g.draw(values)
}

//////////////////////////////////////////////////
// 🔥 CPU per core
//////////////////////////////////////////////////

func getCPUPerCore() []float64 {
	v, err := cpu.Percent(0, true) // true = per core
	if err != nil {
		log.Println(err)
		return nil
	}
	return v
}

// ============================================================================
// MAIN
// ============================================================================

func main() {

	a := app.NewWithID("com.nawakarit.iHertz")
	icon := loadIcon(64)
	w := a.NewWindow("iHertz")
	w.SetIcon(icon)
	//w.Resize(fyne.NewSize(300, 100))

	//w.Resize(fyne.NewSize(200, 200))
	//data := binding.NewFloat()
	//bar := widget.NewProgressBar()
	//label := widget.NewLabel("0%")
	//bar := widget.NewProgressBarWithData(data)
	//label := widget.NewLabelWithData(binding.FloatToString(data))
	//w.SetContent(container.NewVBox(bar, label))
	/*
		go func() {
			for {
				val := getCPU() // 0.0 - 1.0
				data.Set(val)   // ✅ thread-safe
				time.Sleep(500 * time.Millisecond)
			}
		}()
	*/
	//ProgressCpu0 := widget.NewProgressBar()
	//fmt.Println("=== ข้อมูล CPU0 ===")

	//globalProgress.SetValue(float64(fi) / float64(totalFolders))
	//getCPUFreqInfo(0)
	/*
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
	*/
	// สร้าง graph
	graph := NewMultiGraph(800, 300, 8)

	// Raster ใช้ buffer เดิมตลอด
	raster := canvas.NewRaster(func(w, h int) image.Image {
		return graph.img
	})

	w.Resize(fyne.NewSize(800, 300))

	// 🔁 loop real-time
	go func() {
		for {
			values := getCPUPerCore()
			if len(values) > 0 {
				graph.Update(values)
			}

			// thread-safe (ต้อง Fyne v2.4+)
			fyne.Do(func() {
				raster.Refresh()
			})

			time.Sleep(100 * time.Millisecond)
		}
	}()
	/*
		bt1 := widget.NewButton("TTT", func() {
			onButtonClick()
		})
	*/
	w.SetContent(container.NewBorder(
		//container.NewVBox(bar, label),

		nil,
		nil,
		nil,
		nil,
		container.NewMax(raster)),
	)

	w.ShowAndRun()
}
