FROM postgres:17.6

RUN apt-get update && apt-get install -y --no-install-recommends awscli && rm -rf /var/lib/apt/lists/*

COPY backup.sh /usr/local/bin/backup.sh
RUN chmod +x /usr/local/bin/backup.sh

ENTRYPOINT ["/bin/sh", "-c", "while true; do /usr/local/bin/backup.sh; sleep 86400; done"]
