#!/bin/bash

# Test script for Crush API Key Input Paste Fix
# This script helps verify that the paste functionality works correctly

echo "Testing Crush API Key Input Paste Fix"
echo "====================================="

# Check if the binary was built
if [ ! -f "./crush" ]; then
    echo "Error: Crush binary not found. Please run 'go build -o crush .' first."
    exit 1
fi

echo "✓ Crush binary found"

# Run tests
echo "Running tests..."
go test ./internal/tui/components/dialogs/models/ -v
if [ $? -eq 0 ]; then
    echo "✓ All tests passed"
else
    echo "✗ Some tests failed"
    exit 1
fi

# Check for any compilation errors
echo "Building application..."
go build -o crush .
if [ $? -eq 0 ]; then
    echo "✓ Build successful"
else
    echo "✗ Build failed"
    exit 1
fi

echo ""
echo "Manual Testing Instructions:"
echo "==========================="
echo "1. Run: ./crush"
echo "2. Select Grok LLM or any provider that requires API key"
echo "3. When prompted for API key, try Ctrl+V to paste"
echo "4. Verify that the application doesn't crash"
echo "5. Verify that the pasted content appears in the input field"
echo "6. Verify that the cursor remains focused after paste"
echo ""
echo "Expected Behavior:"
echo "- Ctrl+V should work without crashing"
echo "- Pasted API key should appear in the input field"
echo "- Cursor should remain focused on the input field"
echo "- Application should continue to function normally"
echo ""
echo "If you encounter any issues, please report them with:"
echo "- The exact steps to reproduce"
echo "- Any error messages or crash logs"
echo "- Your operating system and terminal type"