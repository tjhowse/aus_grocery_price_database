ARG BASE_IMAGE=grafana/grafana:10.0.0

FROM ${BASE_IMAGE}

# This needs to be copied into /data/etc/grafana/grafana.ini once the container
# has booted. We can't cook stuff into the mounted /data volume.
COPY grafana.ini /grafana.ini

USER root

# TODO Put these in once we have a favicon/logo
# ## Replace Favicon and Apple Touch
# COPY img/fav32.png /usr/share/grafana/public/img
# COPY img/fav32.png /usr/share/grafana/public/img/apple-touch-icon.png

# ## Replace Logo
# COPY img/logo.svg /usr/share/grafana/public/img/grafana_icon.svg

# Replace grafana here and there with AUSCost
RUN sed -i 's|<title>\[\[.AppTitle\]\]</title>|<title>AUSCost</title>|g' /usr/share/grafana/public/views/index.html
RUN sed -i 's|Loading Grafana|Loading AUSCost|g' /usr/share/grafana/public/views/index.html

RUN find /usr/share/grafana/public/build/ -name *.js \
## Update Title
    -exec sed -i 's|AppTitle="Grafana"|AppTitle="AUSCost"|g' {} \; \
## Update Login Title
    -exec sed -i 's|LoginTitle="Welcome to Grafana"|LoginTitle="Welcome to AUSCost"|g' {} \; \

USER grafana

# Use this to override the base image's entrypoint if you need the container to just stay up
# regardless of the app crashing every time it starts.
# ENTRYPOINT ["tail", "-f", "/dev/null"]