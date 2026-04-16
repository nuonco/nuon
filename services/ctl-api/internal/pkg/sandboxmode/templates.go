package sandboxmode

// SandboxLogTemplate represents a pre-built log template that can be used
// to populate sandbox config log lines.
type SandboxLogTemplate struct {
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Type        string   `json:"type"` // e.g. "failing-action", "kube-action", "success"
	Lines       []string `json:"lines"`
}

// SandboxPlanTemplate represents a pre-built plan template that can be used
// to populate sandbox config plan contents.
type SandboxPlanTemplate struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Type        string `json:"type"` // e.g. "noop", "s3", "database", "full-sandbox"
	Contents    string `json:"contents"`
}

// SandboxTemplates is the response for the templates endpoint.
type SandboxTemplates struct {
	LogTemplates  []SandboxLogTemplate  `json:"log_templates"`
	PlanTemplates []SandboxPlanTemplate `json:"plan_templates"`
}

func DefaultSandboxTemplates() SandboxTemplates {
	return SandboxTemplates{
		LogTemplates:  defaultLogTemplates(),
		PlanTemplates: defaultPlanTemplates(),
	}
}

func defaultLogTemplates() []SandboxLogTemplate {
	return []SandboxLogTemplate{
		{
			Key:         "terraform-apply",
			Description: "Terraform apply output creating AWS resources",
			Category:    "deploy",
			Lines: []string{
				"Terraform v1.7.5",
				"Initializing plugins and modules...",
				"",
				"aws_s3_bucket.main: Creating...",
				"aws_s3_bucket.main: Creation complete after 2s [id=nuon-app-data-prod]",
				"aws_s3_bucket_versioning.main: Creating...",
				"aws_s3_bucket_versioning.main: Creation complete after 1s [id=nuon-app-data-prod]",
				"aws_s3_bucket_server_side_encryption_configuration.main: Creating...",
				"aws_s3_bucket_server_side_encryption_configuration.main: Creation complete after 1s [id=nuon-app-data-prod]",
				"aws_iam_role.app_role: Creating...",
				"aws_iam_role.app_role: Creation complete after 2s [id=nuon-app-role-prod]",
				"aws_iam_role_policy_attachment.s3_access: Creating...",
				"aws_iam_role_policy_attachment.s3_access: Creation complete after 1s [id=nuon-app-role-prod-20240315120000]",
				"aws_security_group.app: Creating...",
				"aws_security_group.app: Creation complete after 3s [id=sg-0abc123def456]",
				"aws_security_group_rule.ingress: Creating...",
				"aws_security_group_rule.ingress: Creation complete after 1s [id=sgrule-abc123]",
				"",
				"Apply complete! Resources: 6 added, 0 changed, 0 destroyed.",
				"",
				"Outputs:",
				"",
				"bucket_arn = \"arn:aws:s3:::nuon-app-data-prod\"",
				"role_arn = \"arn:aws:iam::123456789012:role/nuon-app-role-prod\"",
			},
		},
		{
			Key:         "terraform-plan",
			Description: "Terraform plan output showing resource changes",
			Category:    "deploy",
			Lines: []string{
				"Terraform v1.7.5",
				"Refreshing Terraform state in-memory prior to plan...",
				"",
				"------------------------------------------------------------------------",
				"",
				"An execution plan has been generated and is shown below.",
				"Resource actions are indicated with the following symbols:",
				"  + create",
				"  ~ update in-place",
				"",
				"Terraform will perform the following actions:",
				"",
				"  # aws_s3_bucket.main will be created",
				"  + resource \"aws_s3_bucket\" \"main\" {",
				"      + arn            = (known after apply)",
				"      + bucket         = \"nuon-app-data-prod\"",
				"      + force_destroy  = false",
				"      + id             = (known after apply)",
				"      + tags           = {",
				"          + \"Environment\" = \"production\"",
				"          + \"ManagedBy\"   = \"terraform\"",
				"        }",
				"    }",
				"",
				"  # aws_iam_role.app_role will be created",
				"  + resource \"aws_iam_role\" \"app_role\" {",
				"      + arn                = (known after apply)",
				"      + name               = \"nuon-app-role-prod\"",
				"      + assume_role_policy = jsonencode({...})",
				"    }",
				"",
				"Plan: 2 to add, 0 to change, 0 to destroy.",
			},
		},
		{
			Key:         "helm",
			Description: "Helm install/upgrade output",
			Category:    "deploy",
			Lines: []string{
				"Release \"app-release\" does not exist. Installing it now.",
				"NAME: app-release",
				"LAST DEPLOYED: Mon Mar 15 12:00:00 2024",
				"NAMESPACE: default",
				"STATUS: deployed",
				"REVISION: 1",
				"TEST SUITE: None",
				"NOTES:",
				"  Application has been deployed successfully.",
				"  ",
				"  To access the application:",
				"    export POD_NAME=$(kubectl get pods -l app=app-release -o jsonpath=\"{.items[0].metadata.name}\")",
				"    kubectl port-forward $POD_NAME 8080:8080",
				"  ",
				"  Service endpoint: http://app-release.default.svc.cluster.local:8080",
				"",
				"coalesce.go:286: warning: destination for env is a table. Ignoring non-table value ([])",
				"",
				"HOOKS:",
				"---",
				"# Source: app/templates/tests/test-connection.yaml",
				"apiVersion: v1",
				"kind: Pod",
				"metadata:",
				"  name: \"app-release-test-connection\"",
			},
		},
		{
			Key:         "kubernetes-manifest",
			Description: "kubectl apply output",
			Category:    "deploy",
			Lines: []string{
				"namespace/app-prod created",
				"serviceaccount/app-sa created",
				"configmap/app-config created",
				"secret/app-secrets created",
				"deployment.apps/app-deployment created",
				"service/app-service created",
				"ingress.networking.k8s.io/app-ingress created",
				"horizontalpodautoscaler.autoscaling/app-hpa created",
				"",
				"Waiting for deployment \"app-deployment\" rollout to finish...",
				"  1 of 3 updated replicas are available...",
				"  2 of 3 updated replicas are available...",
				"  3 of 3 updated replicas are available...",
				"deployment \"app-deployment\" successfully rolled out",
			},
		},
		{
			Key:         "docker-build",
			Description: "Docker build output with layer caching",
			Category:    "build",
			Lines: []string{
				"#1 [internal] load build definition from Dockerfile",
				"#1 transferring dockerfile: 847B done",
				"#1 DONE 0.0s",
				"",
				"#2 [internal] load .dockerignore",
				"#2 transferring context: 52B done",
				"#2 DONE 0.0s",
				"",
				"#3 [internal] load metadata for docker.io/library/golang:1.22-alpine",
				"#3 DONE 1.2s",
				"",
				"#4 [build 1/6] FROM docker.io/library/golang:1.22-alpine@sha256:abc123",
				"#4 CACHED",
				"",
				"#5 [build 2/6] WORKDIR /app",
				"#5 CACHED",
				"",
				"#6 [build 3/6] COPY go.mod go.sum ./",
				"#6 DONE 0.1s",
				"",
				"#7 [build 4/6] RUN go mod download",
				"#7 DONE 12.3s",
				"",
				"#8 [build 5/6] COPY . .",
				"#8 DONE 0.5s",
				"",
				"#9 [build 6/6] RUN CGO_ENABLED=0 go build -o /app/server .",
				"#9 DONE 28.7s",
				"",
				"#10 [runtime 1/2] COPY --from=build /app/server /usr/local/bin/",
				"#10 DONE 0.1s",
				"",
				"#11 exporting to image",
				"#11 exporting layers",
				"#11 writing image sha256:def456 done",
				"#11 naming to docker.io/library/app:latest done",
				"#11 DONE 0.2s",
			},
		},
		{
			Key:         "bash",
			Description: "Bash script execution output for actions",
			Category:    "actions",
			Lines: []string{
				"$ echo \"Starting deployment workflow...\"",
				"Starting deployment workflow...",
				"",
				"$ npm run build",
				"> app@1.0.0 build",
				"> next build",
				"",
				"  ▲ Next.js 14.1.0",
				"  - Environments: .env.production",
				"",
				"   Creating an optimized production build ...",
				"   Compiled successfully",
				"   Linting and checking validity of types ...",
				"   Collecting page data ...",
				"   Generating static pages (0/12) ...",
				"   Generating static pages (3/12)",
				"   Generating static pages (6/12)",
				"   Generating static pages (9/12)",
				"   Generating static pages (12/12)",
				"   Finalizing page optimization ...",
				"",
				"Route (app)                    Size     First Load JS",
				"┌ ○ /                          5.42 kB        87.2 kB",
				"├ ○ /about                     3.18 kB        84.9 kB",
				"└ ○ /api/health                0 B            81.7 kB",
				"",
				"○  (Static)  prerendered as static content",
				"",
				"$ echo \"Build completed successfully\"",
				"Build completed successfully",
				"",
				"$ aws s3 sync ./out s3://app-bucket/latest/",
				"upload: out/index.html to s3://app-bucket/latest/index.html",
				"upload: out/about.html to s3://app-bucket/latest/about.html",
				"upload: out/_next/static/chunks/main.js to s3://app-bucket/latest/_next/static/chunks/main.js",
				"",
				"$ echo \"Deployment complete!\"",
				"Deployment complete!",
			},
		},
		{
			Key:         "pulumi",
			Description: "Pulumi up output creating resources",
			Category:    "deploy",
			Lines: []string{
				"Updating (prod)",
				"",
				"View Live: https://app.pulumi.com/nuon/app/prod/updates/42",
				"",
				"     Type                          Name              Status",
				" +   pulumi:pulumi:Stack            app-prod          created",
				" +   ├─ aws:s3:Bucket               app-data          created",
				" +   ├─ aws:s3:BucketPolicy         app-data-policy   created",
				" +   ├─ aws:iam:Role                 app-role          created",
				" +   ├─ aws:iam:RolePolicy           app-policy        created",
				" +   └─ aws:cloudfront:Distribution  app-cdn           created",
				"",
				"Outputs:",
				"    bucketName  : \"app-data-a1b2c3d\"",
				"    cdnDomain   : \"d1234567.cloudfront.net\"",
				"    roleArn     : \"arn:aws:iam::123456789012:role/app-role-prod\"",
				"",
				"Resources:",
				"    + 6 created",
				"",
				"Duration: 45s",
			},
		},
		{
			Key:         "oci-sync",
			Description: "OCI artifact sync output",
			Category:    "sync",
			Lines: []string{
				"Pulling artifact from registry...",
				"Resolved digest: sha256:abc123def456",
				"Downloading layer 1/3: config (256 bytes)",
				"Downloading layer 2/3: application/vnd.oci.image.layer.v1.tar+gzip (12.4 MB)",
				"Downloading layer 3/3: application/vnd.oci.image.layer.v1.tar+gzip (3.2 MB)",
				"Download complete.",
				"",
				"Verifying artifact integrity...",
				"Checksum verified: OK",
				"",
				"Pushing artifact to target registry...",
				"Uploading layer 1/3: config",
				"Uploading layer 2/3: application/vnd.oci.image.layer.v1.tar+gzip",
				"Uploading layer 3/3: application/vnd.oci.image.layer.v1.tar+gzip",
				"Pushed successfully to target.",
				"",
				"Artifact synced: sha256:abc123def456",
			},
		},
	}
}

