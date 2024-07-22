ARG BASE_IMAGE=influxdb:2.7.7

FROM ${BASE_IMAGE}

RUN rm -r /etc/influxdb2
RUN rm -r /var/lib/influxdb2

# Phase 1: Bootstrap config to get the server to start so we can populate /data/data and /data/config
# RUN ln -s /data /etc/influxdb2
# RUN ln -s /data /var/lib/influxdb2

# Phase 2
RUN ln -s /data/data /etc/influxdb2
RUN ln -s /data/config /var/lib/influxdb2

ENV DOCKER_INFLUXDB_INIT_MODE=setup
ENV DOCKER_INFLUXDB_INIT_ORG=groceries
ENV DOCKER_INFLUXDB_INIT_BUCKET=groceries

# These envars are set in secrets.
# ENV DOCKER_INFLUXDB_INIT_USERNAME=
# ENV DOCKER_INFLUXDB_INIT_PASSWORD=
# ENV DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=

# Use this to override the base image's entrypoint if you need the container to just stay up
# regardless of influx crashing every time it starts.
# ENTRYPOINT ["tail", "-f", "/dev/null"]
