FROM postgres:17.6

RUN apt-get update && apt-get install -y --no-install-recommends awscli curl ca-certificates && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL -o /usr/local/bin/supercronic https://github.com/aptible/supercronic/releases/download/v0.2.51/supercronic-linux-amd64 && \
    echo "cc232108374e8e2b537045f2d5eaa1229f619db314c3ba6a14d38b35dc0aa123  /usr/local/bin/supercronic" | sha256sum -c - && \
    chmod +x /usr/local/bin/supercronic

COPY backup.sh /usr/local/bin/backup.sh
RUN chmod +x /usr/local/bin/backup.sh

COPY crontab /etc/crontabs/backup

ENTRYPOINT ["/usr/local/bin/supercronic", "/etc/crontabs/backup"]
