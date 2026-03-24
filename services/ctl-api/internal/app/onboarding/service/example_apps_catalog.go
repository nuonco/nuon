package service

// ExampleApp represents a pre-configured example application available for onboarding.
type ExampleApp struct {
	Slug          string   `json:"slug"`
	DisplayName   string   `json:"display_name"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Difficulty    string   `json:"difficulty"`
	Tags          []string `json:"tags"`
	CloudProvider string   `json:"cloud_provider"`
	Repo          string   `json:"repo"`
	Directory     string   `json:"directory"`
	Branch        string   `json:"branch"`
}

var exampleAppsCatalog = []ExampleApp{
	{
		Slug:          "httpbin",
		DisplayName:   "HTTPBin",
		Description:   "HTTPBin app on AWS EC2 using our minimal AWS Sandbox",
		Category:      "architecture",
		Difficulty:    "simple",
		Tags:          []string{"ec2", "docker", "debugging"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "httpbin",
		Branch:        "main",
	},
	{
		Slug:          "eks-simple",
		DisplayName:   "EKS Simple",
		Description:   "A simple Whoami HTTP service on AWS EKS",
		Category:      "architecture",
		Difficulty:    "simple",
		Tags:          []string{"eks", "kubernetes", "alb", "certificate"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "eks-simple",
		Branch:        "main",
	},
	{
		Slug:          "eks-simple-auto",
		DisplayName:   "EKS Simple Auto",
		Description:   "Auto-provisioned EKS cluster with a simple HTTP service",
		Category:      "architecture",
		Difficulty:    "simple",
		Tags:          []string{"eks", "kubernetes", "auto-provision"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "eks-simple-auto",
		Branch:        "main",
	},
	{
		Slug:          "gke-simple",
		DisplayName:   "GKE Simple",
		Description:   "A simple HTTP service on Google Kubernetes Engine",
		Category:      "architecture",
		Difficulty:    "simple",
		Tags:          []string{"gke", "kubernetes"},
		CloudProvider: "gcp",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "gke-simple",
		Branch:        "main",
	},
	{
		Slug:          "grafana",
		DisplayName:   "Grafana",
		Description:   "Grafana observability stack on AWS EKS",
		Category:      "self-hosted",
		Difficulty:    "medium",
		Tags:          []string{"eks", "kubernetes", "helm", "observability"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "grafana",
		Branch:        "main",
	},
	{
		Slug:          "mattermost",
		DisplayName:   "Mattermost",
		Description:   "Mattermost team messaging on AWS EKS",
		Category:      "self-hosted",
		Difficulty:    "medium",
		Tags:          []string{"eks", "kubernetes", "helm", "messaging"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "mattermost",
		Branch:        "main",
	},
	{
		Slug:          "coder",
		DisplayName:   "Coder",
		Description:   "Coder cloud development environments on AWS EKS",
		Category:      "self-hosted",
		Difficulty:    "medium",
		Tags:          []string{"eks", "kubernetes", "helm", "dev-environments"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "coder",
		Branch:        "main",
	},
	{
		Slug:          "twenty",
		DisplayName:   "Twenty CRM",
		Description:   "Twenty open-source CRM on AWS EKS",
		Category:      "self-hosted",
		Difficulty:    "medium",
		Tags:          []string{"eks", "kubernetes", "helm", "crm"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "twenty",
		Branch:        "main",
	},
	{
		Slug:          "penpot",
		DisplayName:   "Penpot",
		Description:   "Penpot design platform on AWS EKS",
		Category:      "self-hosted",
		Difficulty:    "medium",
		Tags:          []string{"eks", "kubernetes", "helm", "design"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "penpot",
		Branch:        "main",
	},
	{
		Slug:          "baserow",
		DisplayName:   "Baserow",
		Description:   "Baserow no-code database on AWS EKS",
		Category:      "self-hosted",
		Difficulty:    "medium",
		Tags:          []string{"eks", "kubernetes", "helm", "no-code"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "baserow",
		Branch:        "main",
	},
	{
		Slug:          "clickhouse",
		DisplayName:   "ClickHouse",
		Description:   "ClickHouse analytics database on AWS EKS",
		Category:      "self-hosted",
		Difficulty:    "medium",
		Tags:          []string{"eks", "kubernetes", "helm", "analytics"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "clickhouse",
		Branch:        "main",
	},
	{
		Slug:          "clickhouse-tailscale",
		DisplayName:   "ClickHouse + Tailscale",
		Description:   "ClickHouse analytics with Tailscale private networking on AWS EKS",
		Category:      "self-hosted",
		Difficulty:    "medium",
		Tags:          []string{"eks", "kubernetes", "helm", "analytics", "networking"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "clickhouse-tailscale",
		Branch:        "main",
	},
	{
		Slug:          "aws-lambda",
		DisplayName:   "AWS Lambda",
		Description:   "Serverless function deployment on AWS Lambda",
		Category:      "architecture",
		Difficulty:    "simple",
		Tags:          []string{"lambda", "serverless", "terraform"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "aws-lambda",
		Branch:        "main",
	},
	{
		Slug:          "ecs-simple",
		DisplayName:   "ECS Simple",
		Description:   "A simple container service on AWS ECS",
		Category:      "architecture",
		Difficulty:    "simple",
		Tags:          []string{"ecs", "docker", "terraform"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "ecs-simple",
		Branch:        "main",
	},
	{
		Slug:          "cockroachdb",
		DisplayName:   "CockroachDB",
		Description:   "CockroachDB distributed SQL database on AWS EKS",
		Category:      "self-hosted",
		Difficulty:    "advanced",
		Tags:          []string{"eks", "kubernetes", "helm", "database"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "cockroachdb",
		Branch:        "main",
	},
	{
		Slug:          "uptime-monitor",
		DisplayName:   "Uptime Monitor",
		Description:   "Uptime monitoring service on AWS",
		Category:      "architecture",
		Difficulty:    "simple",
		Tags:          []string{"monitoring", "terraform"},
		CloudProvider: "aws",
		Repo:          "https://github.com/nuonco/example-app-configs",
		Directory:     "uptime-monitor",
		Branch:        "main",
	},
}

func getExampleAppBySlug(slug string) *ExampleApp {
	for _, app := range exampleAppsCatalog {
		if app.Slug == slug {
			return &app
		}
	}
	return nil
}
