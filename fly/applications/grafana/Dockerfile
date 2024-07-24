ARG BASE_IMAGE=grafana/grafana:10.0.0

FROM ${BASE_IMAGE}

# This needs to be copied into /data/etc/grafana/grafana.ini once the container
# has booted. We can't cook stuff into the mounted /data volume.
COPY grafana.ini /grafana.ini

# Use this to override the base image's entrypoint if you need the container to just stay up
# regardless of the app crashing every time it starts.
# ENTRYPOINT ["tail", "-f", "/dev/null"]