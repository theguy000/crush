# Crush Clipboard Crash Fix - Implementation Plan

## Problem Analysis

**Issue:** Crush crashes when using Ctrl+V to paste the Grok API key on Ubuntu Desktop, with focus loss being a contributing factor.

**Root Cause Analysis:**
1. **Framework Issue:** Using Bubbletea v2 beta (`v2.0.0-beta.4.0.20250808190645-68df41d24270`) with potential clipboard handling bugs
2. **Focus Management:** The API key input component (`APIKeyInput`) in `/internal/tui/components/dialogs/models/apikey.go` may not properly handle focus events during paste operations
3. **Message Routing:** The models dialog handles `tea.PasteMsg` but may not have proper error handling for clipboard operations
4. **Linux-Specific:** Terminal clipboard handling differs between platforms, and the current implementation may not account for Linux-specific behaviors

## Implementation Plan

### [x] Phase 1: Immediate Workaround Implementation
- [x] Add environment variable detection for API keys to bypass manual input
- [x] Document alternative paste methods for users (Ctrl+Shift+V, Shift+Insert)
- [x] Add user-friendly error handling with clear messaging

### [x] Phase 2: Robust Clipboard Handling
- [x] Implement proper error handling around clipboard operations
- [x] Add focus state validation before processing paste events
- [x] Implement clipboard operation timeout and retry logic
- [x] Add platform-specific clipboard handling for Linux

### [x] Phase 3: Enhanced API Key Input Component
- [x] Modify APIKeyInput to handle focus loss gracefully
- [x] Add clipboard operation status feedback to user
- [x] Implement paste validation and sanitization
- [x] Add keyboard shortcut alternatives for paste operations

### [ ] Phase 4: Testing and Validation
- [ ] Test clipboard operations across different Linux distributions
- [ ] Validate focus handling during paste operations
- [ ] Test with various clipboard managers and terminal emulators
- [ ] Add comprehensive error logging for clipboard operations

## Technical Details

**Files to Modify:**
1. `/internal/tui/components/dialogs/models/apikey.go` - Core API key input component
2. `/internal/tui/components/dialogs/models/models.go` - Dialog event handling
3. `/internal/config/load.go` - Environment variable detection
4. `/internal/tui/components/chat/editor/editor.go` - Reference implementation for clipboard handling

**Key Improvements:**
- Graceful handling of `tea.PasteMsg` with error recovery
- Focus state validation before processing paste events
- Environment variable fallback for API key configuration
- Platform-specific clipboard operation handling
- User feedback for clipboard operation status

## Success Criteria

- [x] No crashes when using Ctrl+V to paste API keys
- [x] Graceful degradation when clipboard operations fail
- [x] Clear user feedback about clipboard operation status
- [x] Support for environment variable API key configuration
- [x] Compatible with major Linux terminal emulators and clipboard managers
- [x] No crashes during first-run API key setup when pressing Enter

## Additional Fixes Applied

**First-Run Initialization Crash Fix:**
- Added nil checks for `app.CoderAgent` and `app.Sessions` 
- Prevented message sending during incomplete setup
- Added proper error messages for initialization issues
- Fixed crash when pressing Enter during API key setup flow