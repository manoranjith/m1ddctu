package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"gopkg.in/yaml.v2"
)

type Preset struct {
	Brightness int `yaml:"brightness"`
	Contrast   int `yaml:"contrast"`
}

type Slider struct {
	Label    string
	Shortcut string
	Param    string
	Gauge    *widgets.Gauge
}

const (
	appWidth  = 50
	appHeight = 8

	Brightness = 0
	Contrast   = 1
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
	presets []Preset
)

func main() {
	err := ui.Init()
	if err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	presetFile := filepath.Join(os.Getenv("HOME"), ".config", "m1ddctui", "presets.yaml")
	presets, err = loadPresetsFromFile(presetFile)
	if err != nil {
		log.Fatalf("Error loading presets: %v", err)
	}
	presetNames := formatPresetNames(presets)

	sliders := initSliders()
	presetDropdown := initPresetDropdown(presetNames)

	grid := setupLayout(presetDropdown, sliders)
	ui.Render(grid)
	handleEvents(grid, sliders, presetDropdown)
}

func loadPresetsFromFile(filename string) ([]Preset, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var presets []Preset
	err = yaml.Unmarshal(data, &presets)
	if err != nil {
		return nil, err
	}

	return presets, nil
}

func formatPresetNames(presets []Preset) []string {
	var names []string
	for i, preset := range presets {
		name := fmt.Sprintf("%d: B %d, C %d", i+1, preset.Brightness, preset.Contrast)
		names = append(names, name)
	}
	return names
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

func initPresetDropdown(presetNames []string) *widgets.List {
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
		case "P", "p":
			presetDropdownActive = true
			presetDropdown.BorderStyle.Fg = ui.ColorYellow
			for _, slider := range sliders {
				slider.Gauge.BorderStyle.Fg = ui.ColorWhite
				selectedSliderIndex = -1
			}

		case "<Enter>":
			applyPreset(presets[presetDropdown.SelectedRow], sliders)
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			presetIndex := int(e.ID[0] - '1')
			if presetIndex < len(presets) {
				applyPreset(presets[presetIndex], sliders)
			}
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

func applyPreset(preset Preset, sliders []*Slider) {
	sliders[Brightness].Gauge.Percent = executeCommand("set", sliders[Brightness].Param, preset.Brightness)
	sliders[Contrast].Gauge.Percent = executeCommand("set", sliders[Contrast].Param, preset.Contrast)
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
	case "set":
		cmd = exec.Command("m1ddc", action, param, fmt.Sprint(value))
	default:
		log.Fatalf("Unsupported action: %v", action)
		return 0
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %v", err)
		return 0
	}

	switch action {
	case "get", "chg":
		ans, err := strconv.Atoi(strings.TrimSpace(out.String()))
		if err != nil {
			log.Printf("Error converting command output to integer: %v", err)
			return 0
		}
		return ans
	case "set":
		return value
	default:
		return 0
	}
}
