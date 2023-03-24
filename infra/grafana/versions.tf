terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "launchpaddev"

    workspaces {
      name = "infra-grafana"
    }
  }

  required_providers {
    grafana = {
      source  = "grafana/grafana"
      version = "1.34.0"
    }
  }
}
