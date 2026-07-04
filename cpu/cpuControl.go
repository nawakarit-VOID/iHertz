// Copyright (c) 2026 Nawakarit
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License v3.0.
package cpu

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// -------------------------------------------------------------------------------------------

func getCPUhardware(cpuIndex int) (fyne.CanvasObject, uint64, uint64) {
	base := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/", cpuIndex)
	files := []struct {
		file  string
		label string
	}{
		{"cpuinfo_min_freq", "ความถี่ต่ำสุด"},
		{"cpuinfo_max_freq", "ความถี่สูงสุด"},
	}
	x := widget.NewLabel("กำลังโหลด...")
	x.Alignment = fyne.TextAlignCenter //ทำให้ตรงกลาง

	var x1 strings.Builder
	var val_cpuinfo_min_freq uint64
	var val_cpuinfo_max_freq uint64

	//x1.WriteString("Default Kernel and Hardware")
	//x1.WriteString("\n|")

	for _, item := range files {
		data, err := os.ReadFile(base + item.file)
		if err != nil {
			x1.WriteString(fmt.Sprintf("\n%s: ไม่สามารถอ่านได้", item.label))
			continue
		}
		//string ไม่จำเป็นต้อง update()
		value := strings.TrimSpace(string(data))
		x1.WriteString(fmt.Sprintf("\n%s : %s", item.label, value))

		if strings.Contains(item.file, "freq") {
			val, _ := strconv.ParseFloat(value, 64)
			x1.WriteString(fmt.Sprintf(" kHz // (%.2f GHz)", val/1e6))
		}

		//เอาค่าออกมา
		if strings.Contains(item.file, "cpuinfo_min_freq") {
			val, _ := strconv.ParseFloat(value, 64)
			val_cpuinfo_min_freq = uint64(val)
		}
		if strings.Contains(item.file, "cpuinfo_max_freq") {
			val, _ := strconv.ParseFloat(value, 64)
			val_cpuinfo_max_freq = uint64(val)
		}

	}
	fyne.Do(func() {
		x.SetText(x1.String())
	})

	return x, val_cpuinfo_min_freq, val_cpuinfo_max_freq
}

// getCPUFreqInfo อ่านข้อมูลความถี่ของ CPU
func getCPUFreqUpdate(cpuIndex int) (fyne.CanvasObject, uint64, uint64) {

	base := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/", cpuIndex)
	files := []struct {
		file  string
		label string
	}{
		{"scaling_cur_freq", "ความถี่ปัจจุบัน"},
		{"scaling_min_freq", "ความถี่ต่ำสุด (จำกัด)"},
		{"scaling_max_freq", "ความถี่สูงสุด (จำกัด)"},
		{"cpuinfo_transition_latency", "เวลาในการเปลี่ยนความเร็ว"},
		{"scaling_governor", "Governor ที่ใช้อยู่"},
	}

	var val_scaling_min_freq uint64
	var val_scaling_mmax_freq uint64

	x := widget.NewLabel("กำลังโหลด...")

	update := func() {
		var x1 strings.Builder
		//x1.WriteString("ยังไม่รองรับหลาย cpu")
		x1.WriteString(fmt.Sprintf("core [ %d ]", cpuIndex))

		for _, item := range files {
			data, err := os.ReadFile(base + item.file)
			if err != nil {
				x1.WriteString(fmt.Sprintf("\n%s: ไม่สามารถอ่านได้", item.label))
				continue
			}

			value := strings.TrimSpace(string(data))
			x1.WriteString(fmt.Sprintf("\n%s : %s", item.label, value))

			if strings.Contains(item.file, "freq") {
				val, _ := strconv.ParseFloat(value, 64)
				x1.WriteString(fmt.Sprintf(" kHz // (%.2f GHz)", val/1e6))
			}
			if strings.Contains(item.file, "latency") {
				val, _ := strconv.ParseFloat(value, 64)
				x1.WriteString(fmt.Sprintf(" nS // (%.f uS)", val/1e3))
			}
			//เอาค่าปัจจุบันออกมา
			if strings.Contains(item.file, "scaling_min_freq") {
				val, _ := strconv.ParseFloat(value, 64)
				val_scaling_min_freq = uint64(val)
			}
			if strings.Contains(item.file, "scaling_max_freq") {
				val, _ := strconv.ParseFloat(value, 64)
				val_scaling_mmax_freq = uint64(val)
			}

		}

		fyne.Do(func() {
			x.SetText(x1.String())
		})
	}

	update()
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			update()
		}
	}()

	return x, val_scaling_min_freq, val_scaling_mmax_freq
}

