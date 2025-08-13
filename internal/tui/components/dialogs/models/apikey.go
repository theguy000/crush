package models

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/v2/key"
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
	focused      bool  // Track focus state to prevent freeze issues
	lastFocusTime time.Time  // Track when focus was last confirmed
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
		focused:      true,
		lastFocusTime: time.Now(),
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
	case tea.FocusMsg:
		// Handle focus gained
		a.focused = true
		a.lastFocusTime = time.Now()
		a.input.Focus()
		slog.Debug("API key input gained focus")
		return a, nil
		
	case tea.BlurMsg:
		// Handle focus lost
		a.focused = false
		slog.Debug("API key input lost focus")
		// Don't blur the input immediately to prevent clipboard freeze
		return a, nil
		
	case tea.KeyPressMsg:
		// Update focus state on any key press
		a.focused = true
		a.lastFocusTime = time.Now()
		
		// Intercept Ctrl+V before it reaches textinput to prevent freeze
		if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+v"))) {
			return a.handleSafePaste()
		}
		
		// Also handle alternative paste shortcuts that might work better on Linux
		if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+shift+v"))) || 
		   key.Matches(msg, key.NewBinding(key.WithKeys("shift+insert"))) {
			return a.handleSafePaste()
		}
		
		// Let other key events pass through to textinput
		return a.updateTextInput(msg)
		
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
		// Handle paste messages safely - this is a fallback
		return a.handlePasteMsg(msg)
	default:
		// Handle all other messages safely
		return a.updateTextInput(msg)
	}
}

// handleSafePaste safely handles paste operations without freezing
func (a *APIKeyInput) handleSafePaste() (tea.Model, tea.Cmd) {
	slog.Debug("Handling safe paste operation in API key input", "focused", a.focused)
	
	// Only allow paste in states where input is editable
	if a.state != APIKeyInputStateInitial && a.state != APIKeyInputStateError {
		slog.Debug("Paste ignored - input not in editable state", "state", a.state)
		return a, nil
	}
	
	// Check if we have recent focus - prevent freeze from focus loss
	timeSinceFocus := time.Since(a.lastFocusTime)
	if !a.focused || timeSinceFocus > 2*time.Second {
		slog.Warn("Paste operation blocked due to focus loss", "focused", a.focused, "timeSinceFocus", timeSinceFocus)
		
		// Try to regain focus and show user feedback
		a.input.Focus()
		a.focused = true
		a.lastFocusTime = time.Now()
		
		// Return without attempting clipboard operation to prevent freeze
		return a, nil
	}
	
	// Ensure input is focused before clipboard operation
	a.input.Focus()
	a.focused = true
	a.lastFocusTime = time.Now()
	
	// Use a timeout for clipboard operations to prevent indefinite freeze
	type clipboardResult struct {
		content string
		err     error
	}
	
	resultChan := make(chan clipboardResult, 1)
	
	// Run clipboard operation in goroutine with timeout
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Clipboard read panic in goroutine", "error", r)
				resultChan <- clipboardResult{err: fmt.Errorf("clipboard panic: %v", r)}
			}
		}()
		
		content, err := clipboard.ReadAll()
		resultChan <- clipboardResult{content: content, err: err}
	}()
	
	// Wait for result with timeout
	select {
	case result := <-resultChan:
		if result.err != nil {
			slog.Error("Failed to read clipboard", "error", result.err)
			return a, nil
		}
		
		// Sanitize clipboard content (remove newlines, trim spaces)
		clipboardContent := strings.TrimSpace(strings.ReplaceAll(result.content, "\n", ""))
		
		if clipboardContent == "" {
			slog.Debug("Clipboard is empty or contains only whitespace")
			return a, nil
		}
		
		slog.Debug("Successfully read from clipboard", "length", len(clipboardContent))
		
		// Safely set the value in textinput
		currentValue := a.input.Value()
		cursorPos := a.input.Position()
		
		// Insert clipboard content at cursor position
		newValue := currentValue[:cursorPos] + clipboardContent + currentValue[cursorPos:]
		a.input.SetValue(newValue)
		
		// Move cursor to end of pasted content
		a.input.SetCursor(cursorPos + len(clipboardContent))
		
		return a, nil
		
	case <-time.After(1 * time.Second):
		// Timeout - prevent freeze
		slog.Warn("Clipboard operation timed out, preventing freeze")
		return a, nil
	}
}

// handlePasteMsg handles tea.PasteMsg safely
func (a *APIKeyInput) handlePasteMsg(msg tea.PasteMsg) (tea.Model, tea.Cmd) {
	slog.Debug("Processing paste message in API key input", "hasValue", a.input.Value() != "")
	
	// Validate input focus state before processing paste
	if a.state == APIKeyInputStateInitial || a.state == APIKeyInputStateError {
		a.input.Focus()
	}
	
	// Try to use the paste message content directly
	pasteContent := strings.TrimSpace(strings.ReplaceAll(string(msg), "\n", ""))
	if pasteContent != "" {
		currentValue := a.input.Value()
		cursorPos := a.input.Position()
		newValue := currentValue[:cursorPos] + pasteContent + currentValue[cursorPos:]
		a.input.SetValue(newValue)
		a.input.SetCursor(cursorPos + len(pasteContent))
	}
	
	return a, nil
}

// updateTextInput safely updates the underlying text input
func (a *APIKeyInput) updateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Ensure focus is maintained to prevent freeze issues
		if a.focused {
			a.input.Focus()
		}
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
		a.focused = true
		a.lastFocusTime = time.Now()
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
		focusWarning := ""
		if !a.focused || time.Since(a.lastFocusTime) > 2*time.Second {
			focusWarning = t.S().Muted.Foreground(t.Cherry).Render("  ⚠ Click here to focus before pasting\n")
		}
		
		tips = lipgloss.JoinVertical(
			lipgloss.Left,
			t.S().Muted.Render("💡 Tips:"),
			focusWarning,
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
	a.focused = true
	a.lastFocusTime = time.Now()
	a.updateStatePresentation()
}
