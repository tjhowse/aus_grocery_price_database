# Australian grocery price database

This service grocery prices from Woolworth's website to an influxdb timeseries database.

In the future it could read from other Australian grocers, based on time, motivation, etc.

**I need help!**

I will set up a bog-standard grafana instance for visualising the data collected. However it would be very nice to have a more tailored frontend. Let me know if you can help out.

Additionally: Hosting this will be fairly cheap, but not free. Consider a cheeky github sponsorship if you reckon this project is worthwhile.

## Hosting

This is setup for hosting on fly.io. I'm not completely happy with investing effort on hosting infrastructure using a for-profit service, but they sure do make it straightforward. It would not be much more effort to throw together a docker-compose to make it more platform-independent.

### Setup process

Here's the process I went through for setting up the integration. **This is not a guide on how to deploy this app**, this is a record for my later reference.

    fly launch
    # All default answers to the CLI interface questions
    # Change the memory allocation in the fly.toml to memory_mb = 256
    # Shut down the VMs after creation. Next time use `fly launch --no-deploy`
    fly scale count 0
    fly volumes create
    # Ignore warning about volume pinning
    # Add the [mounts] block to fly.toml.
