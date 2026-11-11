#!/usr/bin/env bash

stopping=0
sleeper=
trap 'stopping=1; [[ -n "$sleeper" ]] && kill "$sleeper" 2>/dev/null' TERM INT

while (( ! stopping )); do
  now=$(date -u +%s)
  next=$(date -u -d 'today 03:00' +%s)
  (( next <= now )) && next=$(date -u -d 'tomorrow 03:00' +%s)

  sleep $(( next - now )) & sleeper=$!
  (( stopping )) && kill "$sleeper" 2>/dev/null
  wait "$sleeper"
  sleeper=
  (( stopping )) && break

  /usr/local/bin/backup.sh & job=$!
  while true; do
    wait "$job"
    status=$?
    if (( status > 128 )) && kill -0 "$job" 2>/dev/null; then
      continue
    fi
    break
  done

  if (( status != 0 )); then
    echo "backup.sh exited with status $status" >&2
  fi
done