// ============================================================================
// เพิ่ม label ตามจำนวนคอร์
// ============================================================================
func sysCPUFreqUpdate() fyne.CanvasObject {
	coreCount := CpuCoreCount()
	//box := container.NewVBox()
	box := container.NewGridWithColumns(2)

	for i := 0; i < coreCount; i++ {
		coreInfo, _, _ := getCPUFreqUpdate(i)
		box.Add(coreInfo)
	}
	if coreCount == 0 {
		return widget.NewLabel("ไม่พบข้อมูลจำนวนคอร์ CPU")
	}
	return box
}

// ============================================================================
// เลือกทั้งหมด
// ============================================================================
func CheckAllBoxCpu(checkboxes []*widget.Check, selected []bool, updateLabel func()) {
	for idx, check := range checkboxes {
		check.SetChecked(true)
		selected[idx] = true
	}
	updateLabel()
	//fmt.Println("เลือกทั้งหมด")
}

// ============================================================================
// ไม่เลือกทั้งหมด
// ============================================================================
func nonCheckBoxCpu(checkboxes []*widget.Check, selected []bool, updateLabel func()) {
	for idx, check := range checkboxes {
		check.SetChecked(false)
		selected[idx] = false
	}
	updateLabel()
	//fmt.Println("ล้างทั้งหมด")
}

// ============================================================================
// เพิ่ม checkbox ตามจำนวนคอร์
// ============================================================================
func checkboxNumcpu() (fyne.CanvasObject, []bool, []*widget.Check, func()) {
	coreCount := CpuCoreCount()
	if coreCount == 0 {
		return widget.NewLabel("ไม่พบข้อมูลจำนวนคอร์ CPU"), nil, nil, nil
	}

	selected := make([]bool, coreCount)
	checkboxes := make([]*widget.Check, coreCount)
	for i := 0; i < coreCount; i++ {
		selected[i] = true
	}

	selectedGet, _ := getSelectedCoresText(selected)
	//fmt.Println("core เริ่ม", selectedGet)
	selectedLabel := widget.NewLabel(selectedGet)

	box := container.NewGridWithColumns(10) //8

	for i := 0; i < coreCount; i++ {
		idx := i
		coreName := strconv.Itoa(idx)
		x := widget.NewCheck("[ "+coreName+" ]", func(checked bool) {
			selected[idx] = checked
			if checked {
				//fmt.Println("core", idx, "เปิด")
			} else {
				//fmt.Println("core", idx, "ปิด")
			}
			selectedGet, _ := getSelectedCoresText(selected)
			//fmt.Println("core ใน for", selectedGet)

			selectedLabel.SetText(selectedGet)
		})

		x.SetChecked(true)
		checkboxes[idx] = x
		box.Add(x)
	}

	// สร้าง function เพื่ออัปเดต label
	updateLabel := func() {
		selectedGet, _ := getSelectedCoresText(selected)
		selectedLabel.SetText(selectedGet)
	}

	return container.NewVBox(selectedLabel, widget.NewSeparator(), box), selected, checkboxes, updateLabel
}

func getSelectedCoresText(selected []bool) (string, []int) {
	var cores []string
	var coresIndices []int
	for idx, checked := range selected {
		if checked {
			cores = append(cores, strconv.Itoa(idx))
			coresIndices = append(coresIndices, idx)
		}
	}

	if len(cores) == 0 {
		return "ยังไม่ได้เลือก", nil
	}

	var lines []string

	for i := 0; i < len(cores); i += 48 { //48
		end := i + 48 //48
		if end > len(cores) {
			end = len(cores)
		}

		lines = append(lines,
			strings.Join(cores[i:end], ", "),
		)
	}
	//return "คอร์ที่เลือก : " + strings.Join(lines, "\n"), coresIndices
	//return "คอร์ที่เลือก : " + lines[0] + "\n                     " + strings.Join(lines[1:], "\n                     "), coresIndices
	return "" + lines[0] + "\n" + strings.Join(lines[1:], "\n"), coresIndices

} //21

// ============================================================================
// เรียกไฟล์ govenors
// ============================================================================
func GetGovernors() ([]string, error) {
	data, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors")
	if err != nil {
		return nil, err
	}

	return strings.Fields(string(data)), nil
}

