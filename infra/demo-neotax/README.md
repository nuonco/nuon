# infra
GCP cloud infrastructure defined in Terraform code.

## Initial setup 

1. Install Terraform on Mac
   ```bash
   $ brew tap hashicorp/tap
   $ brew install hashicorp/tap/terraform
   ```

1. Install Terraform auto-complete (useful in CLI)
   ```bash
   $ terraform -install-autocomplete
   ```

## Useful scripts

```bash
# initialize Terraform or local modules
$ terraform init

# fix lint errors
$ terraform fmt -recursive

# plan and apply changes (will confirm again in prompt)
$ terraform apply
```

## Development cycle

The development cycle looks like the following:

1. Clarify the requirements (e.g. need to add one Cloud Run instance with 4GB memory, across all environments dev/staging/prod).

1. Write Terraform code (e.g. define a new module, update fullstack code in `./modules/fullstack/main.tf`, or something else).

1. Run `terraform plan` locally, see if the output matches your expectation, and refine your code otherwise.

1. Once the code change looks good, create a Github Pull Request (PR). On every commit pushed to the PR, **Terraform Cloud** will run `terraform plan` remotely and post its output as PR comments, so you and PR reviewers can take a look.

1. Merge the PR after reviewers approved it. **Terraform Cloud** will kick off a `terraform apply`.

1. Visit the **Terraform Cloud** dashboard and make sure the `terraform apply` completed without any problem.

## Code structure

The top-level `main.tf` is the entrypoint, which imports other modules.

``` bash
├── README.md
├── main.tf # top-level main.tf is the entrypoint of the entire Terraform code
└── modules # local modules to DRY up the code
    └── fullstack # defines a fullstack in a single environment (one of dev/staging/prod)
        ├── main.tf
        ├── outputs.tf
        ├── variables.tf
        └── modules
            └── gaap-cap-workflow # defines everything needed for the "gaap cap workflow" (e.g. Cloud Tasks)
               ├── main.tf
               ├── outputs.tf
               └── variables.tf
```

The overall structure can be summarized as this:

* Top level: `main.tf`

  * imports a `modules/fullstack` for `dev`
  * imports a `modules/fullstack` for `staging`
  * imports a `modules/fullstack` for `prod`

* 2nd level: `modules/fullstack/main.tf` (for a single environment: one of dev/staging/prod)

  * imports a `modules/vpc` (e.g. the VPC)
  * imports a `modules/app_database` (e.g. some Postgres database)
  * imports a `modules/app_service` (e.g. some Cloud Run instances)
  * imports a `modules/pa_service` (e.g. some Cloud Run instances)
  * imports a `modules/load_balancer` (e.g. for switching between 2 zones)

* lowest level: `modules/fullstack/modules/vpc/main.tf` (a particular piece of modular infra code)

  * defines a `resource "google_compute_network" ...` (barebone Terraform resources)
  * defines more `resource` in the GCP world (ref: [API docs](https://registry.terraform.io/providers/hashicorp/google/latest/docs))


In each module (under the `./modules` folder), it's always the same 3 files that defines the input (`variables.tf`), the infra code (`main.tf`), and the output (`outputs.tf`).

```bash
─ <module_name> # name the folder to represent the module
  ├── README.md # (optional) describe the module usage
  ├── variables.tf # input to this module
  ├── main.tf # infra defined by this module
  ├── outputs.tf # output from this module (e.g. VPC ID that will be used by other modules)
  └── modules/ # (optional) sub modules folder only used by this module
```

## FAQ

### How to authenticate CLI to access Terraform Cloud?

```bash
# allows CLI to access Terraform Cloud
$ terraform login

# now you can plan/apply locally (runs on Terraform Cloud and streamed locally)
$ terraform plan
```

### How to grant GCP access to Terraform Cloud?

This is a one-time setup required for Terraform Cloud to be able to access GCP, necessary because our Terraform run are performed on Terraform Cloud.

1. Create and download a service account key. You almost never need to do this as we already created and put the file in **1password**.

    1. Visit the Google Cloud dashboard.

    1. Download the JSON key file for the `terraform` service account.

    1. Assuming the downloaded file is called `terraform-service-account.json`

1. Add the GCP credential to Terraform.

    1. Visit the Terraform Cloud dashboard using the shared login account stored in **1password**.

    1. Create an environment variable called `GOOGLE_CREDENTIALS` in the Terraform Cloud workspace.

    1. Remove the newline characters from your JSON key file and then paste the credentials into the environment variable value field. You can use the following command to do so:

        ```bash
        $ cat terraform-service-account.json | tr -s '\n' ' '
        ```

    1. Mark the variable as `Sensitive` and click `Save`.

### How is Terraform Cloud integrated with Github Pull Requests?

We created Github Actions for this integration, with the following steps:

1. In Terraform Cloud, create an **API token** named `GitHub Actions`, never expiring. Copy the API token value.

1. In Github repo `Settings > Secrets and variables > Actions`, create a `New repository secret` named `TF_API_TOKEN`, pasted the API token value.

1. Github Actions are defined under the `./.github/workflows/` folder.

## More readings

- [Get started on Terraform (GCP tutorial)](https://developer.hashicorp.com/terraform/tutorials/gcp-get-started)

- [Best practices for writing Terraform](https://cloud.google.com/docs/terraform/best-practices-for-terraform)

- [API docs of Google Cloud modules](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
