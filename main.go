package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fredjeck/timely/pkg/platform"
	"github.com/fredjeck/timely/pkg/timeutils"
)

type systemStartupTime time.Time

const listHeight = 14
const defaultWidth = 20
const padding = 4
const maxWidth = 80

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	unreachedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000ff")).Bold(true)
	reachedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
	helperStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(string(i)))
}

type model struct {
	list              list.Model
	textInput         textinput.Model
	durations         timeutils.Durations
	total             time.Duration
	totalProvisionnal time.Duration
	overtime          time.Duration
	planned           string
	percentage        float64
	quitting          bool
	progress          progress.Model
	target            time.Duration
	startupTime       time.Time
}

func (m model) Append(t time.Time) model {
	m.durations = m.durations.Append(t)

	items := make([]list.Item, len(m.durations))
	for i, t := range m.durations.StringSlice() {
		items[i] = item(t)
	}
	m.list.SetItems(items)
	m.textInput.Reset()
	m = m.RecalculateDurations()
	return m
}

func (m model) RecalculateDurations() model {
	m.totalProvisionnal = timeutils.SumPairedDurationsWithNow(m.durations, time.Now())
	m.total = timeutils.SumPairedDurationsWithNow(m.durations, time.Time{})
	m.overtime = m.total - m.target
	last := m.durations.Last()
	if !last.IsZero() {
		remaining := m.target - m.total
		m.planned = last.Add(remaining).Format("15:04")
	}

	tmin := m.total.Minutes()
	ta := m.target.Minutes()
	if tmin > ta {
		m.percentage = 1
	} else {
		m.percentage = ((tmin * 100) / ta) / 100
	}
	return m
}

func initialModel(target time.Duration) model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 5
	ti.Width = 20

	l := list.New([]list.Item{}, itemDelegate{}, defaultWidth, listHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("x"),
				key.WithHelp("x", "delete"),
			),
		}
	}

	return model{
		textInput:         ti,
		list:              l,
		durations:         make(timeutils.Durations, 0),
		total:             0,
		totalProvisionnal: 0,
		quitting:          false,
		progress:          progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C")),
		target:            target,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case systemStartupTime:
		m.startupTime = time.Time(msg)
		if len(m.durations) == 0 {
			return m.Append(m.startupTime), nil
		}

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			t, err := timeutils.ParseTime(m.textInput.Value())
			if err != nil {
				m.textInput.Reset()
				return m, nil
			}
			return m.Append(t), nil
		case "x":
			m.list.RemoveItem(m.list.Index())
			m.durations = m.durations.RemoveItem(m.list.Index())
			m = m.RecalculateDurations()
			return m, nil
		}
	}

	// Handle both list and text input updates
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("Enjoy your day !")
	}

	style := reachedStyle
	if m.total < m.target {
		style = unreachedStyle
	}

	return style.Render(timeutils.FormatDuration(m.total)) +
		helperStyle.Render(" / "+timeutils.FormatDuration(m.target)) +
		helperStyle.Render(" • previsional ") + reachedStyle.Render(timeutils.FormatDuration(m.totalProvisionnal)) +
		helperStyle.Render(" • start ") + reachedStyle.Render(timeutils.FormatTime(m.startupTime)) +
		helperStyle.Render(" • exit ") + reachedStyle.Render(m.planned) +
		helperStyle.Render(" • overtime ") + reachedStyle.Render(timeutils.FormatDuration(m.overtime)) +
		"\n" +
		m.textInput.View() +
		"\n" +
		m.list.View() +
		"\n" +
		m.progress.ViewAs(m.percentage)
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please provide a target time in HH:MM format as an argument.")
		os.Exit(1)
	}

	targetTime, err := timeutils.ParseTime(os.Args[1])
	if err != nil {
		fmt.Println("Unknown target time", os.Args[1])
	}
	target := time.Duration(targetTime.Hour())*time.Hour + time.Duration(targetTime.Minute())*time.Minute

	p := tea.NewProgram(initialModel(target), tea.WithAltScreen())

	go func() {
		up, err := platform.Startup()
		if err != nil {
			return
		}
		p.Send(systemStartupTime(up))
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
