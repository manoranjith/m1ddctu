package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type Slider struct {
	Label    string
	Shortcut string
	Param    string
	Gauge    *widgets.Gauge
}

const (
	appWidth  = 50
	appHeight = 8
)

var (
	sliderConfigs = []struct {
		Label    string
		Shortcut string
		Command  string
	}{
		{"(B)rightness", "B", "luminance"},
		{"(C)ontrast", "C", "contrast"},
	}
	presets = [][]int{
		{30, 40},
		{70, 20},
		{50, 60},
	}
)

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	sliders := initSliders()

	presetDropdown := initPresetDropdown()

	grid := setupLayout(presetDropdown, sliders)

	ui.Render(grid)

	handleEvents(grid, sliders, presetDropdown)
}

func initSliders() []*Slider {
	sliders := make([]*Slider, len(sliderConfigs))
	for i, config := range sliderConfigs {
		g := widgets.NewGauge()
		g.Title = config.Label
		g.Percent = executeCommand("get", config.Command, 0)
		sliders[i] = &Slider{Label: config.Label, Shortcut: config.Shortcut, Param: config.Command, Gauge: g}
	}
	return sliders
}

func initPresetDropdown() *widgets.List {
	presetNames := []string{"<Custom>", "Preset 1", "Preset 2", "Preset 3"}
	presetDropdown := widgets.NewList()
	presetDropdown.Title = "Presets"
	presetDropdown.Rows = presetNames
	presetDropdown.SelectedRowStyle = ui.NewStyle(ui.ColorYellow, ui.ColorClear)
	presetDropdown.WrapText = false
	return presetDropdown
}

func setupLayout(presetDropdown *widgets.List, sliders []*Slider) *ui.Grid {
	grid := ui.NewGrid()
	grid.SetRect(0, 0, appWidth, appHeight)
	grid.Set(
		ui.NewCol(2.0/6,
			ui.NewRow(1, presetDropdown)),
		ui.NewCol(4.0/6,
			ui.NewRow(1.0/2, sliders[0].Gauge),
			ui.NewRow(1.0/2, sliders[1].Gauge)),
	)
	return grid
}

func handleEvents(grid *ui.Grid, sliders []*Slider, presetDropdown *widgets.List) {
	selectedSliderIndex := -1
	presetDropdownActive := false

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "<C-c>", "q":
			if selectedSliderIndex != -1 {
				sliders[selectedSliderIndex].Gauge.BorderStyle.Fg = ui.ColorWhite
				selectedSliderIndex = -1
			}
			ui.Close()
			return
		case "<Down>", "<Up>", "<Left>", "<Right>", "j", "J", "k", "K":
			handleArrowKeys(e.ID, sliders, &selectedSliderIndex, presetDropdown, &presetDropdownActive)
		case "<Enter>":
			applyPreset(presetDropdown.SelectedRow, sliders)
		default:
			handleSliderSelection(e.ID, sliders, &selectedSliderIndex, presetDropdown, &presetDropdownActive)
		}
		ui.Render(grid)
	}
}

func handleArrowKeys(e string, sliders []*Slider, selectedSliderIndex *int, presetDropdown *widgets.List, presetDropdownActive *bool) {
	if *presetDropdownActive {
		switch e {
		case "<Down>", "<Left>", "j", "J":
			presetDropdown.ScrollDown()
		case "<Up>", "<Right>", "k", "K":
			presetDropdown.ScrollUp()
		}
		return
	}

	if *selectedSliderIndex == -1 {
		return
	}

	delta := 0
	switch e {
	case "<Down>", "<Left>", "j", "J":
		delta = -5
	case "<Up>", "<Right>", "k", "K":
		delta = 5
	}

	sliders[*selectedSliderIndex].Gauge.Percent = executeCommand("chg", sliders[*selectedSliderIndex].Param, delta)
}

func applyPreset(selectedPreset int, sliders []*Slider) {
	if selectedPreset <= 0 || selectedPreset > len(presets) {
		return
	}

	for i, value := range presets[selectedPreset-1] {
		sliders[i].Gauge.Percent = value
	}
}

func handleSliderSelection(e string, sliders []*Slider, selectedSliderIndex *int, presetDropdown *widgets.List, presetDropdownActive *bool) {
	for i, slider := range sliders {
		if strings.ToUpper(e) == slider.Shortcut {
			if *selectedSliderIndex != -1 && *selectedSliderIndex != i {
				sliders[*selectedSliderIndex].Gauge.BorderStyle.Fg = ui.ColorWhite
			}
			*presetDropdownActive = false
			presetDropdown.BorderStyle.Fg = ui.ColorWhite
			slider.Gauge.BorderStyle.Fg = ui.ColorYellow
			*selectedSliderIndex = i
			break
		}
	}
}

func executeCommand(action, param string, value int) int {
	var cmd *exec.Cmd
	switch action {
	case "get":
		cmd = exec.Command("m1ddc", action, param)
	case "chg":
		cmd = exec.Command("m1ddc", action, param, fmt.Sprint(value))
	default:
		log.Fatalf("Unsupported action: %v", action)
		return 0
	}
	// fmt.Println(cmd.String())

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %v", err)
		return 0
	}

	ans, err := strconv.Atoi(strings.TrimSpace(out.String()))
	if err != nil {
		log.Printf("Error converting command output to integer: %v", err)
		return 0
	}

	return ans
}
