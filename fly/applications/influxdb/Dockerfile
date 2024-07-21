ARG BASE_IMAGE=influxdb:2.7.7

FROM ${BASE_IMAGE}

RUN ln -s /data/config /etc/influxdb2

RUN ln -s /data /var/lib/influxdb2

# ENV INFLUX_CONFIGS_PATH /data/config
# TODO this stuff.
ENV DOCKER_INFLUXDB_INIT_MODE setup
# ENV DOCKER_INFLUXDB_INIT_USERNAME=blerp
# ENV DOCKER_INFLUXDB_INIT_PASSWORD=blorp
# ENV DOCKER_INFLUXDB_INIT_ORG=groceries
# ENV DOCKER_INFLUXDB_INIT_BUCKET=groceries

ENTRYPOINT ["/usr/bin/entrypoint.sh"]

CMD ["serve"]