func defaultPlanTemplates() []SandboxPlanTemplate {
	return []SandboxPlanTemplate{
		{
			Key:         "terraform-s3",
			Description: "Terraform plan: create S3 bucket with versioning and encryption",
			Category:    "terraform",
			Contents: `Terraform will perform the following actions:

  # aws_s3_bucket.main will be created
  + resource "aws_s3_bucket" "main" {
      + acceleration_status         = (known after apply)
      + acl                         = (known after apply)
      + arn                         = (known after apply)
      + bucket                      = "nuon-app-data-prod"
      + bucket_domain_name          = (known after apply)
      + bucket_prefix               = (known after apply)
      + bucket_regional_domain_name = (known after apply)
      + force_destroy               = false
      + hosted_zone_id              = (known after apply)
      + id                          = (known after apply)
      + region                      = (known after apply)
      + request_payer               = (known after apply)
      + tags                        = {
          + "Environment" = "production"
          + "ManagedBy"   = "nuon"
        }
      + tags_all                    = (known after apply)
      + website_domain              = (known after apply)
      + website_endpoint            = (known after apply)
    }

  # aws_s3_bucket_versioning.main will be created
  + resource "aws_s3_bucket_versioning" "main" {
      + bucket = (known after apply)
      + id     = (known after apply)

      + versioning_configuration {
          + mfa_delete = (known after apply)
          + status     = "Enabled"
        }
    }

  # aws_s3_bucket_server_side_encryption_configuration.main will be created
  + resource "aws_s3_bucket_server_side_encryption_configuration" "main" {
      + bucket = (known after apply)
      + id     = (known after apply)

      + rule {
          + apply_server_side_encryption_by_default {
              + sse_algorithm = "aws:kms"
            }
          + bucket_key_enabled = true
        }
    }

Plan: 3 to add, 0 to change, 0 to destroy.`,
		},
		{
			Key:         "terraform-iam",
			Description: "Terraform plan: create IAM role with policy",
			Category:    "terraform",
			Contents: `Terraform will perform the following actions:

  # aws_iam_role.app will be created
  + resource "aws_iam_role" "app" {
      + arn                   = (known after apply)
      + assume_role_policy    = jsonencode(
            {
              + Statement = [
                  + {
                      + Action    = "sts:AssumeRole"
                      + Effect    = "Allow"
                      + Principal = {
                          + Service = "ecs-tasks.amazonaws.com"
                        }
                    },
                ]
              + Version   = "2012-10-17"
            }
        )
      + create_date           = (known after apply)
      + force_detach_policies = false
      + id                    = (known after apply)
      + name                  = "nuon-app-role-prod"
      + path                  = "/"
      + tags                  = {
          + "Environment" = "production"
          + "ManagedBy"   = "nuon"
        }
      + unique_id             = (known after apply)
    }

  # aws_iam_policy.app will be created
  + resource "aws_iam_policy" "app" {
      + arn         = (known after apply)
      + id          = (known after apply)
      + name        = "nuon-app-policy-prod"
      + path        = "/"
      + policy      = jsonencode(
            {
              + Statement = [
                  + {
                      + Action   = [
                          + "s3:GetObject",
                          + "s3:PutObject",
                          + "s3:ListBucket",
                        ]
                      + Effect   = "Allow"
                      + Resource = [
                          + "arn:aws:s3:::nuon-app-data-prod",
                          + "arn:aws:s3:::nuon-app-data-prod/*",
                        ]
                    },
                ]
              + Version   = "2012-10-17"
            }
        )
      + tags_all    = (known after apply)
    }

  # aws_iam_role_policy_attachment.app will be created
  + resource "aws_iam_role_policy_attachment" "app" {
      + id         = (known after apply)
      + policy_arn = (known after apply)
      + role       = "nuon-app-role-prod"
    }

Plan: 3 to add, 0 to change, 0 to destroy.`,
		},
		{
			Key:         "terraform-vpc",
			Description: "Terraform plan: create VPC with subnets and security groups",
			Category:    "terraform",
			Contents: `Terraform will perform the following actions:

  # aws_vpc.main will be created
  + resource "aws_vpc" "main" {
      + arn                  = (known after apply)
      + cidr_block           = "10.0.0.0/16"
      + enable_dns_hostnames = true
      + enable_dns_support   = true
      + id                   = (known after apply)
      + tags                 = {
          + "Name" = "nuon-vpc-prod"
        }
    }

  # aws_subnet.public[0] will be created
  + resource "aws_subnet" "public" {
      + arn                = (known after apply)
      + availability_zone  = "us-west-2a"
      + cidr_block         = "10.0.1.0/24"
      + id                 = (known after apply)
      + tags               = {
          + "Name" = "nuon-public-subnet-1"
        }
    }

  # aws_subnet.public[1] will be created
  + resource "aws_subnet" "public" {
      + arn                = (known after apply)
      + availability_zone  = "us-west-2b"
      + cidr_block         = "10.0.2.0/24"
      + id                 = (known after apply)
      + tags               = {
          + "Name" = "nuon-public-subnet-2"
        }
    }

  # aws_security_group.app will be created
  + resource "aws_security_group" "app" {
      + arn         = (known after apply)
      + description = "Security group for application"
      + id          = (known after apply)
      + name        = "nuon-app-sg-prod"
      + vpc_id      = (known after apply)

      + ingress {
          + cidr_blocks = ["0.0.0.0/0"]
          + from_port   = 443
          + protocol    = "tcp"
          + to_port     = 443
        }

      + egress {
          + cidr_blocks = ["0.0.0.0/0"]
          + from_port   = 0
          + protocol    = "-1"
          + to_port     = 0
        }
    }

Plan: 4 to add, 0 to change, 0 to destroy.`,
		},
		{
			Key:         "helm-nginx",
			Description: "Helm diff: nginx ingress controller install",
			Category:    "helm",
			Contents: `default, ingress-nginx-controller, Deployment (apps) has been added:
-
+ apiVersion: apps/v1
+ kind: Deployment
+ metadata:
+   name: ingress-nginx-controller
+   namespace: ingress-nginx
+   labels:
+     app.kubernetes.io/name: ingress-nginx
+     app.kubernetes.io/component: controller
+ spec:
+   replicas: 2
+   selector:
+     matchLabels:
+       app.kubernetes.io/name: ingress-nginx
+   template:
+     metadata:
+       labels:
+         app.kubernetes.io/name: ingress-nginx
+     spec:
+       containers:
+       - name: controller
+         image: registry.k8s.io/ingress-nginx/controller:v1.9.6
+         ports:
+         - containerPort: 80
+           name: http
+         - containerPort: 443
+           name: https
+         resources:
+           requests:
+             cpu: 100m
+             memory: 128Mi
+           limits:
+             cpu: 500m
+             memory: 512Mi

default, ingress-nginx-controller, Service (v1) has been added:
-
+ apiVersion: v1
+ kind: Service
+ metadata:
+   name: ingress-nginx-controller
+   namespace: ingress-nginx
+ spec:
+   type: LoadBalancer
+   ports:
+   - name: http
+     port: 80
+     targetPort: http
+   - name: https
+     port: 443
+     targetPort: https
+   selector:
+     app.kubernetes.io/name: ingress-nginx`,
		},
		{
			Key:         "helm-app",
			Description: "Helm diff: application chart with deployment and service",
			Category:    "helm",
			Contents: `default, app-deployment, Deployment (apps) has been added:
-
+ apiVersion: apps/v1
+ kind: Deployment
+ metadata:
+   name: app-deployment
+   namespace: default
+   labels:
+     app: myapp
+     version: "1.2.3"
+ spec:
+   replicas: 3
+   selector:
+     matchLabels:
+       app: myapp
+   template:
+     metadata:
+       labels:
+         app: myapp
+     spec:
+       containers:
+       - name: app
+         image: 123456789012.dkr.ecr.us-west-2.amazonaws.com/myapp:1.2.3
+         ports:
+         - containerPort: 8080
+         env:
+         - name: DATABASE_URL
+           valueFrom:
+             secretKeyRef:
+               name: app-secrets
+               key: database-url
+         resources:
+           requests:
+             cpu: 250m
+             memory: 256Mi
+           limits:
+             cpu: "1"
+             memory: 1Gi

default, app-service, Service (v1) has been added:
-
+ apiVersion: v1
+ kind: Service
+ metadata:
+   name: app-service
+   namespace: default
+ spec:
+   type: ClusterIP
+   ports:
+   - port: 80
+     targetPort: 8080
+   selector:
+     app: myapp

default, app-config, ConfigMap (v1) has been added:
-
+ apiVersion: v1
+ kind: ConfigMap
+ metadata:
+   name: app-config
+   namespace: default
+ data:
+   LOG_LEVEL: "info"
+   API_PORT: "8080"
+   CACHE_TTL: "300"`,
		},
		{
			Key:         "kube-manifest",
			Description: "kubectl diff: deployment and service changes",
			Category:    "kubernetes",
			Contents: `diff -u -N /tmp/existing/apps.v1.Deployment.default.app-deployment /tmp/desired/apps.v1.Deployment.default.app-deployment
--- /tmp/existing/apps.v1.Deployment.default.app-deployment
+++ /tmp/desired/apps.v1.Deployment.default.app-deployment
@@ -10,7 +10,7 @@
   labels:
     app: myapp
 spec:
-  replicas: 2
+  replicas: 3
   selector:
     matchLabels:
       app: myapp
@@ -22,7 +22,7 @@
       containers:
       - name: app
-        image: myapp:1.2.2
+        image: myapp:1.2.3
         ports:
         - containerPort: 8080
         resources:
@@ -30,6 +30,10 @@
             cpu: 250m
             memory: 256Mi
+          limits:
+            cpu: "1"
+            memory: 1Gi
+        readinessProbe:
+          httpGet:
+            path: /healthz
+            port: 8080

diff -u -N /tmp/existing/v1.Service.default.app-service /tmp/desired/v1.Service.default.app-service
--- /tmp/existing/v1.Service.default.app-service
+++ /tmp/desired/v1.Service.default.app-service
@@ -8,6 +8,8 @@
   ports:
   - port: 80
     targetPort: 8080
+  - port: 9090
+    targetPort: 9090
+    name: metrics
   selector:
     app: myapp`,
		},
		{
			Key:         "pulumi-s3",
			Description: "Pulumi preview: S3 bucket with CloudFront",
			Category:    "pulumi",
			Contents: `Previewing update (prod)

View Live: https://app.pulumi.com/nuon/infra/prod/previews/abc123

     Type                              Name              Plan
 +   pulumi:pulumi:Stack               infra-prod        create
 +   ├─ aws:s3:Bucket                  app-data          create
 +   │  ├─ BucketVersioningV2          app-data          create
 +   │  └─ BucketEncryptionV2          app-data          create
 +   ├─ aws:cloudfront:Distribution    app-cdn           create
 +   │  └─ aws:cloudfront:OAI          app-cdn-oai       create
 +   └─ aws:route53:Record             app-dns           create

Resources:
    + 7 to create

Do you want to perform this update?
  yes
> no
  details`,
		},
		{
			Key:         "pulumi-eks",
			Description: "Pulumi preview: EKS cluster creation",
			Category:    "pulumi",
			Contents: `Previewing update (prod)

View Live: https://app.pulumi.com/nuon/k8s/prod/previews/def456

     Type                              Name              Plan
 +   pulumi:pulumi:Stack               k8s-prod          create
 +   ├─ aws:ec2:Vpc                    cluster-vpc       create
 +   ├─ aws:ec2:Subnet                 cluster-pub-1     create
 +   ├─ aws:ec2:Subnet                 cluster-pub-2     create
 +   ├─ aws:ec2:Subnet                 cluster-priv-1    create
 +   ├─ aws:ec2:Subnet                 cluster-priv-2    create
 +   ├─ aws:ec2:InternetGateway        cluster-igw       create
 +   ├─ aws:ec2:NatGateway             cluster-nat       create
 +   ├─ aws:iam:Role                   cluster-role      create
 +   ├─ aws:iam:Role                   node-role         create
 +   ├─ aws:eks:Cluster                cluster           create
 +   ├─ aws:eks:NodeGroup              default-nodes     create
 +   └─ aws:eks:Addon                  vpc-cni           create

Resources:
    + 13 to create

Do you want to perform this update?
  yes
> no
  details`,
		},
		// Noop plans (no changes)
		{
			Key:         "terraform-noop",
			Description: "Terraform plan: no changes detected",
			Category:    "terraform",
			Contents: `Running terraform plan...

No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration
and found no differences, so no changes are needed.`,
		},
		{
			Key:         "helm-noop",
			Description: "Helm diff: no changes detected",
			Category:    "helm",
			Contents: `Release "app" has been compared to the running version.
There are no differences to apply.

No changes detected between the installed release and the chart.`,
		},
		{
			Key:         "kube-manifest-noop",
			Description: "kubectl diff: no changes detected",
			Category:    "kubernetes-manifest",
			Contents: `comparing against live cluster state...

No differences found. All resources are up-to-date.

deployment.apps/app-deployment unchanged
service/app-service unchanged
configmap/app-config unchanged`,
		},
		{
			Key:         "pulumi-noop",
			Description: "Pulumi preview: no changes detected",
			Category:    "pulumi",
			Contents: `Previewing update (prod)

View Live: https://app.pulumi.com/nuon/infra/prod/previews/abc123

     Type                 Name          Plan
     pulumi:pulumi:Stack  infra-prod

Resources:
    5 unchanged

Duration: 3s`,
		},
		// Destroy/teardown plans
		{
			Key:         "terraform-destroy",
			Description: "Terraform plan: destroying all resources",
			Category:    "terraform",
			Contents: `Running terraform plan -destroy...

Terraform will perform the following actions:

  # aws_s3_bucket.data will be destroyed
  - resource "aws_s3_bucket" "data" {
      - id                          = "nuon-app-data-prod" -> null
      - bucket                      = "nuon-app-data-prod" -> null
      - region                      = "us-west-2" -> null
      - versioning {
          - enabled = true -> null
        }
    }

  # aws_iam_role.app_role will be destroyed
  - resource "aws_iam_role" "app_role" {
      - id                    = "nuon-app-role" -> null
      - name                  = "nuon-app-role" -> null
      - assume_role_policy    = jsonencode({...}) -> null
    }

  # aws_iam_role_policy_attachment.app_policy will be destroyed
  - resource "aws_iam_role_policy_attachment" "app_policy" {
      - role       = "nuon-app-role" -> null
      - policy_arn = "arn:aws:iam::policy/nuon-app-policy" -> null
    }

Plan: 0 to add, 0 to change, 3 to destroy.`,
		},
		{
			Key:         "helm-uninstall",
			Description: "Helm uninstall: removing release",
			Category:    "helm",
			Contents: `Uninstalling release "app" from namespace "default"...

These resources will be deleted:
  - deployment.apps/app-deployment
  - service/app-service
  - configmap/app-config
  - serviceaccount/app-sa
  - ingress.networking.k8s.io/app-ingress

release "app" uninstalled

All resources associated with the release have been removed.`,
		},
		{
			Key:         "pulumi-destroy",
			Description: "Pulumi destroy: removing all resources",
			Category:    "pulumi",
			Contents: `Previewing destroy (prod)

View Live: https://app.pulumi.com/nuon/infra/prod/previews/destroy789

     Type                              Name              Plan
 -   pulumi:pulumi:Stack               infra-prod        delete
 -   ├─ aws:s3:BucketObject            index-html        delete
 -   ├─ aws:cloudfront:Distribution    cdn               delete
 -   ├─ aws:s3:BucketPolicy            bucket-policy     delete
 -   └─ aws:s3:Bucket                  website-bucket    delete

Resources:
    - 5 to delete

Do you want to perform this destroy?
  yes
> no
  details`,
		},
		{
			Key:         "kube-manifest-destroy",
			Description: "kubectl delete: removing resources",
			Category:    "kubernetes-manifest",
			Contents: `Deleting resources from namespace "default"...

deployment.apps "app-deployment" deleted
service "app-service" deleted
configmap "app-config" deleted
horizontalpodautoscaler.autoscaling "app-hpa" deleted
poddisruptionbudget.policy "app-pdb" deleted

All resources successfully deleted.`,
		},
		// Update/change plans
		{
			Key:         "terraform-update",
			Description: "Terraform plan: updating existing resources",
			Category:    "terraform",
			Contents: `Running terraform plan...

Terraform will perform the following actions:

  # aws_s3_bucket.data will be updated in-place
  ~ resource "aws_s3_bucket" "data" {
        id     = "nuon-app-data-prod"
      ~ tags   = {
          + "Environment" = "production"
          + "ManagedBy"   = "terraform"
        }
    }

  # aws_iam_role_policy.app_policy will be updated in-place
  ~ resource "aws_iam_role_policy" "app_policy" {
        id     = "nuon-app-role:nuon-app-policy"
      ~ policy = jsonencode(
          ~ {
              ~ Statement = [
                  ~ {
                      ~ Action   = [
                          + "s3:GetObject",
                          + "s3:PutObject",
                            "s3:ListBucket",
                        ]
                    },
                ]
            }
        )
    }

Plan: 0 to add, 2 to change, 0 to destroy.`,
		},
		{
			Key:         "helm-upgrade",
			Description: "Helm diff: upgrading release with changes",
			Category:    "helm",
			Contents: `Comparing release "app" (revision 3) to chart version 1.5.0...

default, app-deployment, Deployment (apps) has changed:
  spec:
    template:
      spec:
        containers:
        - name: app
-         image: nuon/app:v1.4.0
+         image: nuon/app:v1.5.0
          resources:
            limits:
-             memory: "256Mi"
+             memory: "512Mi"
            requests:
-             memory: "128Mi"
+             memory: "256Mi"

default, app-config, ConfigMap (v1) has changed:
  data:
-   LOG_LEVEL: "info"
+   LOG_LEVEL: "debug"
+   FEATURE_FLAG_V2: "true"

Upgrading release "app" in namespace "default"...
Release "app" has been upgraded. Happy Helming!
REVISION: 4`,
		},
		{
			Key:         "pulumi-update",
			Description: "Pulumi preview: updating existing resources",
			Category:    "pulumi",
			Contents: `Previewing update (prod)

View Live: https://app.pulumi.com/nuon/infra/prod/previews/update456

     Type                              Name              Plan       Info
     pulumi:pulumi:Stack               infra-prod
 ~   ├─ aws:s3:Bucket                  website-bucket    update     [diff: +tags]
 ~   ├─ aws:cloudfront:Distribution    cdn               update     [diff: ~defaultCacheBehavior]
 +-  └─ aws:s3:BucketObject            index-html        replace    [diff: ~source]

Resources:
    ~ 2 to update
    +-1 to replace
    2 unchanged

Do you want to perform this update?
  yes
> no
  details`,
		},
	}
}
