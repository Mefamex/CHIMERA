package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// -- AYARLAR --
const LogDirRelative = "../../data/logs"

// -- VERİ YAPILARI --
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Msg       string `json:"msg"`
	TraceID   string `json:"trace_id"`
	Source    string
	Raw       string
	TimeObj   time.Time
}

// Uygulama Durumu (State)
type AppState struct {
	AllLogs      []LogEntry
	FilteredLogs []LogEntry

	UniqueFiles      map[string]bool
	UniqueComponents map[string]bool

	SelectedFiles      map[string]bool
	SelectedComponents map[string]bool

	SearchText   string
	AutoScroll   bool
	LiveWatch    bool
	HasSelection bool

	FileOffsets map[string]int64
	mu          sync.RWMutex
}

var state = AppState{
	UniqueFiles:        make(map[string]bool),
	UniqueComponents:   make(map[string]bool),
	SelectedFiles:      make(map[string]bool),
	SelectedComponents: make(map[string]bool),
	FileOffsets:        make(map[string]int64),
	AutoScroll:         true,
	LiveWatch:          true,
	SearchText:         "",
	HasSelection:       false,
}

func main() {
	a := app.New()
	w := a.NewWindow("CHIMERA Log Inspector (Live)")
	w.Resize(fyne.NewSize(1400, 900))

	logPath := findLogDir()
	fmt.Println("logPath:", logPath)

	// -- UI ELEMENTLERİ --

	// 1. Detay Paneli
	detailRichText := widget.NewRichTextFromMarkdown("")
	detailRichText.Wrapping = fyne.TextWrapBreak
	detailScroll := container.NewVScroll(detailRichText)

	copyBtn := widget.NewButtonWithIcon("Kopyala", theme.ContentCopyIcon(), func() {})
	copyBtn.Disable()

	detailHeader := container.NewBorder(
		nil, nil,
		widget.NewLabelWithStyle("📋 LOG DETAYI", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		copyBtn,
	)
	detailPanel := container.NewBorder(detailHeader, nil, nil, nil, detailScroll)

	// 2. Log Tablosu
	logTable := widget.NewTable(
		func() (int, int) {
			state.mu.RLock()
			defer state.mu.RUnlock()
			return len(state.FilteredLogs), 4
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			icon := widget.NewIcon(theme.InfoIcon())
			label := widget.NewLabel("Text")
			return container.NewStack(bg, icon, label)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			state.mu.RLock()
			if id.Row >= len(state.FilteredLogs) {
				state.mu.RUnlock()
				return
			}
			entry := state.FilteredLogs[id.Row]
			state.mu.RUnlock()

			stack := o.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			iconWidget := stack.Objects[1].(*widget.Icon)
			labelWidget := stack.Objects[2].(*widget.Label)

			// Renkler
			switch entry.Level {
			case "ERROR":
				bg.FillColor = color.RGBA{R: 80, G: 20, B: 20, A: 200}
			case "WARN":
				bg.FillColor = color.RGBA{R: 60, G: 60, B: 0, A: 200}
			default:
				bg.FillColor = color.Transparent
			}
			bg.Refresh()

			iconWidget.Hide()
			labelWidget.Show()

			switch id.Col {
			case 0: // ICON
				labelWidget.Hide()
				iconWidget.Show()
				if entry.Level == "ERROR" {
					iconWidget.SetResource(theme.WarningIcon())
				} else if entry.Source == "chat" {
					iconWidget.SetResource(theme.MailComposeIcon())
				} else if entry.Source == "tools" {
					iconWidget.SetResource(theme.SettingsIcon())
				} else {
					iconWidget.SetResource(theme.InfoIcon())
				}
			case 1: // TIME
				if len(entry.Timestamp) > 19 {
					labelWidget.SetText(entry.Timestamp[11:19])
				} else {
					labelWidget.SetText(entry.Timestamp)
				}
			case 2: // COMPONENT
				labelWidget.SetText(entry.Component)
			case 3: // MESSAGE
				labelWidget.SetText(entry.Msg)
			}
		},
	)

	logTable.SetColumnWidth(0, 40)
	logTable.SetColumnWidth(1, 90)
	logTable.SetColumnWidth(2, 120)
	logTable.SetColumnWidth(3, 600)

	logTable.OnSelected = func(id widget.TableCellID) {
		state.mu.RLock()
		if id.Row < len(state.FilteredLogs) {
			selectedLog := state.FilteredLogs[id.Row]
			copyBtn.Enable()
			copyBtn.OnTapped = func() { w.Clipboard().SetContent(selectedLog.Raw) }

			detailRichText.Segments = parseColoredJSON(selectedLog.Raw)
			detailRichText.Refresh()
			detailScroll.Offset = fyne.NewPos(0, 0)
			detailScroll.Refresh()
		}
		state.mu.RUnlock()

		state.mu.Lock()
		state.HasSelection = true
		state.mu.Unlock()
	}

	// -- FİLTRE UI --
	fileChecksBox := container.NewVBox()
	compChecksBox := container.NewVBox()
	fileScroll := container.NewVScroll(fileChecksBox)
	compScroll := container.NewVScroll(compChecksBox)

	var applyFilters func(resetDetail bool)
	var updateFilterUI func()

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Loglarda ara (Mesaj, TraceID)...")
	searchEntry.OnChanged = func(text string) {
		state.mu.Lock()
		state.SearchText = text
		state.mu.Unlock()
		applyFilters(true)
	}

	updateFilterUI = func() {
		fileChecksBox.RemoveAll()
		compChecksBox.RemoveAll()

		state.mu.RLock()
		var fileKeys []string
		for k := range state.UniqueFiles {
			fileKeys = append(fileKeys, k)
		}
		var compKeys []string
		for k := range state.UniqueComponents {
			compKeys = append(compKeys, k)
		}

		availableComponents := make(map[string]bool)
		for _, log := range state.AllLogs {
			if state.SelectedFiles[log.Source] {
				availableComponents[log.Component] = true
			}
		}
		state.mu.RUnlock()

		sort.Strings(fileKeys)
		sort.Strings(compKeys)

		// DOSYALAR UI
		fileChecksBox.Add(widget.NewLabelWithStyle("📂 DOSYALAR", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		btnAllFiles := container.NewHBox(
			widget.NewButton("Tümü", func() {
				state.mu.Lock()
				for k := range state.UniqueFiles {
					state.SelectedFiles[k] = true
				}
				state.mu.Unlock()
				updateFilterUI()
				applyFilters(true)
			}),
			widget.NewButton("Hiçbiri", func() {
				state.mu.Lock()
				for k := range state.UniqueFiles {
					state.SelectedFiles[k] = false
				}
				state.mu.Unlock()
				updateFilterUI()
				applyFilters(true)
			}),
		)
		fileChecksBox.Add(btnAllFiles)

		for _, fname := range fileKeys {
			f := fname
			check := widget.NewCheck(f, func(b bool) {
				state.mu.Lock()
				state.SelectedFiles[f] = b
				state.mu.Unlock()
				applyFilters(true)
			})
			state.mu.RLock()
			check.Checked = state.SelectedFiles[f]
			state.mu.RUnlock()
			fileChecksBox.Add(check)
		}

		// COMPONENTLER UI
		compChecksBox.Add(widget.NewLabelWithStyle("🧩 BİLEŞENLER", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		btnAllComps := container.NewHBox(
			widget.NewButton("Tümü", func() {
				state.mu.Lock()
				for k := range state.UniqueComponents {
					state.SelectedComponents[k] = true
				}
				state.mu.Unlock()
				updateFilterUI()
				applyFilters(true)
			}),
			widget.NewButton("Hiçbiri", func() {
				state.mu.Lock()
				for k := range state.UniqueComponents {
					state.SelectedComponents[k] = false
				}
				state.mu.Unlock()
				updateFilterUI()
				applyFilters(true)
			}),
		)
		compChecksBox.Add(btnAllComps)

		for _, cname := range compKeys {
			c := cname
			check := widget.NewCheck(c, func(b bool) {
				state.mu.Lock()
				state.SelectedComponents[c] = b
				state.mu.Unlock()
				applyFilters(true)
			})
			state.mu.RLock()
			check.Checked = state.SelectedComponents[c]
			state.mu.RUnlock()
			if !availableComponents[c] {
				check.Disable()
			} else {
				check.Enable()
			}
			compChecksBox.Add(check)
		}
		fileChecksBox.Refresh()
		compChecksBox.Refresh()
	}

	applyFilters = func(resetDetail bool) {
		state.mu.Lock()
		var newFiltered []LogEntry
		searchTerm := strings.ToLower(state.SearchText)

		for _, log := range state.AllLogs {
			if !state.SelectedFiles[log.Source] {
				continue
			}
			if !state.SelectedComponents[log.Component] {
				continue
			}

			if searchTerm != "" {
				inMsg := strings.Contains(strings.ToLower(log.Msg), searchTerm)
				inTrace := strings.Contains(strings.ToLower(log.TraceID), searchTerm)
				if !inMsg && !inTrace {
					continue
				}
			}
			newFiltered = append(newFiltered, log)
		}
		state.FilteredLogs = newFiltered
		autoScroll := state.AutoScroll
		hasSelection := state.HasSelection
		if resetDetail {
			state.HasSelection = false
		}
		state.mu.Unlock()

		logTable.Refresh()
		if autoScroll && !hasSelection {
			logTable.ScrollToBottom()
		}
		if resetDetail {
			detailRichText.Segments = nil
			detailRichText.Refresh()
			copyBtn.Disable()
		}
	}

	// -- KLASÖR YÖNETİMİ --
	var folders []string
	selectedFolder := -1
	folderList := widget.NewList(
		func() int { return len(folders) },
		func() fyne.CanvasObject { return widget.NewLabel("Log Folder") },
		func(i widget.ListItemID, o fyne.CanvasObject) { o.(*widget.Label).SetText(folders[i]) },
	)

	folderList.OnSelected = func(id widget.ListItemID) {
		selectedFolder = int(id)
		fullPath := filepath.Join(logPath, folders[id])
		loadLogs(fullPath, true)
		state.mu.Lock()
		for k := range state.UniqueFiles {
			state.SelectedFiles[k] = true
		}
		for k := range state.UniqueComponents {
			state.SelectedComponents[k] = true
		}
		state.mu.Unlock()
		updateFilterUI()
		applyFilters(true)
	}

	refreshFolders := func() {
		folders = []string{}
		files, err := os.ReadDir(logPath)
		if err == nil {
			for i := len(files) - 1; i >= 0; i-- {
				if files[i].IsDir() {
					folders = append(folders, files[i].Name())
				}
			}
		}
		selectedFolder = -1
		folderList.Refresh()
	}
	refreshFolders()

	// -- CANLI İZLEME (Ticker) --
	// Arka planda çalışacak Goroutine
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			state.mu.RLock()
			isLive := state.LiveWatch
			state.mu.RUnlock()

			// Eğer Canlı mod açıksa ve bir klasör seçiliyse
			if isLive && selectedFolder != -1 {
				fullPath := filepath.Join(logPath, folders[selectedFolder])
				hasUpdates := loadLogs(fullPath, false)
				if hasUpdates {
					applyFilters(false)
				}
			}
		}
	}()

	// -- ÜST BAR --

	// Canlı İzle Checkbox'ı
	liveCheck := widget.NewCheck("🔴 Canlı İzle", func(b bool) {
		state.mu.Lock()
		state.LiveWatch = b
		state.mu.Unlock()
	})
	// Varsayılan kapalı, ikon kırmızı/yeşil yapılabilir ama basit checkbox yeterli.

	// Oto Kaydır Checkbox'ı
	autoScrollCheck := widget.NewCheck("⬇️ Oto. Kaydır", func(b bool) {
		state.mu.Lock()
		state.AutoScroll = b
		state.mu.Unlock()
	})
	autoScrollCheck.Checked = true

	topToolBar := container.NewBorder(
		nil, nil,
		widget.NewLabelWithStyle("CHIMERA LOGS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(
			liveCheck,
			autoScrollCheck,
			widget.NewButtonWithIcon("Yenile", theme.ViewRefreshIcon(), func() {
				if selectedFolder >= 0 && selectedFolder < len(folders) {
					fullPath := filepath.Join(logPath, folders[selectedFolder])
					hasUpdates := loadLogs(fullPath, false)
					if hasUpdates {
						updateFilterUI()
						applyFilters(false)
					}
				} else {
					refreshFolders()
				}
			}),
		),
		searchEntry,
	)

	// -- YERLEŞİM --
	rightSplit := container.NewVSplit(container.NewPadded(fileScroll), container.NewPadded(compScroll))
	rightSplit.SetOffset(0.4)
	centerSplit := container.NewVSplit(logTable, detailPanel)
	centerSplit.SetOffset(0.7)
	mainSplit := container.NewHSplit(centerSplit, rightSplit)
	mainSplit.SetOffset(0.8)

	fullLayout := container.NewHSplit(
		container.NewBorder(nil, nil, nil, nil, folderList),
		container.NewBorder(topToolBar, nil, nil, nil, mainSplit),
	)
	fullLayout.SetOffset(0.15)

	w.SetContent(fullLayout)
	w.ShowAndRun()
}

// -- FONKSİYONLAR --

func parseColoredJSON(rawJSON string) []widget.RichTextSegment {
	var segments []widget.RichTextSegment
	scanner := bufio.NewScanner(strings.NewReader(rawJSON))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			segments = append(segments, &widget.TextSegment{
				Text:  parts[0] + ":",
				Style: widget.RichTextStyle{ColorName: theme.ColorNameError, TextStyle: fyne.TextStyle{Bold: true, Monospace: true}, Inline: true},
			})
			valText := parts[1]
			colorName := theme.ColorNameSuccess
			if strings.Contains(valText, "\"") {
				colorName = theme.ColorNamePrimary
			}
			segments = append(segments, &widget.TextSegment{
				Text:  valText + "\n",
				Style: widget.RichTextStyle{ColorName: colorName, TextStyle: fyne.TextStyle{Monospace: true}, Inline: true},
			})
		} else {
			segments = append(segments, &widget.TextSegment{
				Text:  line + "\n",
				Style: widget.RichTextStyle{ColorName: theme.ColorNameForeground, TextStyle: fyne.TextStyle{Monospace: true}},
			})
		}
	}
	return segments
}

func findLogDir() string {
	if _, err := os.Stat(LogDirRelative); !os.IsNotExist(err) {
		return LogDirRelative
	}
	if _, err := os.Stat("data/logs"); !os.IsNotExist(err) {
		return "data/logs"
	}
	return "./"
}

// Loglari Yukler
// forceFull: True ise her seyi bastan okur (Klasor degistiginde)
// False ise sadece son kaldigi yerden devam eder (Canli Izleme)
func loadLogs(dirPath string, forceFull bool) bool {
	state.mu.Lock()
	defer state.mu.Unlock()

	if forceFull {
		state.AllLogs = []LogEntry{}
		state.UniqueFiles = make(map[string]bool)
		state.UniqueComponents = make(map[string]bool)
		state.FileOffsets = make(map[string]int64)
	}

	files, _ := os.ReadDir(dirPath)
	var newEntries []LogEntry
	hasUpdates := false

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".jsonl" {
			continue
		}

		sourceType := strings.TrimSuffix(f.Name(), ".jsonl")
		state.UniqueFiles[sourceType] = true
		path := filepath.Join(dirPath, f.Name())

		file, err := os.Open(path)
		if err != nil {
			continue
		}

		offset := state.FileOffsets[path]
		stat, _ := file.Stat()
		if stat.Size() < offset {
			offset = 0
		}

		_, _ = file.Seek(offset, 0)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}

			var rawMap map[string]interface{}
			if json.Unmarshal([]byte(line), &rawMap) == nil {
				var entry LogEntry
				if v, ok := rawMap["timestamp"].(string); ok {
					entry.Timestamp = v
				}
				if v, ok := rawMap["level"].(string); ok {
					entry.Level = v
				}
				if v, ok := rawMap["component"].(string); ok {
					entry.Component = v
				}
				if v, ok := rawMap["msg"].(string); ok {
					entry.Msg = v
				}
				if v, ok := rawMap["trace_id"].(string); ok {
					entry.TraceID = v
				}

				entry.Source = sourceType
				entry.Raw = formatJSON(line)
				t, _ := time.Parse("2006-01-02T15:04:05.000", entry.Timestamp)
				entry.TimeObj = t
				if entry.Component == "" {
					entry.Component = "UNKNOWN"
				}

				state.UniqueComponents[entry.Component] = true
				newEntries = append(newEntries, entry)
				hasUpdates = true
			}
		}

		currentPos, _ := file.Seek(0, 1)
		state.FileOffsets[path] = currentPos
		file.Close()
	}

	if hasUpdates {
		state.AllLogs = append(state.AllLogs, newEntries...)
		sort.Slice(state.AllLogs, func(i, j int) bool {
			return state.AllLogs[i].TimeObj.Before(state.AllLogs[j].TimeObj)
		})
	}

	return hasUpdates
}

func formatJSON(raw string) string {
	var pretty map[string]interface{}
	json.Unmarshal([]byte(raw), &pretty)
	b, _ := json.MarshalIndent(pretty, "", "  ")
	return string(b)
}
