FROM alpine:3
ARG PLUGIN_MODULE=github.com/DoVietHoang1712/demo
ARG PLUGIN_GIT_REPO=https://github.com/DoVietHoang1712/demo.git
ARG PLUGIN_GIT_BRANCH=master
RUN apk add --update git && \
    git clone ${PLUGIN_GIT_REPO} /plugins-local/src/${PLUGIN_MODULE} \
      --depth 1 --single-branch --branch ${PLUGIN_GIT_BRANCH}

FROM traefik:v2.5
COPY --from=0 /plugins-local /plugins-local