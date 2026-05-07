// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package main

import (
	"embed"
	"image/color"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

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

//go:embed assets/lang/English.json
var enJSON []byte

//go:embed assets/lang/THAI.json
var thJSON []byte

//////////////////////////////////////////////////
// 🚀 main
//////////////////////////////////////////////////

func main() {
	a := app.NewWithID("com.nawakarit.iHert")
	icon := loadIcon(64)
	w := a.NewWindow("iHert")
	w.SetIcon(icon)
	//a.Settings().SetTheme(&MyTheme{})
	coreCount := runtime.NumCPU()

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

	cards := make([]*CoreCard, coreCount)
	items := make([]fyne.CanvasObject, coreCount)

	for i := 0; i < coreCount; i++ {
		c := NewCoreCard(i, colors[i%len(colors)])
		cards[i] = c
		items[i] = c.root
	}

	grid := container.NewGridWithColumns(2, items...)

	go func() {
		for {
			values := getCPU()

			for i, c := range cards {
				if i < len(values) {
					v := values[i]
					c.graph.Update(v)
					c.val.Set(v)
				}
			}

			fyne.Do(func() {
				for _, c := range cards {
					c.raster.Refresh()
				}
			})

			time.Sleep(80 * time.Millisecond)
		}

	}()

	//container.NewBorder(nil,nil,nil,nil, getCPUFreqInfo(0))
	//getCPUFreqInfo(0)
	w.SetContent(container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		grid,
	))

	w.Resize(fyne.NewSize(650, 500))
	w.ShowAndRun()
}
