// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.

package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/shirou/gopsutil/v3/cpu"
)

//////////////////////////////////////////////////
// 🔥 MultiGraph Wave
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

// 🔥 shift pixel (เร็วมาก)
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
// 📊 CPU
//////////////////////////////////////////////////

func getCPU() []float64 {
	v, err := cpu.Percent(0, true)
	if err != nil {
		log.Println(err)
		return nil
	}
	return v
}

//////////////////////////////////////////////////
// 🚀 main
//////////////////////////////////////////////////

func main() {
	a := app.New()
	w := a.NewWindow("Wave Multi-Core Monitor")

	graph := NewMultiGraph(900, 300, 8)

	raster := canvas.NewRaster(func(w, h int) image.Image {
		return graph.img
	})

	w.Resize(fyne.NewSize(900, 300))

	go func() {
		for {
			values := getCPU()
			if len(values) > 0 {
				graph.Update(values)
			}

			// ต้องใช้ Fyne v2.4+
			fyne.Do(func() {
				raster.Refresh()
			})

			time.Sleep(100 * time.Millisecond) // ~20 FPS
		}
	}()
	w.SetContent(container.NewBorder(nil, nil, nil, nil, raster))
	w.ShowAndRun()
}
