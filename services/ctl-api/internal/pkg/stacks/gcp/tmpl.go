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
      },
      "provision": {
        "account_id": "${substr(local.prefix, 0, 20)}-prov",
        "display_name": "Nuon provision for ${local.prefix}"
      },
      "maintenance": {
        "account_id": "${substr(local.prefix, 0, 20)}-maint",
        "display_name": "Nuon maintenance for ${local.prefix}"
      },
      "deprovision": {
        "account_id": "${substr(local.prefix, 0, 20)}-dep",
        "display_name": "Nuon deprovision for ${local.prefix}"
      }{{if .HasBreakGlass}},
      "break_glass": {
        "account_id": "${substr(local.prefix, 0, 20)}-bg",
        "display_name": "Nuon break glass for ${local.prefix}"
      }{{end}}
    },
    "google_project_iam_custom_role": {
      "provision": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role_id": "${local.prefix}_prov_role",
        "title": "Nuon Provision for ${local.prefix}",
        "permissions": {{.ProvisionPermissions}}
      },
      "maintenance": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role_id": "${local.prefix}_maint_role",
        "title": "Nuon Maintenance for ${local.prefix}",
        "permissions": {{.MaintenancePermissions}}
      },
      "deprovision": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role_id": "${local.prefix}_dep_role",
        "title": "Nuon Deprovision for ${local.prefix}",
        "permissions": {{.DeprovisionPermissions}}
      }{{if .HasBreakGlass}},
      "break_glass": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role_id": "${local.prefix}_bg_role",
        "title": "Nuon Break Glass for ${local.prefix}",
        "permissions": {{.BreakGlassPermissions}}
      }{{end}}
    },
    "google_project_iam_member": {
      "runner_container_admin": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "roles/container.admin",
        "member": "serviceAccount:${google_service_account.runner.email}"
      },
      "provision_role_binding": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "${google_project_iam_custom_role.provision.id}",
        "member": "serviceAccount:${google_service_account.provision.email}"
      },
      "maintenance_role_binding": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "${google_project_iam_custom_role.maintenance.id}",
        "member": "serviceAccount:${google_service_account.maintenance.email}"
      },
      "deprovision_role_binding": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "${google_project_iam_custom_role.deprovision.id}",
        "member": "serviceAccount:${google_service_account.deprovision.email}"
      }{{if .HasBreakGlass}},
      "break_glass_role_binding": {
        "project": "{{.Install.GCPAccount.ProjectID}}",
        "role": "${google_project_iam_custom_role.break_glass.id}",
        "member": "serviceAccount:${google_service_account.break_glass.email}"
      }{{end}}
    },
    "google_service_account_iam_member": {
      "provision_token_creator": {
        "service_account_id": "${google_service_account.provision.name}",
        "role": "roles/iam.serviceAccountTokenCreator",
        "member": "serviceAccount:${google_service_account.runner.email}"
      },
      "maintenance_token_creator": {
        "service_account_id": "${google_service_account.maintenance.name}",
        "role": "roles/iam.serviceAccountTokenCreator",
        "member": "serviceAccount:${google_service_account.runner.email}"
      },
      "deprovision_token_creator": {
        "service_account_id": "${google_service_account.deprovision.name}",
        "role": "roles/iam.serviceAccountTokenCreator",
        "member": "serviceAccount:${google_service_account.runner.email}"
      }{{if .HasBreakGlass}},
      "break_glass_token_creator": {
        "service_account_id": "${google_service_account.break_glass.name}",
        "role": "roles/iam.serviceAccountTokenCreator",
        "member": "serviceAccount:${google_service_account.runner.email}"
      }{{end}}
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
        "metadata_startup_script": "#!/bin/bash\nset -e\nexport NUON_RUNNER_ID=${var.runner_id}\nexport NUON_RUNNER_API_URL=${var.runner_api_url}\nexport NUON_RUNNER_API_TOKEN=${var.runner_api_token}\nexport NUON_INSTALL_ID=${var.nuon_install_id}\ncurl -fsSL ${var.runner_init_script_url} | bash\n",
        "lifecycle": {
          "ignore_changes": ["metadata_startup_script"]
        }
      }
    },
    "null_resource": {
      "phone_home": {
        "depends_on": [
          "google_compute_instance.runner",
          "google_service_account.runner",
          "google_service_account.provision",
          "google_service_account.maintenance",
          "google_service_account.deprovision"{{if .HasBreakGlass}},
          "google_service_account.break_glass"{{end}},
          "google_compute_network.main",
          "google_compute_subnetwork.public",
          "google_compute_subnetwork.private",
          "google_compute_subnetwork.runner"
        ],
        "triggers": {
          "phone_home_url": "{{.CloudFormationStackVersion.PhoneHomeURL}}"
        },
        "provisioner": {
          "local-exec": {
            "command": "curl -sf -X POST '{{.CloudFormationStackVersion.PhoneHomeURL}}' -H 'Content-Type: application/json' -d '{\"request_type\":\"Create\",\"phone_home_type\":\"gcp\",\"project_id\":\"{{.Install.GCPAccount.ProjectID}}\",\"region\":\"{{.Install.GCPAccount.Region}}\",\"network_name\":\"${google_compute_network.main.name}\",\"network_id\":\"${google_compute_network.main.id}\",\"public_subnet_name\":\"${google_compute_subnetwork.public.name}\",\"private_subnet_name\":\"${google_compute_subnetwork.private.name}\",\"runner_subnet_name\":\"${google_compute_subnetwork.runner.name}\",\"runner_service_account_email\":\"${google_service_account.runner.email}\",\"provision_sa_email\":\"${google_service_account.provision.email}\",\"maintenance_sa_email\":\"${google_service_account.maintenance.email}\",\"deprovision_sa_email\":\"${google_service_account.deprovision.email}\"{{if .HasBreakGlass}},\"break_glass_sa_email\":\"${google_service_account.break_glass.email}\"{{end}}}'"
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
    },
    "provision_sa_email": {
      "value": "${google_service_account.provision.email}"
    },
    "maintenance_sa_email": {
      "value": "${google_service_account.maintenance.email}"
    },
    "deprovision_sa_email": {
      "value": "${google_service_account.deprovision.email}"
    }{{if .HasBreakGlass}},
    "break_glass_sa_email": {
      "value": "${google_service_account.break_glass.email}"
    }{{end}}
  }
}`
