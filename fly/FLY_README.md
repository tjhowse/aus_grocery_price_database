# Fly.io

## Notes

https://fly.io/docs/laravel/advanced-guides/multiple-applications/
https://hub.docker.com/_/influxdb

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
    Create Dockerfile with "phase 1" uncommented, and "phase 2" commented out.
    Edit fly.toml to use [Build] to build the image rather than using the raw influxdb one.
    fly deploy
    fly ssh console
    mkdir /data/config
    mkdir /data/data
    Comment out phase 1, uncomment phase 2.
    There was some problem with the directories in /data/data and /data/config being root-owned or something? It wouldn't start properly.
    I added `ENTRYPOINT ["tail", "-f", "/dev/null"]` so the container would start and I could manually SSH in and `chown -R influxdb:influxdb /data/*`