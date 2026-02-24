package gcp

const tmpl = `{
  "terraform": {
    "required_version": ">= 1.5",
    "required_providers": {
      "google": {
        "source": "hashicorp/google",
        "version": ">= 5.0"
      },
      "null": {
        "source": "hashicorp/null",
        "version": ">= 3.0"
      }
    }
  },
  "provider": {
    "google": {
      "project": "{{.Install.GCPAccount.ProjectID}}",
      "region": "{{.Install.GCPAccount.Region}}"
    }
  },
  "variable": {
    "nuon_install_id": {
      "type": "string",
      "default": "{{.Install.ID}}"
    },
    "nuon_org_id": {
      "type": "string",
      "default": "{{.Runner.OrgID}}"
    },
    "nuon_app_id": {
      "type": "string",
      "default": "{{.Install.AppID}}"
    },
    "runner_api_url": {
      "type": "string",
      "default": "{{.Settings.RunnerAPIURL}}"
    },
    "runner_api_token": {
      "type": "string",
      "default": "{{.APIToken}}",
      "sensitive": true
    },
    "runner_id": {
      "type": "string",
      "default": "{{.Runner.ID}}"
    },
    "runner_init_script_url": {
      "type": "string",
      "default": "{{.RunnerInitScriptURL}}"
    }
  },
  "locals": {
    "prefix": "{{.Install.ID}}",
    "region": "{{.Install.GCPAccount.Region}}",
    "labels": {
      "nuon-install-id": "{{.Install.ID}}",
      "nuon-org-id": "{{.Runner.OrgID}}",
      "nuon-app-id": "{{.Install.AppID}}",
      "managed-by": "nuon"
    }
  },
  "resource": {
    "google_compute_network": {
      "main": {
        "name": "${local.prefix}-vpc",
        "auto_create_subnetworks": false
      }
    },
    "google_compute_subnetwork": {
      "public": {
        "name": "${local.prefix}-public-subnet",
        "region": "${local.region}",
        "network": "${google_compute_network.main.id}",
        "ip_cidr_range": "10.128.0.0/24",
        "private_ip_google_access": true
      },
      "private": {
        "name": "${local.prefix}-private-subnet",
        "region": "${local.region}",
        "network": "${google_compute_network.main.id}",
        "ip_cidr_range": "10.128.1.0/24",
        "private_ip_google_access": true
      },
      "runner": {
        "name": "${local.prefix}-runner-subnet",
        "region": "${local.region}",
        "network": "${google_compute_network.main.id}",
        "ip_cidr_range": "10.128.2.0/24",
        "private_ip_google_access": true
      }
    },
    "google_compute_router": {
      "main": {
        "name": "${local.prefix}-router",
        "region": "${local.region}",
        "network": "${google_compute_network.main.id}"
      }
    },
    "google_compute_router_nat": {
      "main": {
        "name": "${local.prefix}-nat",
        "router": "${google_compute_router.main.name}",
        "region": "${local.region}",
        "nat_ip_allocate_option": "AUTO_ONLY",
        "source_subnetwork_ip_ranges_to_nat": "ALL_SUBNETWORKS_ALL_IP_RANGES"
      }
    },
    "google_compute_firewall": {
      "allow_internal": {
        "name": "${local.prefix}-allow-internal",
        "network": "${google_compute_network.main.name}",
        "allow": [
          {
            "protocol": "tcp",
            "ports": ["0-65535"]
          },
          {
            "protocol": "udp",
            "ports": ["0-65535"]
          },
          {
            "protocol": "icmp"
          }
        ],
        "source_ranges": ["10.128.0.0/16"]
      },
      "allow_egress": {
        "name": "${local.prefix}-allow-egress",
        "network": "${google_compute_network.main.name}",
        "direction": "EGRESS",
        "allow": [
          {
            "protocol": "all"
          }
        ],
        "destination_ranges": ["0.0.0.0/0"]
      }
    },
    "google_service_account": {
      "runner": {
        "account_id": "${substr(local.prefix, 0, 23)}-runner",
        "display_name": "Nuon runner for ${local.prefix}"
      }
    },
    "google_project_iam_member": {
      "runner_container_admin": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "roles/container.admin",
        "member": "serviceAccount:${google_service_account.runner.email}"
      },
      "runner_compute_admin": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "roles/compute.networkAdmin",
        "member": "serviceAccount:${google_service_account.runner.email}"
      },
      "runner_artifact_registry": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "roles/artifactregistry.admin",
        "member": "serviceAccount:${google_service_account.runner.email}"
      },
      "runner_dns_admin": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "roles/dns.admin",
        "member": "serviceAccount:${google_service_account.runner.email}"
      },
      "runner_sa_user": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "roles/iam.serviceAccountUser",
        "member": "serviceAccount:${google_service_account.runner.email}"
      }
    },
    "google_compute_instance": {
      "runner": {
        "name": "${local.prefix}-runner",
        "machine_type": "e2-medium",
        "zone": "${local.region}-a",
        "labels": "${local.labels}",
        "tags": ["nuon-runner"],
        "boot_disk": [
          {
            "initialize_params": [
              {
                "image": "ubuntu-os-cloud/ubuntu-2204-lts",
                "size": 30,
                "type": "pd-balanced"
              }
            ]
          }
        ],
        "network_interface": [
          {
            "subnetwork": "${google_compute_subnetwork.runner.id}"
          }
        ],
        "service_account": [
          {
            "email": "${google_service_account.runner.email}",
            "scopes": ["cloud-platform"]
          }
        ],
        "metadata_startup_script": "#!/bin/bash\nset -e\nexport NUON_RUNNER_ID=${var.runner_id}\nexport NUON_RUNNER_API_URL=${var.runner_api_url}\nexport NUON_RUNNER_API_TOKEN=${var.runner_api_token}\nexport NUON_INSTALL_ID=${var.nuon_install_id}\ncurl -fsSL ${var.runner_init_script_url} | bash\n"
      }
    },
    "null_resource": {
      "phone_home": {
        "depends_on": [
          "google_compute_instance.runner",
          "google_service_account.runner",
          "google_compute_network.main",
          "google_compute_subnetwork.public",
          "google_compute_subnetwork.private",
          "google_compute_subnetwork.runner"
        ],
        "provisioner": {
          "local-exec": {
            "command": "curl -sf -X POST '{{.CloudFormationStackVersion.PhoneHomeURL}}' -H 'Content-Type: application/json' -d '{\"request_type\":\"Create\",\"phone_home_type\":\"gcp\",\"project_id\":\"{{.Install.GCPAccount.ProjectID}}\",\"region\":\"{{.Install.GCPAccount.Region}}\",\"network_name\":\"${google_compute_network.main.name}\",\"network_id\":\"${google_compute_network.main.id}\",\"public_subnet_name\":\"${google_compute_subnetwork.public.name}\",\"private_subnet_name\":\"${google_compute_subnetwork.private.name}\",\"runner_subnet_name\":\"${google_compute_subnetwork.runner.name}\",\"runner_service_account_email\":\"${google_service_account.runner.email}\"}'"
          }
        }
      }
    }
  },
  "output": {
    "project_id": {
      "value": "{{.Install.GCPAccount.ProjectID}}"
    },
    "region": {
      "value": "${local.region}"
    },
    "network_name": {
      "value": "${google_compute_network.main.name}"
    },
    "network_id": {
      "value": "${google_compute_network.main.id}"
    },
    "public_subnet_name": {
      "value": "${google_compute_subnetwork.public.name}"
    },
    "private_subnet_name": {
      "value": "${google_compute_subnetwork.private.name}"
    },
    "runner_subnet_name": {
      "value": "${google_compute_subnetwork.runner.name}"
    },
    "runner_service_account_email": {
      "value": "${google_service_account.runner.email}"
    }
  }
}`
