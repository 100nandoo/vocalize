#!/usr/bin/env bash
set -e

if [ -n "$INTI_CONFIG_DIR" ]; then
  dir="$INTI_CONFIG_DIR"
elif [ "$(uname)" = "Darwin" ]; then
  dir="$HOME/Library/Application Support/inti"
else
  dir="${XDG_CONFIG_HOME:-$HOME/.config}/inti"
fi

mkdir -p "$dir"

if [ "$(uname)" = "Darwin" ]; then
  open "$dir"
elif command -v xdg-open &>/dev/null; then
  xdg-open "$dir"
else
  echo "$dir"
fi
