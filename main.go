package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const targetZone = "Asia/Kathmandu"

type screen int

const (
	inputScreen screen = iota
	zoneScreen
	resultScreen
)

type model struct {
	screen screen

	dateInput   textinput.Model
	timeInput   textinput.Model
	searchInput textinput.Model

	ampm string

	allZones []string
	matches  []string

	selectedIndex int
	selectedZone  string

	sourceTime time.Time
	targetTime time.Time

	err error
}

var (
	valueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FACC15"))
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7DD3FC"))

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A1A1AA"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#2563EB")).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D4D4D8")).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4ADE80"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#71717A"))
)

// Initialize application.
func initialModel() model {
	date := textinput.New()
	date.Placeholder = "YYYY-MM-DD"
	date.SetValue(time.Now().Format("2006-01-02"))
	date.Width = 35
	date.Focus()

	clock := textinput.New()
	clock.Placeholder = "HH:MM"
	clock.SetValue("10:00")
	clock.Width = 35

	search := textinput.New()
	search.Placeholder = "Search timezone... (e.g. Chicago, UTC)"
	search.Width = 50

	zones := loadTimeZones()

	m := model{
		screen:        inputScreen,
		dateInput:     date,
		timeInput:     clock,
		searchInput:   search,
		ampm:          "AM",
		allZones:      zones,
		selectedIndex: 0,
	}

	m.filterZones()

	return m
}

// Load timezone names from the system timezone database.
func loadTimeZones() []string {
	paths := []string{
		"/usr/share/zoneinfo/zone.tab",
		"/usr/share/lib/zoneinfo/zone.tab",
		"/usr/lib/zoneinfo/zone.tab",
		"/var/db/timezone/zoneinfo/zone.tab",
	}

	seen := make(map[string]bool)

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			fields := strings.Fields(line)

			if len(fields) >= 3 {
				seen[fields[2]] = true
			}
		}

		file.Close()

		if len(seen) > 0 {
			break
		}
	}

	// Common zones and useful aliases.
	fallback := []string{
		"UTC",
		"Etc/UTC",
		"Asia/Kathmandu",
		"Asia/Kolkata",
		"Asia/Dhaka",
		"Asia/Dubai",
		"Asia/Singapore",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Asia/Seoul",
		"Asia/Bangkok",
		"Asia/Jakarta",
		"Asia/Karachi",
		"Asia/Kabul",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"Europe/Moscow",
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"America/Toronto",
		"America/Vancouver",
		"America/Phoenix",
		"America/Anchorage",
		"America/Halifax",
		"America/Sao_Paulo",
		"Australia/Sydney",
		"Australia/Melbourne",
		"Australia/Perth",
		"Pacific/Auckland",
		"Pacific/Honolulu",
		"Africa/Cairo",
		"Africa/Johannesburg",
	}

	for _, zone := range fallback {
		seen[zone] = true
	}

	zones := make([]string, 0, len(seen))

	for zone := range seen {
		zones = append(zones, zone)
	}

	sort.Strings(zones)

	return zones
}

// MANUAL SEARCH IMPLEMENTATION.
// This does not use Bubble Tea list filtering.
func (m *model) filterZones() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))

	m.matches = m.matches[:0]

	for _, zone := range m.allZones {
		if query == "" ||
			strings.Contains(strings.ToLower(zone), query) {
			m.matches = append(m.matches, zone)
		}
	}

	if len(m.matches) == 0 {
		m.selectedIndex = 0
		return
	}

	if m.selectedIndex >= len(m.matches) {
		m.selectedIndex = len(m.matches) - 1
	}

	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Main update handler.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {

		case inputScreen:
			return m.updateInput(msg)

		case zoneScreen:
			return m.updateZone(msg)

		case resultScreen:
			return m.updateResult(msg)
		}
	}

	return m, nil
}

