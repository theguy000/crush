#!/bin/bash

# Test script for Crush paste functionality
echo "Testing Crush paste functionality..."

# Check if crush binary exists
if [ ! -f "./crush" ]; then
    echo "Error: crush binary not found. Please build it first with 'go build -o crush .'"
    exit 1
fi

echo "Crush binary found. Testing paste functionality..."
echo "This test will verify that Ctrl+V works correctly in the API key input dialog."

# Instructions for manual testing
echo ""
echo "Manual Test Instructions:"
echo "1. Start Crush: ./crush"
echo "2. Select Grok LLM when prompted"
echo "3. When the API key input dialog appears, try Ctrl+V to paste your API key"
echo "4. The application should not freeze and should accept the pasted key"
echo ""
echo "Expected behavior:"
echo "- Ctrl+V should work without freezing the application"
echo "- The API key should be pasted into the input field"
echo "- The cursor should remain focused on the input field"
echo ""
echo "If you encounter any issues, check the debug logs for more information."
echo "Debug logs will show 'DEBUG:' messages indicating paste operations."

echo "Test setup complete. Please run the manual test as described above."