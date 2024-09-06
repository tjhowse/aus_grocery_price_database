# Backlog

## Basic

### Frontend
* Make sure plotlines are accessible from a colour perception perspective
* Light mode

### Optimisations
* Use transactions for all multi-part SQL operations
* Smear updated times a bit to flatten out request rate peaks.
* Cache coles API version to DB.

### Devops
* Set up private network between services.
* Set up health check for AGPD container. Touch /tmp/heartbeat or something.
* Service health monitoring and alerting in general
* Periodic backups of timeseries DB to S3/other block storage. Possibly just clone fly.io volume?

### General
* Add persistence to last-checked time. Store it in the grocery data provider. GetProductsSinceLastCheck(maxcount int).
    * Not sure if this is a great idea. Maybe persist the last-checked time in a main-level DB.
* Calculate $/g where possible
* Export grafana config/dashboards/etc to repo. Embed as a part of dockerfile (?)
* Make sure we're `defer rows.Close()` everywhere we need to.

### Cleanup
* Create a logo and favicon for grafana. Uncomment the stuff in fly/applications/grafana/Dockerfile and deploy

## Further work

### Other storefronts
* Aldi

### Hosting
* Write a docker-compose.yaml for non-fly.io hosting.

### Features
* Scan a barcode to see the price history of that item.