// Input screen.
func (m model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "ctrl+c":
		return m, tea.Quit

	case "tab":
		m.dateInput.Blur()
		m.timeInput.Focus()
		return m, textinput.Blink

	case "shift+tab":
		m.timeInput.Blur()
		m.dateInput.Focus()
		return m, textinput.Blink

	case "ctrl+a":
		if m.ampm == "AM" {
			m.ampm = "PM"
		} else {
			m.ampm = "AM"
		}
		return m, nil

	case "enter":
		if err := m.validateInput(); err != nil {
			m.err = err
			return m, nil
		}

		m.err = nil
		m.searchInput.SetValue("")
		m.selectedIndex = 0
		m.filterZones()

		m.dateInput.Blur()
		m.timeInput.Blur()

		m.searchInput.Focus()
		m.screen = zoneScreen

		return m, textinput.Blink
	}

	var cmd tea.Cmd

	if m.dateInput.Focused() {
		m.dateInput, cmd = m.dateInput.Update(msg)
	} else {
		m.timeInput, cmd = m.timeInput.Update(msg)
	}

	return m, cmd
}

// Validate date and time.
func (m *model) validateInput() error {
	_, err := time.Parse(
		"2006-01-02",
		m.dateInput.Value(),
	)
	if err != nil {
		return fmt.Errorf("invalid date: use YYYY-MM-DD")
	}

	_, err = time.Parse(
		"03:04",
		m.timeInput.Value(),
	)
	if err != nil {
		return fmt.Errorf("invalid time: use HH:MM")
	}

	return nil
}

// Timezone search and navigation.
func (m model) updateZone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.searchInput.Blur()
		m.dateInput.Focus()
		m.screen = inputScreen
		return m, textinput.Blink

	case "up":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
		return m, nil

	case "down":
		if m.selectedIndex < len(m.matches)-1 {
			m.selectedIndex++
		}
		return m, nil

	case "enter":
		if len(m.matches) == 0 {
			return m, nil
		}

		m.selectedZone = m.matches[m.selectedIndex]

		if err := m.convert(); err != nil {
			m.err = err
			return m, nil
		}

		m.err = nil
		m.searchInput.Blur()
		m.screen = resultScreen

		return m, nil

	case "ctrl+u":
		m.searchInput.SetValue("")
		m.selectedIndex = 0
		m.filterZones()
		return m, nil

	case "pgdown":
		m.selectedIndex += 10
		if m.selectedIndex >= len(m.matches) {
			m.selectedIndex = len(m.matches) - 1
		}
		return m, nil

	case "pgup":
		m.selectedIndex -= 10
		if m.selectedIndex < 0 {
			m.selectedIndex = 0
		}
		return m, nil
	}

	// Every other key goes to the dedicated search input.
	// Search is updated on every keystroke.
	var cmd tea.Cmd

	m.searchInput, cmd = m.searchInput.Update(msg)

	// Recalculate matching zones immediately.
	m.selectedIndex = 0
	m.filterZones()

	return m, cmd
}

// Result screen.
func (m model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "ctrl+c", "q", "esc":
		return m, tea.Quit

	case "enter":
		m.screen = inputScreen
		m.dateInput.Focus()
		m.err = nil
		return m, textinput.Blink

	case "z":
		m.screen = zoneScreen
		m.searchInput.SetValue("")
		m.selectedIndex = 0
		m.filterZones()
		m.searchInput.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

// Convert source timezone to Kathmandu.
func (m *model) convert() error {
	sourceLocation, err := time.LoadLocation(m.selectedZone)
	if err != nil {
		return fmt.Errorf("cannot load timezone: %w", err)
	}

	targetLocation, err := time.LoadLocation(targetZone)
	if err != nil {
		return fmt.Errorf("cannot load Kathmandu timezone: %w", err)
	}

	date, _ := time.Parse(
		"2006-01-02",
		m.dateInput.Value(),
	)

	clock, _ := time.Parse(
		"03:04",
		m.timeInput.Value(),
	)

	hour := clock.Hour()
	minute := clock.Minute()

	if m.ampm == "PM" && hour < 12 {
		hour += 12
	}

	if m.ampm == "AM" && hour == 12 {
		hour = 0
	}

	m.sourceTime = time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		hour,
		minute,
		0,
		0,
		sourceLocation,
	)

	m.targetTime = m.sourceTime.In(targetLocation)

	return nil
}

