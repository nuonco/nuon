VERSION --use-cache-command 0.6

IMPORT github.com/powertoolsdev/shared-configs:main

FROM ghcr.io/powertoolsdev/ci-go-builder

WORKDIR /app

ARG GITHUB_ACTIONS=

ARG EARTHLY_GIT_PROJECT_NAME
ARG GHCR_IMAGE=ghcr.io/$EARTHLY_GIT_PROJECT_NAME
ARG BUF_USER=jonmorehouse
ARG BUF_API_TOKEN=4c51e8481ed34404b7ab6a0c62dc7b2db82757d8f86e4caa853750973e2c5083

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

push-proto:
    FROM bufbuild/buf
    WORKDIR /work
    COPY --dir api/ .
    COPY --dir components/ .
    COPY --dir .git/ .
    RUN echo $BUF_API_TOKEN | buf registry login --username $BUF_USER --token-stdin
    RUN cd api && buf push
    RUN cd components && buf push

lint-proto:
    FROM bufbuild/buf
    WORKDIR /work
    COPY --dir api/ .
    COPY --dir components/ .
    RUN cd api && buf lint
    RUN cd api && buf format -d --exit-code || (printf '%s\n' "Buf Format changes exist in current branch">&2 && exit 1)
    RUN cd components && buf lint
    RUN cd components && buf format -d --exit-code || (printf '%s\n' "Buf Format changes exist in current branch">&2 && exit 1)

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
