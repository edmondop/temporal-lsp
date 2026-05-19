#!/bin/bash
set -e

# Start code-server in background
echo "Starting code-server..."
code-server --bind-addr 0.0.0.0:8080 --auth none /workspace &
CODE_PID=$!

# Wait for code-server to be ready
echo "Waiting for code-server to be ready..."
for i in $(seq 1 30); do
  if curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo "code-server is ready"
    break
  fi
  sleep 1
done

# Run Playwright tests
echo "Running Playwright tests..."
cd /playwright
mkdir -p /output/videos /output/snapshots

# If baselines exist, run in comparison mode; otherwise generate them
if [ -d "/output/snapshots" ] && [ "$(ls -A /output/snapshots 2>/dev/null)" ]; then
  echo "Baselines found — running in comparison mode..."
  npx playwright test --reporter=list
else
  echo "No baselines — generating snapshots..."
  npx playwright test --reporter=list --update-snapshots
fi

# Copy video files to output root for easy access
echo "Copying videos..."
find /output -name "*.webm" -exec cp {} /output/ \; 2>/dev/null || true

# Convert video to GIF for README embedding
echo "Converting video to GIF..."
VIDEO=$(find /output -maxdepth 1 -name "*.webm" | head -1)
if [ -n "$VIDEO" ]; then
  ffmpeg -i "$VIDEO" -vf "fps=10,scale=1400:-1" -y /output/vscode-demo.gif 2>/dev/null
fi

echo "Tests complete."
echo "Screenshots:"
ls -la /output/*.png 2>/dev/null || true
echo "Videos:"
ls -la /output/*.webm 2>/dev/null || true
echo "GIF:"
ls -la /output/*.gif 2>/dev/null || true

# Keep running if interactive
if [ "$1" = "keep" ]; then
  wait $CODE_PID
fi
