#!/bin/bash

echo "Starting Crush with debug logging enabled..."
echo "This will show detailed logs of clipboard operations."
echo "Try to reproduce the freeze and watch the logs."
echo ""

# Run with debug logging enabled  
CRUSH_LOG_LEVEL=debug /tmp/crush-debug -d 2>&1 | tee /tmp/crush-debug.log

echo ""
echo "Logs saved to /tmp/crush-debug.log"
echo "If the application froze, the logs will show exactly where it happened."