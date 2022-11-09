VERSION --use-cache-command 0.6

IMPORT github.com/powertoolsdev/shared-configs:main

FROM ghcr.io/powertoolsdev/ci-go-builder

WORKDIR /app

ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG CGO_ENABLED=0
ARG GOPRIVATE=github.com/powertoolsdev/*

ARG GITHUB_ACTIONS=

ARG GHCR_IMAGE=ghcr.io/powertoolsdev/workers-instances

build:
    DO +DEPS
    RUN go build -o bin/workers .
    SAVE ARTIFACT bin/workers /workers

# TODO(jdt): do something better and platform specific
iamauthenticator:
    FROM +base
    WORKDIR /tmp
    # iam authenticator download url
    ENV AUTHR_URL='https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v0.5.9/aws-iam-authenticator_0.5.9_linux_amd64'
    RUN curl -Lo aws-iam-authenticator $AUTHR_URL && \
        chmod +x aws-iam-authenticator
    SAVE ARTIFACT aws-iam-authenticator

docker:
    FROM alpine:3.16
    ARG EARTHLY_GIT_ORIGIN_URL
    ARG EARTHLY_GIT_SHORT_HASH
    ARG EARTHLY_SOURCE_DATE_EPOCH
    ARG EARTHLY_TARGET_PROJECT_NO_TAG
    ARG EARTHLY_TARGET_TAG_DOCKER

    # These are set to run locally. They _must_ be overridden in CI.
    ARG repo=$EARTHLY_TARGET_PROJECT_NO_TAG
    ARG cache_tag=$EARTHLY_TARGET_TAG_DOCKER
    ARG image_tag=$EARTHLY_TARGET_TAG_DOCKER

    BUILD +iamauthenticator
    COPY +build/workers /bin/workers
    COPY +iamauthenticator/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
    RUN apk add --update --no-cache \
        git \
        zip
    ENTRYPOINT ["/bin/workers", "all"]
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

test:
    DO +DEPS
    RUN go test ./...

test-integration:
    DO +DEPS
    # This should match our running version of k8s as closely as possible.
    # Use `+get-envtest-versions` to get the list of supported binaries
    # as they don't publish envtest binaries for every patch release.
    ENV K8S_VERSION=1.23.5
    RUN curl \
            -sSLo /tmp/envtest-bins.tar.gz \
            "https://go.kubebuilder.io/test-tools/${K8S_VERSION}/$(go env GOOS)/$(go env GOARCH)" \
        && mkdir -p /usr/local/kubebuilder \
        && tar -C /usr/local/kubebuilder --strip-components=1 -zvxf /tmp/envtest-bins.tar.gz \
        && rm -f /tmp/envtest-bins.tar.gz
    ENV ACK_GINKGO_DEPRECATIONS=1.16.5
    RUN go test -tags=integration ./...
    SAVE IMAGE --push $GHCR_IMAGE:test-integration

lint:
    FROM ghcr.io/powertoolsdev/ci-reviewdog
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
    COPY --dir cmd internal .
    COPY *.go .
    DO shared-configs+SETUP_SSH --GITHUB_ACTIONS=$GITHUB_ACTIONS
    RUN --ssh git config --global --add safe.directory "$(pwd)" \
        && go mod download


################################### LOCAL #####################################
bin:
   ARG BUILD_SIGNATURE=local
   ARG EARTHLY_GIT_SHORT_HASH
   LOCALLY
   RUN unset GOCACHE GOMODCACHE \
        && go build \
           -v \
           -o bin/workers \
           .

get-envtest-versions:
    FROM alpine:3.16
    RUN apk add --update --no-cache curl ca-certificates xmlstarlet
    RUN xmlstarlet select --help
    # This file is actually in GCS but uses same XMLNS as S3
    RUN curl -L "https://go.kubebuilder.io/test-tools" \
        | xmlstarlet \
            select \
            -N x="http://doc.s3.amazonaws.com/2006-03-01" \
            --template \
            --value-of "//x:Key"