// สร้าง check Govern...
func GovernorscheckBox() (fyne.CanvasObject, *widget.RadioGroup) {

	governors, _ := GetGovernors()
	governorsST := widget.NewRadioGroup(governors, nil)
	governorsST.Horizontal = true

	if len(governors) > 0 {
		governorsST.SetSelected(governors[0])
	}

	return governorsST, governorsST
}

// ฟังชั้นปุ่มกด เพื่ม ลด ต่อเนื่อง---
type RepeatButton struct {
	widget.Button
	ticker *time.Ticker
	stop   chan struct{}
	action func()
}

func NewRepeatButton(text string, action func()) *RepeatButton {
	b := &RepeatButton{
		action: action,
	}
	b.ExtendBaseWidget(b)
	b.SetText(text)
	return b
}

func (b *RepeatButton) MouseDown(*desktop.MouseEvent) {
	if b.ticker != nil {
		return
	}

	// ทำครั้งแรกทันที
	b.action()

	b.stop = make(chan struct{})
	b.ticker = time.NewTicker(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-b.ticker.C:
				fyne.Do(func() {
					b.action()
				})
			case <-b.stop:
				return
			}
		}
	}()
}

func (b *RepeatButton) MouseUp(*desktop.MouseEvent) {
	if b.ticker == nil {
		return
	}

	b.ticker.Stop()
	close(b.stop)

	b.ticker = nil
	b.stop = nil
} // ---ฟังชั้นปุ่มกด เพื่ม ลด ต่อเนื่อง

// ปุ่มกด เพื่ม ลด---
func onButtonMinN(min_freq_Slider *widget.Slider) { //ลดค่า min
	freq_min := min_freq_Slider.Value - 1
	if freq_min >= min_freq_Slider.Min {
		min_freq_Slider.SetValue(freq_min)
	}
}

func onButtonMinP(min_freq_Slider *widget.Slider) { //เพิ่มค่า min
	freq_min := min_freq_Slider.Value + 1
	if freq_min <= min_freq_Slider.Max {
		min_freq_Slider.SetValue(freq_min)
	}
}

func onButtonMaxN(max_freq_Slider *widget.Slider) { //ลดค่า max
	freq_max := max_freq_Slider.Value - 1
	if freq_max >= max_freq_Slider.Min {
		max_freq_Slider.SetValue(freq_max)
	}
}

func onButtonMaxP(max_freq_Slider *widget.Slider) { //เพิ่มค่า max
	freq_max := max_freq_Slider.Value + 1
	if freq_max <= max_freq_Slider.Max {
		max_freq_Slider.SetValue(freq_max)
	}
} //---ปุ่มกด เพื่ม ลด

// หาส่วนต่าง + แปลงเป็น 1%
func freq_percent(percent_freq_Slider *widget.Slider) float64 {
	freq_percent := (percent_freq_Slider.Max - percent_freq_Slider.Min) / 100
	return freq_percent
}

func x101(percent float64, freq_Slider *widget.Slider) {
	freqNow := freq_percent(freq_Slider)
	freqNow = freqNow * percent
	freqNow = freqNow + freq_Slider.Min

	//เช็ค น้อยกว่า ให้ทำงาน
	if freqNow <= freq_Slider.Max {
		freq_Slider.SetValue(freqNow)
	}
}

// test new percent
func newPercent(freq_Slider *widget.Slider, percent float64) {

	switch percent {
	case 10:
		x101(percent, freq_Slider)
		fmt.Println("10%")
	case 20:
		x101(percent, freq_Slider)
		fmt.Println("20%")
	case 30:
		x101(percent, freq_Slider)
		fmt.Println("30%")
	case 40:
		x101(percent, freq_Slider)
		fmt.Println("40%")
	case 50:
		x101(percent, freq_Slider)
		fmt.Println("50%")
	case 60:
		x101(percent, freq_Slider)
		fmt.Println("60%")
	case 70:
		x101(percent, freq_Slider)
		fmt.Println("70%")
	case 80:
		x101(percent, freq_Slider)
		fmt.Println("80%")
	case 90:
		x101(percent, freq_Slider)
		fmt.Println("90%")
	case 100:
		x101(percent, freq_Slider)
		fmt.Println("100%")

	}
}

// ************************************//

