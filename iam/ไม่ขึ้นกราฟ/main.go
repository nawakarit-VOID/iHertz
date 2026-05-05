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
	"math"
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

//////////////////////////////////////////////////
// 🔥 MultiGraph (single buffer, multi-line)
//////////////////////////////////////////////////

type MultiGraph struct {
	img    *image.RGBA
	w, h   int
	maxVal float64

	colors []color.RGBA
	prevY  []float64
	smooth []float64
}

func NewMultiGraph(w, h, cores int) *MultiGraph {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// สีแต่ละ core (วนใช้ได้)
	baseColors := []color.RGBA{
		{0, 255, 0, 255},
		{0, 128, 255, 255},
		{255, 0, 0, 255},
		{255, 255, 0, 255},
		{255, 0, 255, 255},
		{0, 255, 255, 255},
		{255, 128, 0, 255},
		{128, 255, 0, 255},
	}

	colors := make([]color.RGBA, cores)
	for i := 0; i < cores; i++ {
		colors[i] = baseColors[i%len(baseColors)]
	}

	g := &MultiGraph{
		img:    img,
		w:      w,
		h:      h,
		maxVal: 100,
		colors: colors,
		prevY:  make([]float64, cores),
		smooth: make([]float64, cores),
	}
	// init
	for i := range g.prevY {
		g.prevY[i] = float64(h)
		g.smooth[i] = 0
	}

	g.clear()
	return g
}

func (g *MultiGraph) clear() {
	for i := 0; i < len(g.img.Pix); i += 4 {
		g.img.Pix[i+0] = 0
		g.img.Pix[i+1] = 0
		g.img.Pix[i+2] = 0
		g.img.Pix[i+3] = 255
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

		// clear ขวาสุด
		idx := row + (g.w-1)*4
		g.img.Pix[idx+0] = 0
		g.img.Pix[idx+1] = 0
		g.img.Pix[idx+2] = 0
		g.img.Pix[idx+3] = 255
	}
}

// 🔥 วาดค่าของแต่ละ core เป็นเส้นแนวตั้ง
// 🎯 draw wave + glow + smooth
func (g *MultiGraph) draw(values []float64) {
	x := g.w - 1

	for i, v := range values {
		if i >= len(g.colors) {
			break
		}

		// 🔥 smoothing (EMA)
		g.smooth[i] = g.smooth[i]*0.7 + v*0.3

		// map → Y
		y := float64(g.h) - (g.smooth[i]/g.maxVal)*float64(g.h)

		col := g.colors[i]
		prev := g.prevY[i]

		// 🌈 glow (วาดหลายชั้น)
		drawLine(g.img, x-1, int(prev), x, int(y), fade(col, 60))
		drawLine(g.img, x-1, int(prev+1), x, int(y+1), fade(col, 40))
		drawLine(g.img, x-1, int(prev-1), x, int(y-1), fade(col, 40))

		// 🎯 เส้นหลัก (คม)
		drawLine(g.img, x-1, int(prev), x, int(y), col)

		g.prevY[i] = y
	}
}

func (g *MultiGraph) Update(values []float64) {
	g.shiftLeft()
	g.draw(values)
}

//////////////////////////////////////////////////
// ✏️ draw line (Bresenham)
//////////////////////////////////////////////////

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := int(math.Abs(float64(x2 - x1)))
	dy := -int(math.Abs(float64(y2 - y1)))
	sx := 1
	if x1 >= x2 {
		sx = -1
	}
	sy := 1
	if y1 >= y2 {
		sy = -1
	}
	err := dx + dy

	for {
		if x1 >= 0 && x1 < img.Rect.Max.X && y1 >= 0 && y1 < img.Rect.Max.Y {
			idx := (y1*img.Rect.Max.X + x1) * 4
			img.Pix[idx+0] = c.R
			img.Pix[idx+1] = c.G
			img.Pix[idx+2] = c.B
			img.Pix[idx+3] = c.A
		}

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
}

func fade(c color.RGBA, a uint8) color.RGBA {
	return color.RGBA{c.R, c.G, c.B, a}
}

//////////////////////////////////////////////////
// 🔥 CPU per core
//////////////////////////////////////////////////

func getCPU() []float64 {
	v, err := cpu.Percent(0, true)
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

	// สร้าง graph
	graph := NewMultiGraph(800, 300, 8)

	// Raster ใช้ buffer เดิมตลอด
	raster := canvas.NewRaster(func(w, h int) image.Image {
		return graph.img
	})

	// 🔁 loop real-time
	go func() {
		for {
			values := getCPU()
			if len(values) > 0 {
				graph.Update(values)
			}

			// thread-safe (ต้อง Fyne v2.4+)
			fyne.Do(func() {
				raster.Refresh()
			})

			time.Sleep(50 * time.Millisecond) // ~20 FPS
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
		container.NewCenter(raster)),
	)

	w.Resize(fyne.NewSize(900, 300))
	w.ShowAndRun()
}
