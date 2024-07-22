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

# TODO this stuff.
ENV DOCKER_INFLUXDB_INIT_MODE=setup
ENV DOCKER_INFLUXDB_INIT_USERNAME=blerp
ENV DOCKER_INFLUXDB_INIT_PASSWORD=pp75D7eG83kU9NhbsayA
ENV DOCKER_INFLUXDB_INIT_ORG=groceries
ENV DOCKER_INFLUXDB_INIT_BUCKET=groceries
ENV DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=yf67EDurejjBCt2jBafV

# ENTRYPOINT ["tail", "-f", "/dev/null"]

# ENTRYPOINT ["/usr/bin/sleep 100000"]
# ENTRYPOINT ["/usr/bin/entrypoint.sh"]

# CMD ["serve"]