func slider() (*widget.Slider, *widget.Slider, *widget.Label, *widget.Label, *widget.Entry, *widget.Entry) {

	_, val_min, val_max := getCPUhardware(0)
	_, cur_min, cur_max := getCPUFreqUpdate(0)

	entry_min := widget.NewEntry()
	entry_min.SetText(strconv.FormatUint(cur_min, 10)) //10 คือแปลงเป็นเลขฐาน 10
	//entry_min.SetText(fmt.Sprintf("%d", cur_min/1000)) //หรือ

	entry_max := widget.NewEntry()
	entry_max.SetText(strconv.FormatUint(cur_max, 10)) //10 คือแปลงเป็นเลขฐาน 10
	//entry_max.SetText(fmt.Sprintf("%d", cur_max/1000)) //หรือ

	val_ch_min := val_min
	val_ch_max := val_max

	//label slider min และ max
	min_freq_Label := widget.NewLabel(fmt.Sprintf("[ จำกัด - ความถี่ต่ำสุด ] %d kHz [ %.2f Ghz ]", val_ch_min, float64(val_ch_min)/1e6))
	max_freq_Label := widget.NewLabel(fmt.Sprintf("[ จำกัด - ความถี่สูงสุด ] %d kHz [ %.2f Ghz ]", val_ch_max, float64(val_ch_max)/1e6))

	//ค่า slider ต่ำสุด-มากสุด ของ min และ max
	min_freq_Slider := widget.NewSlider(float64(val_min), float64(val_max)) //*min
	max_freq_Slider := widget.NewSlider(float64(val_min), float64(val_max)) //*max

	//*min
	entry_min.OnChanged = func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err == nil {
			min_freq_Slider.SetValue(v)
		}
	}
	//slider_min
	min_freq_Slider.Step = 1
	min_freq_Slider.Value = float64(cur_min) //ตั้งค่าเริ่มต้นของ slider
	min_freq_Slider.OnChanged = func(v float64) {
		//ตรวจสอบ min มากกว่า max ให้ เปลี่ยนค่า max ตาม min
		if v > max_freq_Slider.Value {
			max_freq_Slider.SetValue(v)
			//fmt.Println("max <= min")
		}

		entry_min.SetText(fmt.Sprintf("%.f", v))
		val_ch_min = uint64(v) //แปลงเป็น uint64
		min_freq_Label.SetText(fmt.Sprintf("[ จำกัด - ความถี่ต่ำสุด ] %d kHz [ %.2f Ghz ]", val_ch_min, float64(val_ch_min)/1e6))
	}

	//*max
	entry_max.OnChanged = func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err == nil {
			max_freq_Slider.SetValue(v)
		}
	}
	//slider_max
	max_freq_Slider.Step = 1
	max_freq_Slider.Value = float64(cur_max)
	max_freq_Slider.OnChanged = func(v float64) {
		//ตรวจสอบ max มากกว่า min ให้ เปลี่ยนค่า min ตาม max
		if v < min_freq_Slider.Value {
			min_freq_Slider.SetValue(v)
			//fmt.Println("min >= max")
		}
		val_ch_max = uint64(v)
		entry_max.SetText(fmt.Sprintf("%.f", v))
		max_freq_Label.SetText(fmt.Sprintf("[ จำกัด - ความถี่สูงสุด ] %d kHz [ %.2f Ghz ]", val_ch_max, float64(val_ch_max)/1e6))
	}

	return min_freq_Slider, max_freq_Slider, min_freq_Label, max_freq_Label, entry_min, entry_max
}

func onButtonClickApply(selected []bool, min_freq_Slider, max_freq_Slider *widget.Slider, governorsST *widget.RadioGroup) {

	// อ่านค่าจากวิดเจต slider โดยตรง
	freq_min := uint64(min_freq_Slider.Value)
	freq_max := uint64(max_freq_Slider.Value)
	governorsSt := governorsST.Selected

	go func() { // รันใน goroutine ไม่ให้ UI ค้าง
		var scriptLines []string
		for idx, sel := range selected {
			if !sel {
				continue
			}
			scriptLines = append(scriptLines, fmt.Sprintf("echo %d | tee /sys/devices/system/cpu/cpu%d/cpufreq/scaling_max_freq", freq_max, idx))
			scriptLines = append(scriptLines, fmt.Sprintf("echo %d | tee /sys/devices/system/cpu/cpu%d/cpufreq/scaling_min_freq", freq_min, idx))
			scriptLines = append(scriptLines, fmt.Sprintf("echo %s | tee /sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor", governorsSt, idx))
		}

		if len(scriptLines) == 0 {
			//ฟังชั้น popup++
			fmt.Println("ไม่พบคอร์ที่เลือกให้ปรับค่า")
			return
		}

		script := strings.Join(scriptLines, "\n")

		cmd := exec.Command("pkexec", "bash", "-c", script)
		err := cmd.Run()
		if err != nil {
			fmt.Println("ล้มเหลว:", err)
			return
		}
		fmt.Println("สำเร็จ", "[ min ]", freq_min, "kHz", "[ max ]", freq_max, "kHz")
	}()

}

