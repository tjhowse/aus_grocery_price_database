# Backlog

## Basic

### Egress
* write numbers to an influxdb database

### Devops
* fly.io config
* Automatic deployment from tags
* Github ci/cd
* Influxdb instance
* Grafana instance

### General
* Add timeouts to all http requests
* Calculate $/g where possible
* Configuration management
* Create DB backup on invalid schema, rather than deleting the old one.
* Tune the max product age to roughly the time it takes for a full DB update.
  * Maybe tune this at runtime? Or don't use a max age at all, instead take the
    X oldest records and update them.

## Further work

### Other storefronts
* Coles
* Aldi