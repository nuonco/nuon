locals {
  labels = {           # custom labels applied to resources (whenever possible)
    Terraform = "true" # for easily identifying Terraform-managed resources in GCP dashboard
  }
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.5.0"
    }
  }

  backend "s3" {}
}

provider "aws" {
  # TODO: this should really be a reference to the sandbox region.
  # We should return to this once the connected components feature is complete.
  region = "us-west-2"

  default_tags {
    tags = {
      Terraform   = "true"
    }
  }
}

module "gaap_cap_workflow" {
  source = "./modules/gaap-cap-workflow"
}

module "database" {
  source                      = "terraform-aws-modules/rds/aws"
  identifier                  = "database"
  engine                      = "postgres"
  engine_version              = "14"
  family                      = "postgres14"
  instance_class              = "db.t4g.micro"
  multi_az                    = true
  manage_master_user_password = true
}
