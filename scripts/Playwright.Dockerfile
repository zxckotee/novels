FROM mcr.microsoft.com/playwright/python:v1.44.0-jammy

WORKDIR /app

# noVNC stack: Xvfb + window manager + VNC + websockify/noVNC
# Prevent interactive tzdata prompts during docker build
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    tzdata \
    xvfb x11vnc fluxbox novnc websockify ca-certificates \
  && ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime \
  && echo "${TZ}" > /etc/timezone \
  && rm -rf /var/lib/apt/lists/*

# Ensure python deps are present and match the image's browser bundle
RUN python -m pip install --no-cache-dir --upgrade pip \
  && python -m pip install --no-cache-dir playwright==1.44.0

COPY scripts/shuba_browser_session.py /app/shuba_browser_session.py
COPY scripts/entrypoint_novnc.sh /app/entrypoint_novnc.sh

RUN chmod +x /app/entrypoint_novnc.sh

ENV DISPLAY=:99
EXPOSE 7900

ENTRYPOINT ["/app/entrypoint_novnc.sh"]
