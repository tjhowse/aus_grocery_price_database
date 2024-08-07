# Backlog

## Basic

### Frontend
* Make sure plotlines are accessible from a colour perception perspective

### Optimisations
* Use transactions for all multi-part SQL operations
* Smear updated times a bit to flatten out request rate peaks.

### Devops
* Set up private network between services.
* Set up health check for AGPD container. Touch /tmp/heartbeat or something.
* Report delta between actual product max age and target.
* Periodic backups of timeseries DB to S3/other block storage. Possibly just clone fly.io volume?

### General
* Extract barcode from category page along with the stockcode, E.G. "Stockcode":134034,"Barcode":"0263151000002"
    * See internal/woolworths/junk/product_info_from_category_page.json
* Also consider extracting ALL product information from the category page. It would simplify things hugely.
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

### Cleanup
* I'm sure there are cases in which we're unintentionally shadowing error variables.
    Run `go vet` with a shadowing thingo to find them. I tried the obvious and it didn't work.
* Patch grafana in the dockerfile (???) to change the AppTitle and icon for previews
    https://github.com/VolkovLabs/volkovlabs-app/blob/main/Dockerfile

## Further work

### Other storefronts
* Coles
* Aldi

### Hosting
* Write a docker-compose.yaml for non-fly.io hosting.

### Features
* Scan a barcode to see the price history of that item.