# Switch's dashboard, mirrored to our Docker Hub by the release so Railway runs a version we chose,
# and built from here by docker-compose.yml so local runs the same.
# Pinned: bump this line on purpose, together with switch's and agent-controller's
# CONTROLLER_VERSION; do not float to latest.
FROM ghcr.io/sandbox-quantum/gateway:0.31.0

# Reach switch by name, looked up again on every redeploy (see the template). The image's entrypoint
# renders templates/ into conf.d/, replacing the upstream default.conf.
ENV NGINX_ENTRYPOINT_LOCAL_RESOLVERS=1 \
    SWITCH_UPSTREAM=switch.railway.internal:8000
COPY docker/gateway/default.conf.template /etc/nginx/templates/default.conf.template
