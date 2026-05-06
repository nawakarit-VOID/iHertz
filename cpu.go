// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v3/cpu"
)

//////////////////////////////////////////////////
// 🔥 Graph
//////////////////////////////////////////////////

type Graph struct {
	img    *image.RGBA
	w, h   int
	maxVal float64

	color  color.RGBA
	prevY  float64
	smooth float64
}

func NewGraph(w, h int, col color.RGBA) *Graph {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	g := &Graph{
		img:    img,
		w:      w,
		h:      h,
		maxVal: 100,
		color:  col,
		prevY:  float64(h),
	}
	g.clear()
	return g
}

func (g *Graph) clear() {
	for i := 0; i < len(g.img.Pix); i += 4 {
		g.img.Pix[i+0] = 0
		g.img.Pix[i+1] = 0
		g.img.Pix[i+2] = 0
		g.img.Pix[i+3] = 255
	}
}

func (g *Graph) shiftLeft() {
	for y := 0; y < g.h; y++ {
		row := y * g.img.Stride
		copy(
			g.img.Pix[row:row+(g.w-1)*4],
			g.img.Pix[row+4:row+g.w*4],
		)
		idx := row + (g.w-1)*4
		g.img.Pix[idx+0] = 0
		g.img.Pix[idx+1] = 0
		g.img.Pix[idx+2] = 0
		g.img.Pix[idx+3] = 255
	}
}

func (g *Graph) draw(v float64) {
	g.smooth = g.smooth*0.7 + v*0.3

	y := float64(g.h) - (g.smooth/g.maxVal)*float64(g.h)
	x := g.w - 1
	prev := g.prevY

	// glow
	drawLine(g.img, x-1, int(prev), x, int(y), fade(g.color, 60))
	drawLine(g.img, x-1, int(prev+1), x, int(y+1), fade(g.color, 40))
	drawLine(g.img, x-1, int(prev-1), x, int(y-1), fade(g.color, 40))

	// main
	drawLine(g.img, x-1, int(prev), x, int(y), g.color)

	g.prevY = y
}

func (g *Graph) Update(v float64) {
	g.shiftLeft()
	g.draw(v)
}

//////////////////////////////////////////////////
// ✏️ draw line
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
// 🧩 Card
//////////////////////////////////////////////////

type CoreCard struct {
	root   fyne.CanvasObject
	graph  *Graph
	raster *canvas.Raster
	val    binding.Float
}

func NewCoreCard(idx int, col color.RGBA) *CoreCard {
	g := NewGraph(300, 120, col)

	r := canvas.NewRaster(func(w, h int) image.Image {
		return g.img
	})

	val := binding.NewFloat()

	title := widget.NewLabel(fmt.Sprintf("Core %d", idx))
	percent := widget.NewLabelWithData(binding.FloatToStringWithFormat(val, "%.0f%%"))

	top := container.NewBorder(nil, nil, title, percent)

	card := widget.NewCard("", "", container.NewBorder(top, nil, nil, nil, r))

	return &CoreCard{
		root:   card,
		graph:  g,
		raster: r,
		val:    val,
	}
}

// ============================================================================
// cpuinfo
// ============================================================================
