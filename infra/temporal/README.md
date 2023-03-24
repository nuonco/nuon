# infra-temporal
terraform module for managing our temporal installations

## Upgrades

1. Clone temporalio/helm-charts
    ```
    git clone git@github.com:temporalio/helm-charts.git
    ```
1. Checkout the tag that you're upgrading to e.g.
    ```
    git co v1.19.1
    ```
1. Login to ECR helm registry
    ```
    aws ecr get-login-password --profile infra-shared-prod.NuonPowerUser \
        | helm registry login \
            --username AWS \
            --password-stdin 431927561584.dkr.ecr.us-west-2.amazonaws.com
    ```
1. Run the `publish_temporal.sh` script from this repo
1. Update `helm.tf`
    1. `locals.temporal.image_tag` to the correct image version
    1. `helm_release.temporal.version` to the helm chart version
1. If the release notes indicate that a schema update is necessary:
    1. Use `temporal_sql_tools.sh` script to launch a container with the
    necessary tools
        1. All of the necessary env vars are set to connect to postgres
        1. Use the [upgrade documentation](https://docs.temporal.io/cluster-deployment-guide#postgresql)
        to do the migration against each database as needed. e.g. to upgrade the
        visibility database
        ```
        temporal-sql-tool --db temporal_visibility update-schema --schema-dir ./schema/postgresql/v96/visibility/versioned
        ```
        1. Don't forget to change the `--db` as well as the `--schema-dir` as
        they are different for the default and visibility DBs.
