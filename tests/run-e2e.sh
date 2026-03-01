#!/run/current-system/sw/bin/bash
set -e

# If not already in nix-shell, re-run this script inside nix-shell
# This ensures Playwright browsers and other dependencies are available
if [ -z "$IN_NIX_SHELL" ]; then
  echo "Re-running inside nix-shell for Playwright browsers..."
  exec nix-shell --run "bash tests/run-e2e.sh"
fi

# Ensure port is free before starting
lsof -ti:34115 2>/dev/null | xargs kill 2>/dev/null || true
sleep 1

# Start nvim with an empty file - test will open test.go via keystrokes
# This ensures the test browser captures the content-updated events
export NVIM_GUI_TEST_FILE="tests/fixtures/empty.go"

# Start wails dev in background
wails dev -tags webkit2_41 &
WAILS_PID=$!

cleanup() {
    kill $WAILS_PID 2>/dev/null || true
    lsof -ti:34115 2>/dev/null | xargs kill 2>/dev/null || true
    sleep 1
}
trap cleanup EXIT

# Wait for the Go binary to fully start (not just the Vite server).
# The "navigate to:" message appears after the binary is compiled and running.
echo "Waiting for Wails app to fully start..."
wails_log=/tmp/wails-e2e-$$.log
# Tail wails output to a log file so we can detect readiness
tail -f /proc/$WAILS_PID/fd/1 2>/dev/null > "$wails_log" &
TAIL_PID=$!

for i in $(seq 1 90); do
    if curl -s http://localhost:34115 > /dev/null 2>&1; then
        echo "Dev server accepting connections."
        break
    fi
    if [ "$i" -eq 90 ]; then
        echo "Timed out waiting for dev server."
        kill $TAIL_PID 2>/dev/null || true
        exit 1
    fi
    sleep 1
done

# Give neovim time to fully start, attach UI, open test file, and render
echo "Waiting for nvim to initialize..."
sleep 8

kill $TAIL_PID 2>/dev/null || true

# Run Playwright tests
npx playwright test
