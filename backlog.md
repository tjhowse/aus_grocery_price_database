# Backlog

## Basic

### Devops
* Automatic deployment from tags
* Github ci/cd
* Grafana instance
* Register domain
* Set up private network between services.
* Add some performance telemetry to report to influxdb.
* Try to transfer auscost.com.au to cloudflare.
* Fix domain name to map to the grafana service rather than redirecting the domain name.

### General
* Add timeouts to all http requests
* Calculate $/g where possible
* Create DB backup on invalid schema, rather than deleting the old one.
* Tune the max product age to roughly the time it takes for a full DB update.
    * Maybe tune this at runtime? Or don't use a max age at all, instead take the
        X oldest records and update them.
    * Would it make sense to scale the number of workers to hit a target full-db
        update interva?
    * Perhaps report the time step when updating a product. Use this data to
        tune the number of runners so we hit the target update rate. Care
        must be taken to handle the initial-startup state, where everything
        might appear very stale and we scale too many workers. Perhaps just
        set a 20-ish cap on the number of workers.
* Add persistence to last-checked time. Store it in the grocery data provider. GetProductsSinceLastCheck(maxcount int).
* Set up a public frontpage to grafana that works without logging in and shows a dashboard.
* Export grafana config/dashboards/etc to repo. Embed as a part of dockerfile (?)

## Further work

### Other storefronts
* Coles
* Aldi

### Hosting
* Write a docker-compose.yaml for non-fly.io hosting.