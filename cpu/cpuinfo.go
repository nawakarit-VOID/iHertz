// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package cpu

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v3/cpu"
)

// จำนวนคอร์
func CpuCoreCount() int {
	physicalCore, err := cpu.Counts(false) //core
	if err != nil {
		log.Println(err)
		return (0)
	}
	return physicalCore //core จริง
}

// จำนวนเทรด
func CpuThreadCount() int {
	logical, err := cpu.Counts(true) //thread
	if err != nil {
		log.Println(err)
		return (0)
	}
	return logical
}

func CpuPercentAVG() []float64 { //*[]float64
	// ดึง CPU usage รวม
	percentAVG, err := cpu.Percent(100*time.Millisecond, false)
	if err != nil || len(percentAVG) == 0 {
		log.Println(err)
		return []float64{0.0}
	}
	return percentAVG
}

func CpuPercentPercore() []float64 {
	// ดึง CPU usage ต่อ core
	percentPerCore, err := cpu.Percent(100*time.Millisecond, true)
	if err != nil || len(percentPerCore) == 0 {
		log.Println(err)
		return []float64{0.0}
		//return nil

	}
	return percentPerCore
}

// ============================================================================
// monitor
// ============================================================================
type StCPUData struct {
	Usage string //
	//Timesusage string
	UsagepercentTotal         string
	UsagepercentPerCoreSTRING string
	TimesTotalAvg             string
	TimesSec                  string
	TimesHms                  string
	UsagePerCore              []float64 // CPU usage ต่อ core
	PercentPerCore            string
	Times                     []cpu.TimesStat
	//////////////////////

}
type CPUMonitor struct {
	ticker   *time.Ticker
	callback func(StCPUData)
}

// สร้าง instance ใหม่
func NewCPUMonitor(interval time.Duration, callback func(StCPUData)) *CPUMonitor {
	return &CPUMonitor{
		ticker:   time.NewTicker(interval),
		callback: callback,
	}
}

// เริ่ม monitoring
func (m *CPUMonitor) Start() {
	go func() {
		for range m.ticker.C {

			percentTotal := CpuPercentAVG()
			percentPerCore := CpuPercentPercore()
			//จัดเรียง usage

			usagepercentTotal := fmt.Sprintf("%.2f %%\n", percentTotal[0]) //percentTotal[0]			// แสดง usage ต่อ core
			var usagepercentPerCore string
			//usagepercentPerCore += "[ Usage PerCore ]\n"
			for i, pc := range percentPerCore {
				usagepercentPerCore += fmt.Sprintf("Core [ %d ] : %.2f %%\n", i, pc)
			}

			var timesTotalAvg string
			var timesSec string
			//timesSec += "[ ข้อมูลดิบ ]"
			var timesHms string
			//timesHms += "[ แปลงเป็นเวลาสากล ]"

			if len(percentTotal) > 0 {

				data := StCPUData{
					//Usage: usage,
					//Timesusage: timesusage,
					UsagepercentTotal:         usagepercentTotal,
					UsagepercentPerCoreSTRING: usagepercentPerCore,
					TimesTotalAvg:             timesTotalAvg,
					TimesSec:                  timesSec,
					TimesHms:                  timesHms,
				}

				m.callback(data)

			}

		}

	}()
}

// - // กราฟ - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
func grid() fyne.CanvasObject {
	coreCount := CpuCoreCount()

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

	grid := container.NewGridWithColumns(1, items...)

	go func() {
		for {
			values := CpuPercentPercore()

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
			time.Sleep(100 * time.Millisecond)
		}

	}()
	return grid
}

// ============================================================================
// CpuTabs
// ============================================================================
func CpuTabs(w fyne.Window) fyne.CanvasObject {

	//cpuUsagePage//
	usagepercentTotalLabel := widget.NewLabel("usagepercentTotalLabel...")
	usagepercentTotalLabel.Alignment = fyne.TextAlignCenter
	usagePerCoreSTRINGLabel := widget.NewLabel("usagePerCoreSTRINGLabel...")
	usagePerCoreSTRINGLabel.Alignment = fyne.TextAlignCenter

	//cpuTimesusagePage//
	timesTotalAvg := widget.NewLabel("timesTotalAvg...")
	timesSec := widget.NewLabel("timesSec...")
	timesHms := widget.NewLabel("timesHms...")

	// สร้าง monitor cpu
	monitor := NewCPUMonitor(1*time.Second, func(data StCPUData) {
		fyne.Do(func() {
			usagepercentTotalLabel.SetText(fmt.Sprintf("%s", data.UsagepercentTotal))          //4 // แสดง usage รวม
			usagePerCoreSTRINGLabel.SetText(fmt.Sprintf("%s", data.UsagepercentPerCoreSTRING)) //4 // แสดง usage รวม
			timesTotalAvg.SetText(fmt.Sprintf("%s", data.TimesTotalAvg))
			timesSec.SetText(fmt.Sprintf("%s", data.TimesSec))
			timesHms.SetText(fmt.Sprintf("%s", data.TimesHms))
		})
	})

	monitor.Start() // เริ่ม monitoring

	grid := grid()
	//layout
	Grid := container.NewBorder(nil, nil, nil, nil, grid)

	cpuUsagePage := container.NewVBox(
		container.NewBorder(
			widget.NewCard("กราฟ", "", Grid),
			//Grid,
			nil,
			nil,
			nil,
		),
		container.NewVBox(
			widget.NewCard("AVG", "", usagepercentTotalLabel),
			//usagepercentTotalLabel,
			widget.NewCard("PerCore", "", usagePerCoreSTRINGLabel),
			//usagePerCoreSTRINGLabel,
		),
	)

	cpuControlPage := CpuControl(w)

	return container.NewAppTabs(
		container.NewTabItem("Control", container.NewScroll(cpuControlPage)),
		container.NewTabItem("Usage", container.NewScroll(cpuUsagePage)),
	)
}
