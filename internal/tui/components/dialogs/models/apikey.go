package models

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/tui/styles"
	"github.com/charmbracelet/lipgloss/v2"
)

type APIKeyInputState int

const (
	APIKeyInputStateInitial APIKeyInputState = iota
	APIKeyInputStateVerifying
	APIKeyInputStateVerified
	APIKeyInputStateError
)

type APIKeyStateChangeMsg struct {
	State APIKeyInputState
}

type APIKeyInput struct {
	input        textinput.Model
	width        int
	spinner      spinner.Model
	providerName string
	state        APIKeyInputState
	title        string
	showTitle    bool
}

func NewAPIKeyInput() *APIKeyInput {
	t := styles.CurrentTheme()

	ti := textinput.New()
	ti.Placeholder = "Enter your API key..."
	ti.SetVirtualCursor(false)
	ti.Prompt = "> "
	ti.SetStyles(t.S().TextInput)
	ti.Focus()

	return &APIKeyInput{
		input: ti,
		state: APIKeyInputStateInitial,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(t.S().Base.Foreground(t.Green)),
		),
		providerName: "Provider",
		showTitle:    true,
	}
}

func (a *APIKeyInput) SetProviderName(name string) {
	a.providerName = name
	a.updateStatePresentation()
}

func (a *APIKeyInput) SetShowTitle(show bool) {
	a.showTitle = show
}

func (a *APIKeyInput) GetTitle() string {
	return a.title
}

func (a *APIKeyInput) Init() tea.Cmd {
	a.updateStatePresentation()
	return a.spinner.Tick
}

func (a *APIKeyInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Add defensive error handling for all message processing
	// to prevent crashes from focus loss or clipboard operations
	defer func() {
		if r := recover(); r != nil {
			slog.Error("APIKeyInput update failed", "error", r, "msgType", fmt.Sprintf("%T", msg))
			// Continue execution gracefully
		}
	}()

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if a.state == APIKeyInputStateVerifying {
			var cmd tea.Cmd
			a.spinner, cmd = a.spinner.Update(msg)
			a.updateStatePresentation()
			return a, cmd
		}
		return a, nil
	case APIKeyStateChangeMsg:
		a.state = msg.State
		var cmd tea.Cmd
		if msg.State == APIKeyInputStateVerifying {
			cmd = a.spinner.Tick
		}
		a.updateStatePresentation()
		return a, cmd
	case tea.PasteMsg:
		// Special handling for paste messages to prevent crashes
		// Validate input focus state before processing paste
		slog.Debug("Processing paste message in API key input", "hasValue", a.input.Value() != "")

		// Ensure the input is properly focused before attempting paste
		if a.state == APIKeyInputStateInitial || a.state == APIKeyInputStateError {
			a.input.Focus()
		}
	}

	// Safely update the underlying text input with error recovery
	var cmd tea.Cmd
	func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Text input update failed", "error", r, "msgType", fmt.Sprintf("%T", msg))
				// Keep current state if update fails
			}
		}()
		a.input, cmd = a.input.Update(msg)
	}()

	return a, cmd
}

func (a *APIKeyInput) updateStatePresentation() {
	t := styles.CurrentTheme()

	prefixStyle := t.S().Base.
		Foreground(t.Primary)
	accentStyle := t.S().Base.Foreground(t.Green).Bold(true)
	errorStyle := t.S().Base.Foreground(t.Cherry)

	switch a.state {
	case APIKeyInputStateInitial:
		titlePrefix := prefixStyle.Render("Enter your ")
		a.title = titlePrefix + accentStyle.Render(a.providerName+" API Key") + prefixStyle.Render(".")
		a.input.SetStyles(t.S().TextInput)
		a.input.Prompt = "> "
	case APIKeyInputStateVerifying:
		titlePrefix := prefixStyle.Render("Verifying your ")
		a.title = titlePrefix + accentStyle.Render(a.providerName+" API Key") + prefixStyle.Render("...")
		ts := t.S().TextInput
		// make the blurred state be the same
		ts.Blurred.Prompt = ts.Focused.Prompt
		a.input.Prompt = a.spinner.View()
		a.input.Blur()
	case APIKeyInputStateVerified:
		a.title = accentStyle.Render(a.providerName+" API Key") + prefixStyle.Render(" validated.")
		ts := t.S().TextInput
		// make the blurred state be the same
		ts.Blurred.Prompt = ts.Focused.Prompt
		a.input.SetStyles(ts)
		a.input.Prompt = styles.CheckIcon + " "
		a.input.Blur()
	case APIKeyInputStateError:
		a.title = errorStyle.Render("Invalid ") + accentStyle.Render(a.providerName+" API Key") + errorStyle.Render(". Try again?")
		ts := t.S().TextInput
		ts.Focused.Prompt = ts.Focused.Prompt.Foreground(t.Cherry)
		a.input.Focus()
		a.input.SetStyles(ts)
		a.input.Prompt = styles.ErrorIcon + " "
	}
}

func (a *APIKeyInput) View() string {
	inputView := a.input.View()

	dataPath := config.GlobalConfigData()
	dataPath = strings.Replace(dataPath, config.HomeDir(), "~", 1)
	helpText := styles.CurrentTheme().S().Muted.
		Render(fmt.Sprintf("This will be written to the global configuration: %s", dataPath))

	// Add helpful tips for alternative input methods
	t := styles.CurrentTheme()
	var tips string
	if a.state == APIKeyInputStateInitial {
		tips = lipgloss.JoinVertical(
			lipgloss.Left,
			t.S().Muted.Render("💡 Tips:"),
			t.S().Muted.Render("  • Try Ctrl+Shift+V or Shift+Insert to paste"),
			t.S().Muted.Render(fmt.Sprintf("  • Set %s_API_KEY environment variable to skip this step", strings.ToUpper(a.providerName))),
		)
	}

	var content string
	if a.showTitle && a.title != "" {
		if tips != "" {
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				a.title,
				"",
				inputView,
				"",
				tips,
				"",
				helpText,
			)
		} else {
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				a.title,
				"",
				inputView,
				"",
				helpText,
			)
		}
	} else {
		if tips != "" {
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				inputView,
				"",
				tips,
				"",
				helpText,
			)
		} else {
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				inputView,
				"",
				helpText,
			)
		}
	}

	return content
}

func (a *APIKeyInput) Cursor() *tea.Cursor {
	cursor := a.input.Cursor()
	if cursor != nil && a.showTitle {
		cursor.Y += 2 // Adjust for title and spacing
	}
	return cursor
}

func (a *APIKeyInput) Value() string {
	return a.input.Value()
}

func (a *APIKeyInput) Tick() tea.Cmd {
	if a.state == APIKeyInputStateVerifying {
		return a.spinner.Tick
	}
	return nil
}

func (a *APIKeyInput) SetWidth(width int) {
	a.width = width
	a.input.SetWidth(width - 4)
}

func (a *APIKeyInput) Reset() {
	a.state = APIKeyInputStateInitial
	a.input.SetValue("")
	a.input.Focus()
	a.updateStatePresentation()
}
