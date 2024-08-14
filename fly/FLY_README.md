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


## Influxdb data problems

In the first few releases timeseries data was being fed into influxdb that contained leading or trailing spaces on the `name` field. This, in addition to various formatting whoopsies in the raw data from Woolworth's individual-product-info pages (such as double-spaces before `Each` and suchlike), lead to the same product having multiple values for the name field in the data.

### First pass

I tried using influxdb's built-in backup stuff before I started but it requires a root account token that, as far as I can tell, never existed. So I first cloned the fly volume as a backup, running on my local machine.

    fly volumes fork vol_r7l95jwqy0j8kp94

Where the ID is that of the live /data volume.

Then I created a new bucket inside the vm via `fly ssh console`

    influx bucket create --name delme

Then select all data into the new bucket, one week at a time to stay under our modest RAM budget.

    influx query 'from(bucket:"groceries") |> range(start:-4w, stop: -3w) |> to(bucket:"delme")' > /dev/null
    influx query 'from(bucket:"groceries") |> range(start:-3w, stop: -2w) |> to(bucket:"delme")' > /dev/null
    influx query 'from(bucket:"groceries") |> range(start:-2w, stop: -1w) |> to(bucket:"delme")' > /dev/null
    influx query 'from(bucket:"groceries") |> range(start:-1w) |> to(bucket:"delme")' > /dev/null

Then I got the `delme` bucket ID with

    influx bucket list

I dumped the data out in influxdb's `line protocol` format:

    influxd inspect export-lp --bucket-id f1e7c005e7329b59 --engine-path /data/config/engine --output-path /auscost_backup_2024-08-14.lp

E.G.

    product,department=Baby,id=woolworths_sku_907296,name=A2\ Milk\ Gentle\ Gold\ Stage\ 4\ Formula\ 3\ Years+\ ,store=Woolworths cents=3000i 1723058406827331098
    product,department=Baby,name=Zazu\ \ \ 1EA,store=Woolworths grams=0i 1722993898819474413
    product,department=Baby,name=Zuru\ Coco\ Surprise\ \ Each,store=Woolworths cents=800i 1722994174828510260
    product,department=Baby,name=Zuru\ Coco\ Surprise\ \ Each,store=Woolworths grams=0i 1722994174828510260

More example data is available in `influxdb/spaces_in_name_example_data.lp`. The example data was found with the regexes `,name=\\ .*,` and `,name=.*\\ ,`

This data was sanitised with the following `sed` commands:

    # Remove duplicate spaces.
    sed -E 's/(\\ )+/\\ /g' spaces_in_name_example_data.lp > no_duplicate_spaces.lp
    # Remove leading and trailing spaces from the name field.
    sed -E 's/,name=(\\ )*(.*?)(\\ )*,/,name=\2,/g' no_duplicate_spaces.lp > no_leading_spaces.lp
    # At least that's the theory, but sed doesn't respect non-greedy quantifiers (?!?) so the second group matches the trailing `\ `.
    # I feel betrayed. Let's bodge it up.
    sed -E 's/(\S)\\ ,/\1,/g' no_leading_spaces.lp > no_trailing_spaces.lp

But chained together:

    cat spaces_in_name_example_data.lp | sed -E 's/(\\ )+/\\ /g' | sed -E 's/,name=(\\ )*(.*?)(\\ )*,/,name=\2,/g' | sed -E 's/(\S)\\ ,/\1,/g' > cleaned.lp

P.S. I prefer useless cats.

Then delete the temporary bucket and re-create it, and another one for testing:

    influx bucket delete --name delme
    influx bucket create --name delme
    influx bucket create --name temporary

...and write the cleaned data into it for inspection.

    influx write -b delme -f cleaned.lp
    influx query 'from(bucket:"delme") |> range(start:-4w) |> to(bucket:"temporary")' > /dev/null
    influx delete --bucket temporary --start '2009-01-02T23:00:00Z' --stop '2029-01-02T23:00:00Z'
    influx delete --bucket groceries --start '2009-01-02T23:00:00Z' --stop '2029-01-02T23:00:00Z'
    influx query 'from(bucket:"delme") |> range(start:-4w) |> to(bucket:"groceries")' > /dev/null
    influx bucket delete --name delme
    influx bucket delete --name temporary

### Better option for next time

Local:
    # Take a snapshot of the backing volume
    fly volumes fork vol_r7l95jwqy0j8kp94

VM:
    # Get bucket ID
    influx bucket list --name groceries
    # Dump to lp file
    influxd inspect export-lp --bucket-id 01c1ca212aa54ac6 --engine-path /data/config/engine --output-path /2024-08-14_auscost_backup.lp
    # Clean the data
    cat /2024-08-14_auscost_backup.lp | sed -E 's/(\\ )+/\\ /g' | sed -E 's/,name=(\\ )*(.*?)(\\ )*,/,name=\2,/g' | sed -E 's/(\S)\\ ,/\1,/g' > /2024-08-14_auscost_backup_cleaned.lp
    # Delete the contents of the bucket
    influx delete --bucket groceries --start '2009-01-02T23:00:00Z' --stop '2029-01-02T23:00:00Z'
    # Reimport the cleaned data from the lp file.
    influx write -b groceries -f 2024-08-14_auscost_backup_cleaned.lp
