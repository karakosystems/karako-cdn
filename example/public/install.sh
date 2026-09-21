#!/bin/sh
# Served at /install.sh: scripts revalidate every 10 minutes, so
# `curl -fsSL https://cdn.example.com/install.sh | sh` runs this version.
set -eu
echo "installing example v1"
