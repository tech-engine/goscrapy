package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tech-engine/goscrapy/pkg/builtin/middlewares"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	metricLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	metricValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA"))

	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1).
			MarginRight(1)

	logStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)
)

type model struct {
	logBuffer  *LogBuffer
	lastSnap   stats.GlobalSnapshot
	currSnap   stats.GlobalSnapshot
	currentRPS float64
	width      int
	height     int
}

func newModel(lb *LogBuffer) model {
	return model{
		logBuffer: lb,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case stats.GlobalSnapshot:
		m.lastSnap = m.currSnap
		m.currSnap = msg

		currHttp, okCurr := m.currSnap.Components["http"].(middlewares.HttpMetrics)
		lastHttp, okLast := m.lastSnap.Components["http"].(middlewares.HttpMetrics)

		// calculate current RPS
		if okCurr && okLast && !currHttp.StartTime.IsZero() && m.currSnap.Interval > 0 {
			delta := currHttp.TotalRequests - lastHttp.TotalRequests
			m.currentRPS = float64(delta) / m.currSnap.Interval.Seconds()
		}

		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Initialising..."
	}

	header := titleStyle.Render("GoScrapy TUI Dashboard")

	currHttp, _ := m.currSnap.Components["http"].(middlewares.HttpMetrics)

	uptime := fmt.Sprintf("Uptime: %s", m.currSnap.Uptime.Truncate(time.Second))

	// stats
	statsRows := []string{
		fmt.Sprintf("%s %d", metricLabelStyle.Render("Requests:"), currHttp.TotalRequests),
		fmt.Sprintf("%s %s", metricLabelStyle.Render("Bandwidth:"), formatBytes(currHttp.TotalBytes)),
		fmt.Sprintf("%s %.2f req/s", metricLabelStyle.Render("Speed:    "), m.currentRPS),
		fmt.Sprintf("%s %s", metricLabelStyle.Render("Latency:  "), currHttp.AvgLatency.Truncate(time.Millisecond)),
	}
	statsBox := boxStyle.Width(m.width/2 - 4).Render(lipgloss.JoinVertical(lipgloss.Left, statsRows...))

	// Status codes
	statusRows := []string{metricLabelStyle.Render("Status Codes:")}
	for code, count := range currHttp.StatusCodes {
		style := successStyle
		if code >= 400 {
			style = errorStyle
		} else if code >= 300 {
			style = warnStyle
		}
		statusRows = append(statusRows, fmt.Sprintf("  %d: %s", code, style.Render(fmt.Sprint(count))))
	}
	statusBox := boxStyle.Width(m.width/2 - 4).Render(lipgloss.JoinVertical(lipgloss.Left, statusRows...))

	panels := lipgloss.JoinHorizontal(lipgloss.Top, statsBox, statusBox)

	// Logs
	logs := m.logBuffer.GetLogs()
	logContent := ""
	maxLogLines := m.height - 15
	if maxLogLines < 5 {
		maxLogLines = 5
	}

	start := 0
	if len(logs) > maxLogLines {
		start = len(logs) - maxLogLines
	}

	for _, l := range logs[start:] {
		color := lipgloss.Color("#A0A0A0")
		if strings.Contains(l, "DEBUG") {
			color = lipgloss.Color("#626262")
		} else if strings.Contains(l, "INFO") {
			color = lipgloss.Color("#00D7AF")
		} else if strings.Contains(l, "WARN") {
			color = lipgloss.Color("#FFAF00")
		} else if strings.Contains(l, "ERROR") {
			color = lipgloss.Color("#FF5F87")
		}

		logContent += lipgloss.NewStyle().Foreground(color).Render(l) + "\n"
	}

	logView := logStyle.Width(m.width - 4).Height(maxLogLines).Render(logContent)

	footer := "Press 'q' or 'Ctrl+C' to exit dashboard and return to terminal"

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		uptime,
		"\n",
		panels,
		"\n",
		logView,
		"\n",
		footer,
	)
}

// dashboard implements IDashboard.
type dashboard struct {
	program   *tea.Program
	logBuffer *LogBuffer
	logger    core.ILogger
}

// Represents a dashboard that can observe stats and run.
type IDashboard interface {
	stats.IStatsObserver
	Run() error
}

// New creates a new Dashboard.
func New(logger core.ILogger) IDashboard {
	lb := NewLogBuffer(500)

	program := tea.NewProgram(
		newModel(lb),
		tea.WithAltScreen(),
	)

	return &dashboard{
		program:   program,
		logBuffer: lb,
		logger:    logger,
	}
}

// implements stats.StatsObserver.
func (d *dashboard) OnSnapshot(snap stats.GlobalSnapshot) {
	d.program.Send(snap)
}

// Run starts the dashboard. It redirects the provided logger to the dashboard's
// internal buffer, runs the TUI (blocking), and restores the logger on exit.
func (d *dashboard) Run() error {
	// redirect logger to capture buffer if it's configurable
	if cl, ok := d.logger.(core.IConfigurableLogger); ok {
		cl.WithWriter(NewLogWriter(d.logBuffer))
		defer cl.WithWriter(os.Stderr)
	}

	// blocks until quit
	_, err := d.program.Run()

	return err
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
