FROM golang:buster
ARG USER
ARG GROUP
ARG UID
ARG GID
ARG DOCKER_GID

ENV DEVROOT=/app
ENV GOPATH=/app/.direnv
ENV GOBIN=/app/.direnv/bin
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
ENV PATH="${PATH}:/app/.direnv/bin:/app/sandbox/bin"

# Run sandbox specific tasks in the container.
RUN echo '#!/bin/bash\n/app/.direnv/bin/tusk -f /app/sandbox.tusk.yml "$@"' > /usr/bin/tusk && \
    chmod +x /usr/bin/tusk

# Set up local user.
RUN groupadd -f -g ${DOCKER_GID} docker \
    && if grep -q ${GROUP} /etc/group; then \
    groupmod -g ${GID} ${GROUP}; \
    else \
    groupadd -f -g ${GID} ${GROUP}; \
    fi \
    && useradd -m -u ${UID} -g ${GID} -G ${DOCKER_GID} ${USER} -s /bin/bash \
    && mkdir -p /app \
    && chown -R ${UID}:${GID} /app

WORKDIR /app

USER ${USER}
SHELL ["/bin/bash", "-c"]

# # Install test runner.
# RUN go get \
#     github.com/agnivade/wasmbrowsertest \
#     && cp $GOPATH/bin/wasmbrowsertest $GOPATH/bin/go_js_wasm_exec
