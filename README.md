# Australian grocery price database

https://auscost.com.au

This is an open database of grocery prices in Australia. Its goal is to track long-term price trends to help make good purchasing decisions and hold grocery stores to account for price increases.

The service reads grocery prices from Woolworths' and Coles' websites to an influxdb timeseries database.

In the future it could read from other Australian grocers, based on time, motivation, etc.

**I need help!**

I have set up a grafana instance for visualising the data collected. However it would be very nice to have a more tailored frontend. Let me know if you can help out. See `Frontend scope` below.

Additionally: Hosting this will be fairly cheap, but not free. Consider a cheeky github sponsorship if you reckon this project is worthwhile.

## Similar things

The excellently-named [Heisse Preise](https://heisse-preise.io/) does what I'd like Auscost to do eventually. It's [open source](https://github.com/badlogic/heissepreise) and there's a good writeup on [Wired](https://www.wired.com/story/heisse-preise-food-prices/). I think there's much to be learned from its UI.

## Architecture

This service consists of three applications:

* aus_grocery_price_database
  * This application written in golang. Reads from grocery store web APIs and streams price data to the timeseries database.
* InfluxDB2
  * A timeseries database. Efficiently stores tagged numerical information, basic data exploration and graphing capacity built-in.
* Grafana
  * A more capable plotting and dashboarding application that can hook into timeseries databases.

Each of these applications runs in a separate docker container. Each container persists some data to a `/data` directory mounted to persistent storage.

Only the Grafana instance is public-facing via the [main domain](https://auscost.com.au). It has a read-only API token for reading from InfluxDB. AGPD has write-only tokens for writing to InfluxDB.

## Hosting

This is setup for hosting on fly.io. I'm not completely happy with investing effort on hosting infrastructure using a for-profit service, but they sure do make it straightforward. It would be easy to throw together a docker-compose to make it more platform-independent.

## Frontend scope

The current grafana frontend is a stopgap. Ideally it would be replaced by a bespoke frontend. Grafana could still be used for generating plots, under the hood.

### Core Goals

The primary goal of this project is to shift the balance of power in favour of consumers by presenting pricing information on groceries. Use cases include:

* Forecasting low prices based on periodicity
  * E.G. Navel oranges from Woolworths oscillate in price on a 2-week period
* Compare prices between stores
* Track long-term pricing trends
* Not being tricked by false "sales" where the sale price isn't any cheaper than the long-term trends

Another goal of this project is a low budget. This means minimal hosting and maintenance requirements. It needs to be simple and stable. Minimal external dependencies, both in terms of internal software stack and external services. Updating the current stack involves bumping two container version numbers and a few invocations of `fly deploy`.

### Further thoughts

A significant frontend challenge is product differentiation. The data scraped from the grocer's storefronts varies from store to store. There is always a SKU ID, product name, department ID/name and price. The product name is quite dirty, repetitive and awkward. E.G. "40% Salt Reduced", or "Alva Baby Starry Sky Print Reusable Cloth Nappy", "Alva Baby Starry Night Print Reusable Cloth Nappy".

Woolworths provides a barcode number, coles does not. It would be great for a user to be able to scan a barcode on their phone and pull up the price history of that item.