// ส่งออก
func CpuControl(w fyne.Window) fyne.CanvasObject {

	perCore := sysCPUFreqUpdate()
	info, _, _ := getCPUhardware(0)
	slider_min, slider_max, label_min, label_max, entry_min, entry_max := slider()

	chekCpu, selected, checkboxes, updateLabel := checkboxNumcpu()

	allCheck := widget.NewButton("เลือกทั้งหมด", func() {
		CheckAllBoxCpu(checkboxes, selected, updateLabel)
	})

	nonCheck := widget.NewButton("Reset", func() {
		nonCheckBoxCpu(checkboxes, selected, updateLabel)
	})

	//check govern...
	governors, governorsSt := GovernorscheckBox()

	apply := widget.NewButton("Apply", func() {
		onButtonClickApply(selected, slider_min, slider_max, governorsSt)
	})

	//min
	bt_min_n := NewRepeatButton("-", func() {
		onButtonMinN(slider_min)
	})
	bt_min_p := NewRepeatButton("+", func() {
		onButtonMinP(slider_min)
	})

	//test 10% new
	bt_101 := NewRepeatButton("10%", func() {
		newPercent(slider_min, 10)
	})
	//test 20% new
	bt_102 := NewRepeatButton("20%", func() {
		newPercent(slider_min, 20)
	})
	//test 30% new
	bt_103 := NewRepeatButton("30%", func() {
		newPercent(slider_min, 30)
	})
	//test 40% new
	bt_104 := NewRepeatButton("40%", func() {
		newPercent(slider_min, 40)
	})
	//test 50% new
	bt_105 := NewRepeatButton("50%", func() {
		newPercent(slider_min, 50)
	})
	//test 60% new
	bt_106 := NewRepeatButton("60%", func() {
		newPercent(slider_min, 60)
	})
	//test 70% new
	bt_107 := NewRepeatButton("70%", func() {
		newPercent(slider_min, 70)
	})
	//test 80% new
	bt_108 := NewRepeatButton("80%", func() {
		newPercent(slider_min, 80)
	})
	//test 90% new
	bt_109 := NewRepeatButton("90%", func() {
		newPercent(slider_min, 90)
	})
	//test 100% new
	bt_1010 := NewRepeatButton("100%", func() {
		newPercent(slider_min, 100)
	})

	//max
	bt_max_n := NewRepeatButton("-", func() {
		onButtonMaxN(slider_max)
	})

	bt_max_p := NewRepeatButton("+", func() {
		onButtonMaxP(slider_max)
	})

	//test 10% new
	bt_101M := NewRepeatButton("10%", func() {
		newPercent(slider_max, 10)
	})
	//test 20% new
	bt_102M := NewRepeatButton("20%", func() {
		newPercent(slider_max, 20)
	})
	//test 30% new
	bt_103M := NewRepeatButton("30%", func() {
		newPercent(slider_max, 30)
	})
	//test 40% new
	bt_104M := NewRepeatButton("40%", func() {
		newPercent(slider_max, 40)
	})
	//test 50% new
	bt_105M := NewRepeatButton("50%", func() {
		newPercent(slider_max, 50)
	})
	//test 60% new
	bt_106M := NewRepeatButton("60%", func() {
		newPercent(slider_max, 60)
	})
	//test 70% new
	bt_107M := NewRepeatButton("70%", func() {
		newPercent(slider_max, 70)
	})
	//test 80% new
	bt_108M := NewRepeatButton("80%", func() {
		newPercent(slider_max, 80)
	})
	//test 90% new
	bt_109M := NewRepeatButton("90%", func() {
		newPercent(slider_max, 90)
	})
	//test 100% new
	bt_1010M := NewRepeatButton("100%", func() {
		newPercent(slider_max, 100)
	})

	governorsAb := `conservative - เพิ่มความเร็วแบบค่อยเป็นค่อยไป
*ondemand - เพิ่มความเร็วเมื่อมีโหลด
ี**userspace - รักษาความเร็วคงที่ (ตามที่กำหนด)
powersave - ประหยัดพลังงาน
performance - ประสิทธิภาพสูงสุด
*schedutil - ปรับอัตโนมัติตามโหลด

- - - - - - - - - - - - - - - -

*ondemand: ทำงานเป็นกระบวนการแยกต่างหาก คอยสุ่มตรวจ (polling)
ปริมาณการใช้งาน CPU เป็นระยะ - ตอบสนองช้ากว่า schedutil
-ถ้ารออ่าน/เขียนข้อมูล (I/O) จะมองว่า CPU ทำงานหนักและ
เร่งความเร็วซีพียูสูงสุด ทำให้เปลืองพลังงานโดยเปล่าประโยชน์

*schedutil: ดึงข้อมูลการทำงานมาจากตัวจัดการคิวงานของเคอร์เนล
 (CPU Scheduler) โดยตรง ทำให้รู้สถานะความต้องการของโปรแกรมแบบเรียลไทม์
- ตอบสนองเร็วกว่า ondemand
- ฉลาดกว่า จะไม่เร่งความเร็วซีพียูเมื่อมีการรอ I/O

**userspace - ใน CPU Intel และ AMD รุ่นใหม่ อาจไม่ได้ผล เพราะ
การจัดการความถี่ถูกย้ายไปให้ตัวไดรเวอร์และเฟิร์มแวร์เป็นผู้ควบคุม`

	//ระยะห่าง
	space := canvas.NewRectangle(color.Transparent)
	space.SetMinSize(fyne.NewSize(0, 15))

	Info := container.NewVBox(info,
		space,
	)

	selectedCoreCpu := container.NewVBox(
		container.NewCenter(chekCpu),
		//layout.NewSpacer(),
		space,
		container.NewCenter(
			container.NewHBox(
				//container.NewGridWrap(fyne.NewSize(150, 35), allCheck),
				//container.NewGridWrap(fyne.NewSize(150, 35), nonCheck),
				container.NewGridWithColumns(2,
					container.NewGridWrap(fyne.NewSize(150, 35), allCheck),
					container.NewGridWrap(fyne.NewSize(150, 35), nonCheck),
				),
			)),
		space,
	)

	aboutBtn := widget.NewButton("ความหมาย", func() {
		dialog.ShowInformation("โหมดการทำงาน", governorsAb, w)
	})

	work := container.NewVBox(
		container.NewCenter(governors),
		//container.NewGridWrap(fyne.NewSize(35, 35), abbtn),
		space,
		container.NewCenter(
			container.NewHBox(
				container.NewGridWithColumns(1,
					container.NewGridWrap(fyne.NewSize(150, 35), aboutBtn)),
			)),
		space,
	)

	configMin := container.NewVBox(
		container.NewCenter(label_min),

		container.NewCenter(
			container.NewHBox(
				container.NewGridWrap(fyne.NewSize(100, 35), entry_min),
				bt_min_n, bt_min_p, bt_101, bt_102, bt_103, bt_104, bt_105, bt_106, bt_107, bt_108, bt_109, bt_1010),
		),
		slider_min,
	)

	configMax := container.NewVBox(
		container.NewCenter(label_max),

		container.NewCenter(
			container.NewHBox(
				container.NewGridWrap(fyne.NewSize(100, 35), entry_max),
				bt_max_n, bt_max_p, bt_101M, bt_102M, bt_103M, bt_104M, bt_105M, bt_106M, bt_107M, bt_108M, bt_109M, bt_1010M),
		),
		slider_max,
	)

	now := container.NewCenter(perCore)

	//perCore

	x := container.NewBorder(
		container.NewVBox(
			widget.NewCard("Default Kernel and Hardware", "", Info), //info,
			//widget.NewSeparator(),
			widget.NewCard("เลือกคอร์ CPU", "", selectedCoreCpu), //chekCpu,
			widget.NewCard("โหมดการทำงาน", "", work),             //governors,
			widget.NewCard("ความถี่ต่ำสุด", "", configMin),
			widget.NewCard("ความถี่สูงสุด", "", configMax),

			container.NewCenter(
				container.NewVBox(
					space,
					container.NewGridWrap(fyne.NewSize(200, 35), apply),
					space),
			),
			widget.NewCard("สถานะปัจจุบัน", "", now),
			//widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		nil,
	)

	return x
}
