VERSION --use-cache-command 0.6

IMPORT github.com/powertoolsdev/shared-configs:main

FROM ghcr.io/powertoolsdev/ci-go-builder-docker-compose

WORKDIR /work

ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG CGO_ENABLED=0
ARG GOPRIVATE=github.com/powertoolsdev/*
ARG ETCSSL=/etc/ssl/cert.pem

ARG GITHUB_ACTIONS=

ARG EARTHLY_GIT_PROJECT_NAME
ARG GHCR_IMAGE=ghcr.io/${EARTHLY_GIT_PROJECT_NAME}

CACHE $GOCACHE
CACHE $GOMODCACHE

build:
    DO +DEPS
    RUN go build -o bin/service .
    SAVE ARTIFACT bin/service /service
    SAVE IMAGE --push $GHCR_IMAGE:build

certs:
    FROM alpine:3.16
    # NOTE(jdt): ideally we would wget -a but busybox doesn't support it
    # or use `update-ca-certificate` but it can't handle a single file w/ multiple certs
    RUN wget \
            https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem \
        && cat global-bundle.pem >> "$ETCSSL"
    SAVE ARTIFACT "$ETCSSL" /cert.pem
    SAVE IMAGE --cache-hint

docker:
    FROM alpine:3.16
    ARG EARTHLY_GIT_ORIGIN_URL
    ARG EARTHLY_GIT_SHORT_HASH
    ARG EARTHLY_SOURCE_DATE_EPOCH
    ARG EARTHLY_TARGET_PROJECT_NO_TAG
    ARG cache_tag=$EARTHLY_TARGET_TAG_DOCKER
    ARG EARTHLY_TARGET_TAG_DOCKER

    # These are set to run locally. They _must_ be overridden in CI.
    ARG repo=$EARTHLY_TARGET_PROJECT_NO_TAG
    ARG image_tag=$EARTHLY_TARGET_TAG_DOCKER

    COPY +build/service /bin/service
    COPY +certs/cert.pem "$ETCSSL"
    ENTRYPOINT ["/bin/service"]
    LABEL org.opencontainers.image.created=$EARTHLY_SOURCE_DATE_EPOCH
    LABEL org.opencontainers.image.revision=$EARTHLY_GIT_SHORT_HASH
    LABEL org.opencontainers.image.version=$image_tag
    LABEL org.opencontainers.image.source=$EARTHLY_GIT_ORIGIN_URL
    LABEL org.opencontainers.image.authors=nuon
    LABEL org.opencontainers.image.vendor=nuon
    SAVE IMAGE --push --cache-from=$repo:$cache_tag $repo:$image_tag

helm-bump-and-publish:
    FROM ghcr.io/powertoolsdev/ci-helm-releaser
    ARG chart_version
    ARG repos
    WORKDIR /work
    COPY --dir k8s /work/k8s
    DO shared-configs+HELM_RELEASE \
        --chart_version=$chart_version \
        --repos=$repos
    SAVE IMAGE --push $GHCR_IMAGE:helm

deploy:
    FROM ghcr.io/powertoolsdev/ci-helm-releaser
    WORKDIR /work
    COPY --dir k8s /work/k8s
    ARG env
    ARG cluster_name
    ARG role_arn
    ARG chart_version
    ARG repository
    ARG image_tag
    ARG tfenv=.tfenv
    ARG repos
    DO shared-configs+HELM_SETUP \
        --repos="$repos"

    WORKDIR /work/k8s
    COPY $tfenv .
    RUN while IFS='=' read -r k v; do eval export "$k='$v'"; done < "$tfenv" \
        && envsubst < "values.$env.tmpl" > "values.$env.yaml"

    DO shared-configs+K8S_DEPLOY \
        --env=$env \
        --cluster_name=$cluster_name \
        --role_arn=$role_arn \
        --chart_version=$chart_version \
        --repository=$repository \
        --image_tag=$image_tag
    SAVE IMAGE --push $GHCR_IMAGE:deploy

test:
    DO +DEPS
    COPY docker-compose.yml ./
    COPY wait_for_pg.sh /bin
    WITH DOCKER \
        --compose docker-compose.yml \
        --pull ghcr.io/powertoolsdev/ci-temporalite:latest \
        --pull postgres:13.4-alpine
        RUN  \
            wait_for_pg.sh \
            && go run main.go migrate up\
            && go test -v ./...
    END
    SAVE IMAGE --push $GHCR_IMAGE:test

test-integration:
    DO +DEPS
    COPY docker-compose.yml ./
    COPY wait_for_pg.sh /bin
    WITH DOCKER \
        --compose docker-compose.yml \
        --pull ghcr.io/powertoolsdev/ci-temporalite:latest \
        --pull postgres:13.4-alpine
        RUN \
            wait_for_pg.sh \
            && go run main.go migrate up\
            && go test \
                -tags=integration \
                ./...
    END
    SAVE IMAGE --push $GHCR_IMAGE:test-integration

lint:
    FROM ghcr.io/powertoolsdev/ci-reviewdog
    CACHE $GOCACHE
    CACHE $GOMODCACHE
    WORKDIR /work
    DO +DEPS
    COPY --dir . .
    DO shared-configs+LINT \
        --GITHUB_ACTIONS=$GITHUB_ACTIONS \
        --GOCACHE=$GOCACHE \
        --GOMODCACHE=$GOMODCACHE
    SAVE IMAGE --push $GHCR_IMAGE:lint


################################### UDCs ######################################

DEPS:
    COMMAND
    COPY go.mod go.sum .
    COPY --dir cmd internal migrations .
    COPY *.go .
    DO shared-configs+SETUP_SSH --GITHUB_ACTIONS=$GITHUB_ACTIONS
    RUN --ssh git config --global --add safe.directory "$(pwd)" \
        && go mod download
    DO +GEN

GEN:
  COMMAND
  RUN go generate ./...

################################### LOCAL #####################################

bin:
   ARG BUILD_SIGNATURE=local
   ARG EARTHLY_GIT_SHORT_HASH
   LOCALLY
   RUN unset GOCACHE GOMODCACHE \
        && go build \
           -v \
           -o bin/service \
           .
