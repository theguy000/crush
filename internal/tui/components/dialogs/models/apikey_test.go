package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyInput_PasteHandling(t *testing.T) {
	tests := []struct {
		name           string
		state          APIKeyInputState
		pasteContent   string
		shouldProcess  bool
		expectedFocused bool
	}{
		{
			name:           "paste in initial state",
			state:          APIKeyInputStateInitial,
			pasteContent:   "sk-test123456789",
			shouldProcess:  true,
			expectedFocused: true,
		},
		{
			name:           "paste in error state",
			state:          APIKeyInputStateError,
			pasteContent:   "sk-test123456789",
			shouldProcess:  true,
			expectedFocused: true,
		},
		{
			name:           "paste in verifying state",
			state:          APIKeyInputStateVerifying,
			pasteContent:   "sk-test123456789",
			shouldProcess:  false,
			expectedFocused: false,
		},
		{
			name:           "paste in verified state",
			state:          APIKeyInputStateVerified,
			pasteContent:   "sk-test123456789",
			shouldProcess:  false,
			expectedFocused: false,
		},
		{
			name:           "empty paste",
			state:          APIKeyInputStateInitial,
			pasteContent:   "",
			shouldProcess:  false,
			expectedFocused: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewAPIKeyInput()
			input.state = tt.state
			
			// Update state presentation to set the correct focus state
			input.updateStatePresentation()
			
			// Create paste message
			pasteMsg := tea.PasteMsg(tt.pasteContent)
			
			// Process the paste message
			_, cmd := input.Update(pasteMsg)
			
			// Verify focus state
			if tt.expectedFocused {
				assert.True(t, input.Focused(), "Input should be focused")
			} else {
				assert.False(t, input.Focused(), "Input should not be focused")
			}
			
			// If paste should be processed, verify content was added
			if tt.shouldProcess && tt.pasteContent != "" {
				assert.Contains(t, input.Value(), tt.pasteContent, "Pasted content should be in input value")
			}
			
			// Ensure no panic occurred
			require.NotPanics(t, func() {
				if cmd != nil {
					// Execute command if present
					_ = cmd()
				}
			})
		})
	}
}

func TestAPIKeyInput_SetValue(t *testing.T) {
	input := NewAPIKeyInput()
	
	// Test setting a valid value
	testValue := "sk-test123456789"
	require.NotPanics(t, func() {
		input.SetValue(testValue)
	})
	
	assert.Equal(t, testValue, input.Value())
	assert.True(t, input.Focused(), "Input should remain focused after SetValue")
}

func TestAPIKeyInput_Reset(t *testing.T) {
	input := NewAPIKeyInput()
	
	// Set some initial state
	input.SetValue("sk-test123456789")
	input.state = APIKeyInputStateError
	
	// Reset the input
	require.NotPanics(t, func() {
		input.Reset()
	})
	
	assert.Equal(t, "", input.Value(), "Value should be cleared")
	assert.Equal(t, APIKeyInputStateInitial, input.state, "State should be reset to initial")
	assert.True(t, input.Focused(), "Input should be focused after reset")
}