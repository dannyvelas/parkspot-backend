#!/usr/bin/env bash

# `stopping` is set to 1 on SIGINT+SIGTERM
stopping=0
# holds the PID of the sleep command
sleeper=
# if an interrupt is received, and sleep is running, kill it.
# however, if a backup is running, in that case dont kill it.
# there is a small chance that this trap runs once the sleep
# process has already exited but before $sleeper gets unset.
# this is okay. 2>/dev/null discards the error
trap 'stopping=1; [[ -n "$sleeper" ]] && kill "$sleeper" 2>/dev/null' TERM INT

while (( ! stopping )); do
  # now/next are epoch seconds. 10.30 is UTC
  now=$(date -u +%s)
  next=$(date -u -d 'today 10:30' +%s)
  # if `next` is in the past, set it to be tomorrow so it's in the future again
  (( next <= now )) && next=$(date -u -d 'tomorrow 10:30' +%s)

  # sleep in the background. save PID into $sleeper
  sleep $(( next - now )) & sleeper=$!

  # if an interrupt was received anytime since the beginning of this while loop
  # but BEFORE $sleeper was set, that would be an edge-case where the `trap`
  # DOESNT kill the sleep process. so this line will kill it instead
  (( stopping )) && kill "$sleeper" 2>/dev/null

  # wait until sleeper finishes. if an interrupt happens, bash would interrupt
  # the wait so that the trap can run. 
  wait "$sleeper"
  sleeper=

  # if a signal interrupted wait,
  # or if sleep was killed before the wait started (making wait return immediately)
  # or sleep finished normally but a signal came before the line below executes, then just break
  (( stopping )) && break

  # if we're here, it means that the sleep is done. it is now time to run the backup
  # run backup in the background. save PID into job
  /usr/local/bin/backup.sh & job=$!
  while true; do
    wait "$job"
    status=$?
    # if wait returned and status > 128, it's possible that either `backup.sh` was killed via signal
    # or wait itself was interrupted via signal (and backup.sh is still running in the background)
    # if we're in the first case we want to break. otherwise we want to continue waiting
    # kill -0 just checks if the backup process is still running. if yes, continue, otherwise break
    if (( status > 128 )) && kill -0 "$job" 2>/dev/null; then
      continue
    fi
    break
  done

  # if the while loop above broke because backup.sh exited with an error, or was killed by a signal
  # log error
  if (( status != 0 )); then
    echo "backup.sh exited with status $status" >&2
  fi
done
