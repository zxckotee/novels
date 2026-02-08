#!/usr/bin/env bash
set -euo pipefail

mkdir -p /tmp/vnc /data

XVFB_W="${XVFB_W:-1365}"
XVFB_H="${XVFB_H:-768}"
XVFB_DPI="${XVFB_DPI:-96}"

# Start virtual X server
Xvfb :99 -screen 0 "${XVFB_W}x${XVFB_H}x24" -dpi "${XVFB_DPI}" -ac +extension RANDR &

# Simple window manager for a usable desktop
fluxbox >/tmp/fluxbox.log 2>&1 &

# VNC server attached to Xvfb
x11vnc -display :99 -forever -shared -rfbport 5900 -nopw -o /tmp/x11vnc.log >/dev/null 2>&1 &

# noVNC web client (HTTP) -> websockify -> VNC
websockify --web=/usr/share/novnc/ 0.0.0.0:7900 localhost:5900 >/tmp/websockify.log 2>&1 &

echo "noVNC is up on :7900"
echo "Running: python /app/shuba_browser_session.py $*"

exec python /app/shuba_browser_session.py "$@"
