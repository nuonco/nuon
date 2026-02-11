package gcptf

// tmpl is the GCP Terraform stack template.
// It provisions:
// - A VPC network with subnets
// - A GKE-compatible network topology
// - A service account for the runner
// - A GCE instance running the runner
// - A phone-home script that POSTs stack outputs back to the Nuon API
const tmpl = `{
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
    "region": {
      "type": "string",
      "default": "{{.Install.GCPAccount.Region}}"
    },
    "subnet_cidr": {
      "type": "string",
      "default": "10.128.0.0/20"
    },
    "runner_subnet_cidr": {
      "type": "string",
      "default": "10.128.16.0/24"
    },
    "pod_cidr": {
      "type": "string",
      "default": "10.129.0.0/16"
    },
    "service_cidr": {
      "type": "string",
      "default": "10.130.0.0/20"
    }
  },
  "locals": {
    "common_labels": {
      "install-nuon-co-id": "${var.nuon_install_id}",
      "org-nuon-co-id": "${var.nuon_org_id}",
      "app-nuon-co-id": "${var.nuon_app_id}"
    }
  },
  "data": {
    "google_project": {
      "current": {}
    },
    "google_client_config": {
      "current": {}
    }
  },
  "resource": {
    "google_compute_network": {
      "vpc": {
        "name": "${var.nuon_install_id}-vpc",
        "auto_create_subnetworks": false,
        "project": "${data.google_project.current.project_id}"
      }
    },
    "google_compute_subnetwork": {
      "main": {
        "name": "${var.nuon_install_id}-main-subnet",
        "ip_cidr_range": "${var.subnet_cidr}",
        "region": "${var.region}",
        "network": "${google_compute_network.vpc.id}",
        "private_ip_google_access": true,
        "secondary_ip_range": [
          {
            "range_name": "pods",
            "ip_cidr_range": "${var.pod_cidr}"
          },
          {
            "range_name": "services",
            "ip_cidr_range": "${var.service_cidr}"
          }
        ]
      },
      "runner": {
        "name": "${var.nuon_install_id}-runner-subnet",
        "ip_cidr_range": "${var.runner_subnet_cidr}",
        "region": "${var.region}",
        "network": "${google_compute_network.vpc.id}",
        "private_ip_google_access": true
      }
    },
    "google_compute_router": {
      "router": {
        "name": "${var.nuon_install_id}-router",
        "region": "${var.region}",
        "network": "${google_compute_network.vpc.id}"
      }
    },
    "google_compute_router_nat": {
      "nat": {
        "name": "${var.nuon_install_id}-nat",
        "router": "${google_compute_router.router.name}",
        "region": "${var.region}",
        "nat_ip_allocate_option": "AUTO_ONLY",
        "source_subnetwork_ip_ranges_to_nat": "ALL_SUBNETWORKS_ALL_IP_RANGES"
      }
    },
    "google_compute_firewall": {
      "allow_internal": {
        "name": "${var.nuon_install_id}-allow-internal",
        "network": "${google_compute_network.vpc.id}",
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
        "source_ranges": ["${var.subnet_cidr}", "${var.runner_subnet_cidr}"]
      },
      "allow_ssh": {
        "name": "${var.nuon_install_id}-allow-ssh",
        "network": "${google_compute_network.vpc.id}",
        "allow": [
          {
            "protocol": "tcp",
            "ports": ["22"]
          }
        ],
        "source_ranges": ["0.0.0.0/0"],
        "target_tags": ["nuon-runner"]
      }
    },
    "google_service_account": {
      "runner": {
        "account_id": "${substr(var.nuon_install_id, 0, min(28, length(var.nuon_install_id)))}",
        "display_name": "Nuon Runner for ${var.nuon_install_id}"
      }
    },
    "google_project_iam_member": {
      "runner_editor": {
        "project": "${data.google_project.current.project_id}",
        "role": "roles/editor",
        "member": "serviceAccount:${google_service_account.runner.email}"
      },
      "runner_container_admin": {
        "project": "${data.google_project.current.project_id}",
        "role": "roles/container.admin",
        "member": "serviceAccount:${google_service_account.runner.email}"
      },
      "runner_artifact_admin": {
        "project": "${data.google_project.current.project_id}",
        "role": "roles/artifactregistry.admin",
        "member": "serviceAccount:${google_service_account.runner.email}"
      }
    },
    "google_compute_instance": {
      "runner": {
        "name": "${var.nuon_install_id}-runner",
        "machine_type": "e2-medium",
        "zone": "${var.region}-b",
        "tags": ["nuon-runner"],
        "labels": "${local.common_labels}",
        "boot_disk": [
          {
            "initialize_params": [
              {
                "image": "projects/ubuntu-os-cloud/global/images/family/ubuntu-2204-lts",
                "size": 50
              }
            ]
          }
        ],
        "network_interface": [
          {
            "subnetwork": "${google_compute_subnetwork.runner.id}",
            "access_config": [{}]
          }
        ],
        "service_account": [
          {
            "email": "${google_service_account.runner.email}",
            "scopes": ["cloud-platform"]
          }
        ],
        "metadata_startup_script": "#!/bin/bash\n\nRUNNER_ID={{.Runner.ID}}\nRUNNER_API_TOKEN={{.APIToken}}\nRUNNER_API_URL={{.Settings.RunnerAPIURL}}\nGOOGLE_REGION={{.Install.GCPAccount.Region}}\n\n# Remove any existing Docker packages\napt-get remove -y docker docker-engine docker.io containerd runc || true\n\n# Update package index and install prerequisites\napt-get update\napt-get install -y ca-certificates curl gnupg lsb-release\n\n# Add Docker's official GPG key\nmkdir -p /etc/apt/keyrings\ncurl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg\n\n# Set up the repository\necho \"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\" | tee /etc/apt/sources.list.d/docker.list > /dev/null\n\n# Install Docker Engine\napt-get update\napt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin\n\n# Force unmask and start Docker service\nrm -f /etc/systemd/system/docker.service\nrm -f /etc/systemd/system/docker.socket\nsystemctl daemon-reload\nsystemctl unmask docker.service\nsystemctl unmask docker.socket\nsystemctl enable docker\nsystemctl start docker\n\n# Ensure docker group exists and set up runner user\ngroupadd -f docker\nmkdir -p /opt/nuon/runner\nuseradd runner -G docker -c '' -d /opt/nuon/runner -m || true\nchown -R runner:runner /opt/nuon/runner\n\ncat << EOF > /opt/nuon/runner/env\nRUNNER_ID=$RUNNER_ID\nRUNNER_API_TOKEN=$RUNNER_API_TOKEN\nRUNNER_API_URL=$RUNNER_API_URL\nGOOGLE_APPLICATION_CREDENTIALS=/opt/nuon/runner/gcp-credentials.json\nHOST_IP=$(curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip)\nEOF\n\ncat << 'GETIMAGE' > /opt/nuon/runner/get_image_tag.sh\n#!/bin/bash\n\nset -u\n\n. /opt/nuon/runner/env\n\necho \"Fetching runner settings from $RUNNER_API_URL/v1/runners/$RUNNER_ID/settings\"\nRUNNER_SETTINGS=$(curl -s -H \"Authorization: Bearer $RUNNER_API_TOKEN\" \"$RUNNER_API_URL/v1/runners/$RUNNER_ID/settings\")\n\nCONTAINER_IMAGE_URL=$(echo \"$RUNNER_SETTINGS\" | grep -o '\"container_image_url\":\"[^\"]*\"' | cut -d '\"' -f 4)\nCONTAINER_IMAGE_TAG=$(echo \"$RUNNER_SETTINGS\" | grep -o '\"container_image_tag\":\"[^\"]*\"' | cut -d '\"' -f 4)\n\nrm -f /opt/nuon/runner/image\necho \"CONTAINER_IMAGE_URL=$CONTAINER_IMAGE_URL\" >> /opt/nuon/runner/image\necho \"CONTAINER_IMAGE_TAG=$CONTAINER_IMAGE_TAG\" >> /opt/nuon/runner/image\n\nexport CONTAINER_IMAGE_URL=$CONTAINER_IMAGE_URL\nexport CONTAINER_IMAGE_TAG=$CONTAINER_IMAGE_TAG\n\necho \"Using container image: $CONTAINER_IMAGE_URL:$CONTAINER_IMAGE_TAG\"\nGETIMAGE\n\nchmod +x /opt/nuon/runner/get_image_tag.sh\n/opt/nuon/runner/get_image_tag.sh\n\ncat << 'SYSTEMD' > /etc/systemd/system/nuon-runner.service\n[Unit]\nDescription=Nuon Runner Service\nAfter=docker.service\nRequires=docker.service\n\n[Service]\nTimeoutStartSec=0\nUser=runner\nExecStartPre=-/bin/sh -c '/usr/bin/docker stop $(/usr/bin/docker ps -a -q --filter=\"name=%n\")'\nExecStartPre=-/bin/sh -c '/usr/bin/docker rm $(/usr/bin/docker ps -a -q --filter=\"name=%n\")'\nExecStartPre=-/bin/sh -c \"yes | /usr/bin/docker system prune\"\nExecStartPre=-/bin/sh /opt/nuon/runner/get_image_tag.sh\nEnvironmentFile=/opt/nuon/runner/image\nEnvironmentFile=/opt/nuon/runner/env\nExecStartPre=echo \"Using container image: ${CONTAINER_IMAGE_URL}:${CONTAINER_IMAGE_TAG}\"\nExecStartPre=/usr/bin/docker pull ${CONTAINER_IMAGE_URL}:${CONTAINER_IMAGE_TAG}\nExecStart=/usr/bin/docker run --network host -v /tmp/nuon-runner:/tmp --rm --name %n -p 5000:5000 --memory 3750g --cpus 1.75 --env-file /opt/nuon/runner/env ${CONTAINER_IMAGE_URL}:${CONTAINER_IMAGE_TAG} run\nRestart=always\nRestartSec=5\n\n[Install]\nWantedBy=default.target\nSYSTEMD\n\nsystemctl daemon-reload\nsystemctl enable --now nuon-runner\n",
        "depends_on": [
          "google_compute_subnetwork.runner",
          "google_service_account.runner",
          "google_compute_router_nat.nat"
        ]
      }
    },
    "null_resource": {
      "phone_home": {
        "depends_on": [
          "google_compute_network.vpc",
          "google_compute_subnetwork.main",
          "google_compute_subnetwork.runner",
          "google_service_account.runner",
          "google_compute_instance.runner"
        ],
        "provisioner": [
          {
            "local-exec": {
              "command": "curl -X POST '{{.CloudFormationStackVersion.PhoneHomeURL}}' -H 'Content-Type: application/json' -H 'Accept: application/json' --fail --silent --show-error -d '{\"request_type\":\"Create\",\"phone_home_type\":\"gcp\",\"project_id\":\"'${data.google_project.current.project_id}'\",\"project_number\":\"'${data.google_project.current.number}'\",\"region\":\"'${var.region}'\",\"network_id\":\"'${google_compute_network.vpc.id}'\",\"network_name\":\"'${google_compute_network.vpc.name}'\",\"subnet_id\":\"'${google_compute_subnetwork.main.id}'\",\"subnet_name\":\"'${google_compute_subnetwork.main.name}'\",\"runner_subnet_id\":\"'${google_compute_subnetwork.runner.id}'\",\"runner_subnet_name\":\"'${google_compute_subnetwork.runner.name}'\",\"service_account_email\":\"'${google_service_account.runner.email}'\"}'"
            }
          }
        ]
      }
    }
  },
  "terraform": {
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
  "output": {
    "project_id": {
      "value": "${data.google_project.current.project_id}"
    },
    "project_number": {
      "value": "${data.google_project.current.number}"
    },
    "region": {
      "value": "${var.region}"
    },
    "vpc_id": {
      "value": "${google_compute_network.vpc.id}"
    },
    "vpc_name": {
      "value": "${google_compute_network.vpc.name}"
    },
    "main_subnet_id": {
      "value": "${google_compute_subnetwork.main.id}"
    },
    "main_subnet_name": {
      "value": "${google_compute_subnetwork.main.name}"
    },
    "runner_subnet_id": {
      "value": "${google_compute_subnetwork.runner.id}"
    },
    "runner_subnet_name": {
      "value": "${google_compute_subnetwork.runner.name}"
    },
    "service_account_email": {
      "value": "${google_service_account.runner.email}"
    }
  }
}`
