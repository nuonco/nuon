VERSION --use-cache-command 0.6

IMPORT github.com/powertoolsdev/shared-configs:main

FROM ghcr.io/powertoolsdev/ci-go-builder

WORKDIR /app

ARG GITHUB_ACTIONS=

ARG EARTHLY_GIT_PROJECT_NAME
ARG GHCR_IMAGE=ghcr.io/$EARTHLY_GIT_PROJECT_NAME

lint:
    BUILD +lint-standard
    BUILD +lint-proto
    # BUILD +breaking-proto

lint-standard:
    FROM ghcr.io/powertoolsdev/ci-reviewdog
    WORKDIR /work
    COPY --dir . .
    DO shared-configs+LINT \
        --GITHUB_ACTIONS=$GITHUB_ACTIONS \
        --GOCACHE=$GOCACHE \
        --GOMODCACHE=$GOMODCACHE
    SAVE IMAGE --push $GHCR_IMAGE:lint

lint-proto:
    FROM bufbuild/buf
    WORKDIR /work
    COPY --dir protos/ .
    COPY buf.gen.yaml .
    COPY buf.work.yaml .
    RUN buf lint

breaking-proto:
    FROM bufbuild/buf
    WORKDIR /work
    GIT CLONE https://github.com/powertoolsdev/protos ./old
    COPY --dir protos/ .
    COPY buf.gen.yaml .
    COPY buf.work.yaml .
    RUN buf breaking --against "./old"
################################### UDCs ######################################
################################### LOCAL #####################################
