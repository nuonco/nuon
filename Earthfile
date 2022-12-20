VERSION --use-cache-command --use-copy-link 0.6

IMPORT github.com/powertoolsdev/shared-configs:main

FROM ghcr.io/powertoolsdev/ci-go-builder

WORKDIR /app

ARG GITHUB_ACTIONS=

ARG EARTHLY_GIT_PROJECT_NAME
ARG GHCR_IMAGE=ghcr.io/$EARTHLY_GIT_PROJECT_NAME

ARG BUF_USER=jonmorehouse
ARG BUF_API_TOKEN=4c51e8481ed34404b7ab6a0c62dc7b2db82757d8f86e4caa853750973e2c5083

lint:
    # NOTE(jm): enable linting once we can disable the golinter for a specific project
    # BUILD +lint-standard
    BUILD +lint-proto

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
    COPY --dir workflows/ .
    RUN echo $BUF_API_TOKEN | buf registry login --username $BUF_USER --token-stdin
    DO +PUSH --dir=api
    DO +PUSH --dir=components
    DO +PUSH --dir=workflows

gen-proto:
    FROM bufbuild/buf
    WORKDIR /work
    DO +GEN --dir=api
    DO +GEN --dir=components
    DO +GEN --dir=workflows

lint-proto:
    FROM bufbuild/buf
    WORKDIR /work
    DO +LINT --dir=api
    DO +LINT --dir=components
    DO +LINT --dir=workflows

# NOTE(jdt): this still doesn't work. we should avoid cloning...
breaking-proto:
    FROM bufbuild/buf
    RUN mkdir -p /root/.ssh/
    DO shared-configs+SETUP_SSH
    WORKDIR /work
    DO +BREAKING --dir=api

################################### UDCs ######################################

PUSH:
    COMMAND
    ARG dir=./
    ARG oldworkdir=$(pwd)
    WORKDIR $dir
    RUN buf push
    WORKDIR $oldworkdir

GEN:
    COMMAND
    ARG dir=./
    COPY --dir $dir .
    ARG oldworkdir=$(pwd)
    WORKDIR $dir
    RUN \
        sh -c "buf generate"
    WORKDIR $oldworkdir


LINT:
    COMMAND
    ARG dir=./
    COPY --dir $dir .
    ARG oldworkdir=$(pwd)
    WORKDIR $dir
    RUN \
        sh -c "buf lint && buf format -d --exit-code" \
            || sh -c "printf '%s\n' 'Buf Format changes exist in current branch' >&2 && exit 1"
    WORKDIR $oldworkdir

# NOTE(jdt): this still doesn't work. we should avoid cloning...
BREAKING:
    COMMAND
    ARG dir=./
    COPY --dir $dir .
    ARG oldworkdir=$(pwd)
    WORKDIR $dir
    RUN \
        git clone git@github.com:powertoolsdev/protos ./old \
        && buf breaking --against "./old"
    WORKDIR $oldworkdir


################################### LOCAL #####################################
