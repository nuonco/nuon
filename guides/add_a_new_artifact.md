# Add a new artifact

Artifacts are binaries and images that are automatically published to s3 and ECR.

Our CI is tightly integrated with both `terraform` and `earthly` based on conventions to quickly publish new artifacts without custom infrastructure or CI tooling.

## Artifact ci

We define a parent `artifacts.yml` file in our CI that defines a list of directories that have artifacts [here](https://github.com/powertoolsdev/mono/blob/main/.github/workflows/artifacts.yml#L41).

We use [paths-filter](https://github.com/dorny/paths-filter) for all actions "logic" (ie: picking what to run) in `mono`.

Whenever a change happens in the specified directory, we automatically load the correct terraform outputs for that directory from `infra/artifacts/outputs.tf`.

## Artifact infrastructure

Our artifact tooling allows us to publish to both public and private ECR repos, as well as S3. In order to configure artifacts, an output must be set in `infra/artifacts` with the correct outputs:

```terraform
  "bins/nuonctl" = {
    bucket_prefix = "nuonctl"
    ecr           = module.nuonctl.all
  }
```

During CI, the workflow will automatically look up the correct artifact outputs and use that to pick the correct ECR repo to communicate with.

## Earthly targets

Artifact tooling works by looking for two Earthly targets during CI:

* oci-artifacts - an OCI image that needs to be pushed
* artifacts - artifacts to be written to s3

Both of these targets must be defined, or CI will fail.

CI will automatically upload the artifacts outputted by `artifacts` into `s3://nuon-artifacts/<artifact>/<git-ref>`.

CI will automatically push the OCI artifact from `oci-artifacts` into the correct ECR repository defined in the aforementioned outputs.

## Add a new artifact

Adding a new artifact requires the following:

### Update `infra/artifacts`

We define all ECR repos in `infra/artifacts/ecr.tf`. Based on the type of artifact (public / private), simply copy a block and update it for the correct name.

```terraform
module "nuonctl" {
  source = "../modules/ecr"

  name = "nuonctl"
  tags = {
    artifact      = "sandbox-aws-eks"
    artifact_type = "binary"
  }

  region = local.aws_settings.region
  providers = {
    aws = aws.infra-shared-prod
  }
}
```

From there, update `outputs.tf`, and make sure that your artifact is set as an output in the `artifacts` output:

```terraform
output "artifacts" {
  value = {
    // charts
    "charts/demo" = {
      bucket_prefix = "helm-demo"
      ecr           = module.helm_demo.all
    }
    "charts/temporal" = {
      bucket_prefix = "helm-temporal"
      ecr           = module.helm_temporal.all
    }
...
```

Finally, make sure that the correct github actions IAM role can access the newly created ECR repo:


If it's a public repo, add the ARN into the public role:
```terraform
  // grant permissions for public repos
  statement {
    actions = [
      "ecr-public:BatchCheckLayerAvailability",
      "ecr-public:BatchGetImage",
      "ecr-public:BatchDeleteImage",
      "ecr-public:BatchImportUpstreamImage",
      "ecr-public:CompleteLayerUpload",
      "ecr-public:DescribeImages",
      "ecr-public:DescribeRepositories",
      "ecr-public:GetDownloadUrlForLayer",
      "ecr-public:InitiateLayerUpload",
      "ecr-public:ListImages",
      "ecr-public:PutImage",
      "ecr-public:UploadLayerPart",
    ]
    resources = [
      // helm charts
      module.helm_demo.repository_arn,
      module.helm_temporal.repository_arn,
```

### Update `.github/workflows`

Once your infrastructure is setup, update `.github/workflows`, to set your path as a target. Update [this file](https://github.com/powertoolsdev/mono/blob/main/.github/workflows/artifacts.yml#L38):

```yaml
 - uses: dorny/paths-filter@v2
        id: filter
        with:
          filters: |
            bins/nuonctl:
              - bins/nuonctl/**

            charts/demo:
              - charts/demo/**
            charts/temporal:
              - charts/temporal/**
            charts/waypoint:
              - charts/waypoint/**
```

### Add a directory + earthly targets

Finally, add your code + Earthfile.

At a minimum, your Earthly file must contain an implemention of each of the following targets.

```earthly
code:
    FROM ../../.+deps
    COPY ../../pkg+code/pkg pkg
    COPY --dir . ./bins/$BIN
    WORKDIR bins/$BIN
    RUN go generate ./...

lint:
    FROM +code

test:
    FROM +code
    RUN go test ./...

test-integration:
    RUN echo "noop..."

artifacts:
    FROM +code
    RUN mkdir -p out && \
      go build -o ./out/$BIN .

    SAVE ARTIFACT out AS LOCAL ./out

oci-artifacts:
    COPY +artifacts/out/$BIN /bin/$BIN

    ARG repo=
    ARG image_tag=
    SAVE IMAGE --push $repo:$image_tag
```