// Input view.
func (m model) inputView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("TIMEZONE CONVERTER"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("DATE"))
	b.WriteString("\n")
	b.WriteString(m.dateInput.View())
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("TIME (HH:MM)"))
	b.WriteString("\n")
	b.WriteString(m.timeInput.View())
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("AM / PM"))
	b.WriteString("\n")
	b.WriteString(successStyle.Render("[" + m.ampm + "]"))
	b.WriteString("   Ctrl+A to toggle")
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("TARGET TIMEZONE"))
	b.WriteString("\n")
	b.WriteString(successStyle.Render(targetZone))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString(mutedStyle.Render(
		"TAB: Next field | SHIFT+TAB: Previous field",
	))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(
		"CTRL+A: Toggle AM/PM | ENTER: Continue | CTRL+C: Quit",
	))

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

// Search screen.
func (m model) zoneView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("SELECT SOURCE TIMEZONE"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("YOUR INPUT"))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf(
		"%s %s %s\n",
		m.dateInput.Value(),
		m.timeInput.Value(),
		m.ampm,
	))

	b.WriteString("\n")

	b.WriteString(labelStyle.Render("SEARCH TIMEZONE"))
	b.WriteString("\n")

	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf(
		"%d matching timezones\n\n",
		len(m.matches),
	))

	if len(m.matches) == 0 {
		b.WriteString(errorStyle.Render("No timezone found."))
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render(
			"Try: Chicago, Kolkata, UTC, Kathmandu",
		))
	} else {
		const maxVisible = 15

		start := 0

		if m.selectedIndex >= maxVisible {
			start = m.selectedIndex - maxVisible + 1
		}

		end := start + maxVisible

		if end > len(m.matches) {
			end = len(m.matches)
		}

		for i := start; i < end; i++ {
			zone := m.matches[i]

			if i == m.selectedIndex {
				b.WriteString(selectedStyle.Render(
					"> " + zone,
				))
			} else {
				b.WriteString(normalStyle.Render(
					"  " + zone,
				))
			}

			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(
		"UP/DOWN: Navigate | ENTER: Select | CTRL+U: Clear search",
	))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(
		"ESC: Back | CTRL+C: Quit",
	))

	return lipgloss.NewStyle().
		Padding(1, 3).
		Render(b.String())
}

// Result screen.
func (m model) resultView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("TIME CONVERSION RESULT"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("SOURCE TIMEZONE"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render(m.selectedZone))
	b.WriteString("\n")

	b.WriteString(m.sourceTime.Format(
		"Monday, 02 January 2006",
	))
	b.WriteString("\n")

	b.WriteString(valueStyle.Render(
		m.sourceTime.Format("03:04:05 PM"),
	))
	b.WriteString("\n")

	b.WriteString(m.sourceTime.Format("15:04:05"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("CONVERTED TIME"))
	b.WriteString("\n")

	b.WriteString(successStyle.Render(targetZone))
	b.WriteString("\n")

	b.WriteString(m.targetTime.Format(
		"Monday, 02 January 2006",
	))
	b.WriteString("\n")

	b.WriteString(successStyle.Render(
		m.targetTime.Format("03:04:05 PM"),
	))
	b.WriteString("\n")

	b.WriteString(successStyle.Render(
		m.targetTime.Format("15:04:05"),
	))
	b.WriteString("\n\n")

	_, sourceOffset := m.sourceTime.Zone()
	_, targetOffset := m.targetTime.Zone()

	b.WriteString(labelStyle.Render("UTC OFFSETS"))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf(
		"Source: %s\n",
		formatOffset(sourceOffset),
	))

	b.WriteString(fmt.Sprintf(
		"Kathmandu: %s\n",
		formatOffset(targetOffset),
	))

	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(
		"ENTER: Convert another time",
	))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(
		"Z: Select another timezone",
	))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(
		"Q: Quit",
	))

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

func formatOffset(seconds int) string {
	sign := "+"

	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}

	return fmt.Sprintf(
		"UTC%s%02d:%02d",
		sign,
		seconds/3600,
		(seconds%3600)/60,
	)
}

// Main renderer.
func (m model) View() string {
	switch m.screen {
	case inputScreen:
		return m.inputView()

	case zoneScreen:
		return m.zoneView()

	case resultScreen:
		return m.resultView()
	}

	return ""
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
