# Australian grocery price database

This service grocery prices from Woolworth's website to an influxdb timeseries database.

In the future it could read from other Australian grocers, based on time, motivation, etc.

**I need help!**

I will set up a bog-standard grafana instance for visualising the data collected. However it would be very nice to have a more tailored frontend. Let me know if you can help out.

Additionally: Hosting this will be fairly cheap, but not free. Consider a cheeky github sponsorship if you reckon this project is worthwhile.

## Architecture

This service consists of three applications:

* aus_grocery_price_database
  * This application written in golang. Reads from grocery store web APIs and streams price data to the timeseries database.
* InfluxDB2
  * A timeseries database. Efficiently stores tagged numerical information, basic data exploration and graphing capacity built-in.
* Grafana
  * A more capable plotting and dashboarding application that can hook into timeseries databases.

Each of these applications runs in a separate docker container. Each container persists some data to a `/data` directory mounted to persistent storage.

Only the Grafana instance is public-facing via the main domain. It has a read-only API token for reading from InfluxDB. AGPD has write-only tokens for writing to InfluxDB.

## Hosting

This is setup for hosting on fly.io. I'm not completely happy with investing effort on hosting infrastructure using a for-profit service, but they sure do make it straightforward. It would not be much more effort to throw together a docker-compose to make it more platform-independent.


