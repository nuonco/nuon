package cmd

import (
	"github.com/powertoolsdev/go-common/config"
	workers "github.com/powertoolsdev/workers-orgs/internal"
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_namespace", "default")

	// org defaults
	config.RegisterDefault("org.waypoint_server_root_domain", "orgs-stage.nuon.co")
	config.RegisterDefault("org.waypoint_bootstrap_token_namespace", "default")
	config.RegisterDefault("org.bucket", "nuon-installations-stage")
	config.RegisterDefault("org.bucket_region", "us-west-2")
	config.RegisterDefault("org.role_arn", "arn:aws:iam::618886478608:role/install-k8s-admin-stage")
	config.RegisterDefault("org.orgs_k8s_cluster_id", "orgs-stage-main")
	config.RegisterDefault("org.orgs_k8s_role_arn", "arn:aws:iam::766121324316:role/extra-auth-eks-workers-orgs-orgs-stage-main")
	config.RegisterDefault("org.orgs_k8s_ca_data", "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1UQXlPREU0TURNME5Gb1hEVE15TVRBeU5URTRNRE0wTkZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTWlKCmc4ZEprQXlqb2JkeTdKQmhyQ2dpVzhOOTRnWW0wQVkrcks0aHJzZ1FBV1c5bUJmQm1xL05sZXpUSGY4Yng2WVYKa1gvdkhUU1I5QlRvOUpITGM0ME9EM05GaXpibGdTMFh6U3BPeE10TDBLeVFMbk5pVlBMRTZPU1dPN09uUjFOSApiWjF1T3M5VFNMU0trUkZHK21VVVZmZndaQ1YyTG81V0JWWFg4Q2JwaWhnRkU0U2NNc0dydmRjem9OVzlsMk16CkxrTFFJcE9GSmFXbHd3TlBmZnZzSXdJR1ZIaTBLd1lXaDFzbDM5azBUb3NNbXFaSW5oWWVabkUrYmg0NFh5WFMKOTVPbFFYbnpzd3U1TjY2MGJSQUJFVnJSOG1iNUd0Q3h4T1dXVjcrZHRiQU5jdEQvUlhlRUJqSzNyb3ZycXFDZQo2ZlJ1eWQ1RU1BUHgwQW5talprQ0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZPV0VmbC8va3c4Q2Z6NWo4SjVJalZkWkhPeTlNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBSldUaFBlS2pmOVZyQ2dlbnVDNgpLeUk4cXBzSXNxbG9JVEFueG00NENlYjR5bTJDY0hnQ2tNSEpCczhQdURHeGhldU9FOXJ5TTk4SU9SOUpFVHU0CmUwNWsvSUhQeFFGMWk5eldqcjIrREx3QjZnbGx5TGFQbEVCTXk2NTE4V1JpOStUM3ZnZGRiektyNnJTSnpISEsKajNMV0dJM1FQVVVqZEEvYXVoRElvVWdGcDJPOFFmRUI4UG96N0QrQVNmMkdyZEw4SlN2eTdsb0NxNW04RGs1VQpZRDNYY0JhRGMwd0F6T2xua1pyWWYzWWZFMXRwbzNGL3prbU9sOWRoemZhVWdZN0hOb0prOERSRjQ5aXRYclVsCjk2L1QvbzhWdjRHYUxqVHhkUUgyM0oxTzJoTExhRXdmOEQ2dXE3bG8rT2Z6L3NMTHNCUDNsRzFHcEhwRzBSOEgKVllzPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==")
	config.RegisterDefault("org.orgs_k8s_public_endpoint", "https://891656020B6C382DE7E0D96E1B86D224.gr7.us-west-2.eks.amazonaws.com")
}

type Config struct {
	config.Base       `config:",squash"`
	TemporalHost      string `config:"temporal_host"`
	TemporalNamespace string `config:"temporal_namespace"`

	// NOTE: these webhook urls are scoped at the project level, but are workflow specific. This is because we
	// create a slack notifier object at the cmd level and pass it to each individual workflow
	OrgBotsSlackWebhookURL string `config:"org_bots_slack_webhook_url"`

	// Domain specific configs
	OrgCfg workers.Config `config:"org"`
}
