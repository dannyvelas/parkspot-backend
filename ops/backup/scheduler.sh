#!/usr/bin/env bash

# `stopping` is set to 1 on SIGINT 
stopping=0
# holds the PID of the sleep command
sleeper=
# if an interrupt is received, and sleep is running, kill it.
# however, if a backup is running, in that case dont kill it
trap 'stopping=1; [[ -n "$sleeper" ]] && kill "$sleeper" 2>/dev/null' TERM INT

while (( ! stopping )); do
  # now/next are UTC epoch
  now=$(date -u +%s)
  next=$(date -u -d 'today 10:30' +%s)
  # if `next` is in the past, set it to be tomorrow so it's in the future again
  (( next <= now )) && next=$(date -u -d 'tomorrow 10:30' +%s)

  # sleep in the background. save PID into $sleeper
  sleep $(( next - now )) & sleeper=$!

  # on the small chance that an interrupt was received after the sleep was sent to background
  # but BEFORE sleeper was set, that would be an edge-case where the `trap` DOESNT kill the
  # sleep process. so this line will kill it instead
  (( stopping )) && kill "$sleeper" 2>/dev/null

  # wait until sleeper finishes. however, if an interrupt happens the trap would run immediately
  wait "$sleeper"
  sleeper=

  # if an interrupt is received at this point just exit the loop
  (( stopping )) && break

  # run backup in the background. save PID into job
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
