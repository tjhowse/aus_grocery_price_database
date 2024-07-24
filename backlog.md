# Backlog

## Basic

### Devops
* Github ci/cd
* Set up private network between services.
* Monitor disk utilisation

### General
* Add persistence to last-checked time. Store it in the grocery data provider. GetProductsSinceLastCheck(maxcount int).
    * Not sure if this is a great idea. Maybe persist the last-checked time in a main-level DB.
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
* Export grafana config/dashboards/etc to repo. Embed as a part of dockerfile (?)

## Further work

### Other storefronts
* Coles
* Aldi

### Hosting
* Write a docker-compose.yaml for non-fly.io hosting.