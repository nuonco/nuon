VERSION --use-cache-command --use-copy-link 0.6

IMPORT github.com/powertoolsdev/shared-configs:main

FROM ghcr.io/powertoolsdev/ci-go-builder

WORKDIR /app

ARG GITHUB_ACTIONS=
ARG REPO=

ARG EARTHLY_GIT_PROJECT_NAME
ARG GHCR_IMAGE=ghcr.io/$EARTHLY_GIT_PROJECT_NAME

ARG BUF_USER=jonmorehouse
ARG BUF_API_TOKEN=4c51e8481ed34404b7ab6a0c62dc7b2db82757d8f86e4caa853750973e2c5083

proto:
    FROM bufbuild/buf
    WORKDIR /work
    COPY --dir $REPO/ .
    WORKDIR $REPO
    RUN echo $BUF_API_TOKEN | buf registry login --username $BUF_USER --token-stdin

push-proto:
    FROM +proto
    RUN buf push

lint:
    BUILD +lint-standard
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

lint-proto:
    FROM +proto
    DO +BUF_LOGIN
    RUN sh -c "buf lint && buf format -d --exit-code" \
            || sh -c "printf '%s\n' 'Buf Format changes exist in current branch' >&2 && exit 1"

generate-protos:
    WORKDIR /work
    COPY --dir . .

    RUN go install github.com/bufbuild/buf/cmd/buf@v1.12.0 \
        && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
        && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
        && go install github.com/srikrsna/protoc-gen-gotag@v0.6.2

    DO +BUF_LOGIN
    RUN go generate ./...

    SAVE ARTIFACT orgs-api/generated orgs-api/generated
    SAVE ARTIFACT components/generated components/generated
    SAVE ARTIFACT api/generated api/generated
    SAVE ARTIFACT workflows/generated workflows/generated

################################### UDCs ######################################

BUF_LOGIN:
  COMMAND
  RUN echo $BUF_API_TOKEN | buf registry login --username $BUF_USER --token-stdin

################################### LOCAL #####################################
