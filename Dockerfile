FROM alpine:3
ARG PLUGIN_MODULE=github.com/DoVietHoang1712/demo
WORKDIR /plugins-local/src/${PLUGIN_MODULE}
COPY . /plugins-local/src/${PLUGIN_MODULE}

FROM traefik:v2.5
COPY --from=0 /plugins-local /plugins-local