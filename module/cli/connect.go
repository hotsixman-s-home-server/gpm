package cli

import (
	"fmt"
	"gpm/module/client"
	"gpm/module/logger"
	"gpm/module/types"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#32CD32")) // 밝은 녹색 적용

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF5F87"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00D7FF"))
)

type model struct {
	name          string
	client        *client.Client
	messageChan   chan types.LogMessage
	closeChan     chan bool
	logViewport   viewport.Model
	errViewport   viewport.Model
	textInput     textinput.Model
	logs          []string
	errors        []string
	width, height int
	err           error

	// 히스토리 기능 관련 필드
	history      []string
	historyIndex int
	tempInput    string
}

func initialModel(name string, c *client.Client, mChan chan types.LogMessage, cChan chan bool) model {
	ti := textinput.New()
	ti.Placeholder = "Type a command..."
	ti.Focus()
	ti.CharLimit = 156

	return model{
		name:         name,
		client:       c,
		messageChan:  mChan,
		closeChan:    cChan,
		textInput:    ti,
		logs:         make([]string, 0),
		errors:       make([]string, 0),
		logViewport:  viewport.New(0, 0),
		errViewport:  viewport.New(0, 0),
		history:      make([]string, 0),
		historyIndex: -1,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.waitForMessage(),
	)
}

func (m model) waitForMessage() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg, ok := <-m.messageChan:
			if !ok {
				return tea.Quit
			}
			return msg
		case <-m.closeChan:
			return tea.Quit
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		lvCmd tea.Cmd
		evCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			cmdStr := m.textInput.Value()
			if strings.TrimSpace(cmdStr) != "" {
				// 히스토리 저장 및 인덱스 초기화
				if m.history[len(m.history)-1] != cmdStr {
					m.history = append(m.history, cmdStr)
				}
				m.historyIndex = -1

				m.client.Command(cmdStr)
				m.textInput.SetValue("")
			}

		case tea.KeyUp:
			if len(m.history) > 0 {
				if m.historyIndex == -1 {
					m.tempInput = m.textInput.Value()
				}
				if m.historyIndex < len(m.history)-1 {
					m.historyIndex++
					idx := len(m.history) - 1 - m.historyIndex
					m.textInput.SetValue(m.history[idx])
					m.textInput.SetCursor(len(m.history[idx]))
				}
			}

		case tea.KeyDown:
			if m.historyIndex > -1 {
				m.historyIndex--
				if m.historyIndex == -1 {
					m.textInput.SetValue(m.tempInput)
				} else {
					idx := len(m.history) - 1 - m.historyIndex
					m.textInput.SetValue(m.history[idx])
				}
				m.textInput.SetCursor(len(m.textInput.Value()))
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// 고정 높이 계산 (헤더, 타이틀, 여백, 입력창 테두리 등 포함)
		fixedHeight := 12
		availableHeight := m.height - fixedHeight
		if availableHeight < 4 {
			availableHeight = 4
		}

		vHeight := availableHeight / 2

		m.errViewport.Width = m.width - 4
		m.errViewport.Height = vHeight

		m.logViewport.Width = m.width - 4
		m.logViewport.Height = vHeight

		m.textInput.Width = m.width - 6

	case types.LogMessage:
		if msg.Type == "error" {
			m.errors = append(m.errors, logger.SErrorln(msg.Message))
			if len(m.errors) > 500 {
				m.errors = m.errors[len(m.errors)-500:]
			}
			m.errViewport.SetContent(strings.Join(m.errors, "\n"))
			m.errViewport.GotoBottom()
		} else {
			m.logs = append(m.logs, logger.SLogln(msg.Message))
			if len(m.logs) > 500 {
				m.logs = m.logs[len(m.logs)-500:]
			}
			m.logViewport.SetContent(strings.Join(m.logs, "\n"))
			m.logViewport.GotoBottom()
		}
		return m, m.waitForMessage()

	case error:
		m.err = msg
		return m, tea.Quit
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.logViewport, lvCmd = m.logViewport.Update(msg)
	m.errViewport, evCmd = m.errViewport.Update(msg)

	return m, tea.Batch(tiCmd, lvCmd, evCmd)
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// 1. 헤더
	header := titleStyle.Render(fmt.Sprintf(" GPM Connect: %s ", m.name))

	// 2. 에러 섹션 (상단)
	errSection := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Bold(true).Render(" Errors"),
		errorStyle.Width(m.width-2).Render(m.errViewport.View()),
	)

	// 3. 로그 섹션 (중단)
	logSection := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32")).Bold(true).Render(" Logs"),
		logStyle.Width(m.width-2).Render(m.logViewport.View()),
	)

	// 4. 입력 섹션 (하단)
	inputBox := inputStyle.Width(m.width - 2).Render(m.textInput.View())

	// 레이아웃 전체 조립
	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		errSection,
		logSection,
		inputBox,
	)

	// 화면 짤림 방지: 터미널 높이에 고정
	return lipgloss.NewStyle().
		MaxHeight(m.height).
		MaxWidth(m.width).
		Render(ui)
}

var connectCmd = &cobra.Command{
	Use:   "connect [name]",
	Short: "Connect to process",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		closeChan := make(chan bool)
		messageChan := make(chan types.LogMessage)
		conn, reader, err := client.MakeUDSConn()
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		c, err := client.NewClient(args[0], conn, reader, messageChan, closeChan)
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		p := tea.NewProgram(initialModel(args[0], c, messageChan, closeChan), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
