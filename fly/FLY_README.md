# Fly.io

## Notes

https://fly.io/docs/laravel/advanced-guides/multiple-applications/
https://hub.docker.com/_/influxdb

## Secrets

Set the following secrets with `fly secrets set KEY=VALUE KEY=VALUE`. It's best to set them all in one go, because it restarts the app every time. Make sure the password and token are fairly long, otherwise influxdb will go into a boot loop about it.

* DOCKER_INFLUXDB_INIT_USERNAME
* DOCKER_INFLUXDB_INIT_PASSWORD
* DOCKER_INFLUXDB_INIT_ADMIN_TOKEN

## Setup process

Here's the process I went through for setting up the integration. **This is not a guide on how to deploy this app**, this is a record for my later reference.

    fly launch
    # All default answers to the CLI interface questions
    # Change the memory allocation in the fly.toml to memory_mb = 256
    # Shut down the VMs after creation. Next time use `fly launch --no-deploy`
    fly scale count 0
    fly volumes create local_database
    # Ignore warning about volume pinning
    # Add the [mounts] block to fly.toml.
    cd fly/applications/influxdb
    fly launch --image influxdb:2.7.7 --no-deploy --name aus-grocery-price-database-influxdb
    fly volumes create influxdb_database
    # Add the [mounts] block to the influxdb fly.toml
    Create Dockerfile with `rm` and `ln` lines commented out.
    Edit fly.toml to use [Build] to build the image rather than using the raw influxdb one.
    Set the secrets.
    fly deploy
    fly ssh console
    mkdir /data/config
    mkdir /data/data
    Uncomment the `rm` and `ln` lines to map influxdb into the volume.
    There was some problem with the directories in /data/data and /data/config being root-owned or something? It wouldn't start properly.
    I added `ENTRYPOINT ["tail", "-f", "/dev/null"]` so the container would start and I could manually SSH in and `chown -R influxdb:influxdb /data/*`
    This probably wouldn't happen if I had everything set up properly the first time around.
    fly launch --image grafana/grafana:10.0.0 --no-deploy --name aus-grocery-price-database-grafana
    Similar problems with permissions on /data. It looks like the permissions on fly.io volumes are set based on the launch user,
    and don't handle the services dropping out of root after init. Tweak the dockerfile to get a console and set permissions in the volume.
    fly ssh console
    mkdir -p /data/usr/share/grafana
    mkdir -p /data/var/log/grafana
    mkdir -p /data/var/lib/grafana/plugins
    mkdir -p /data/etc/grafana/provisioning
    # This next line may copy from /grafana.ini, rather than using the default shipped with grafana. Watch this space.
    cp /grafana.ini /data/etc/grafana/grafana.ini
    chown -R grafana /data
    Generate a read-only token in influxdb, configure a connector in grafana.
    Use "Authorization" cookie, contents "Token <influx token here>".

