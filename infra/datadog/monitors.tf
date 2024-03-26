resource "datadog_monitor_json" "kubernetes_container_waiting" {
  monitor = <<-EOF
{
	"id": 589567,
	"name": "Kubernetes Container Waiting",
	"type": "query alert",
	"query": "sum(last_2m):sum:kubernetes.containers.state.waiting{(env:stage OR env:prod)} by {kube_container_name,env} > 1",
	"message": "@slack-nuon-alerts-product \n{{kube_container_name.name}} in {{env.name}} has {{value}} containers waiting.",
	"tags": [],
	"options": {
		"thresholds": {
			"critical": 1,
			"warning": 0.5
		},
		"notify_audit": false,
		"include_tags": true,
		"new_group_delay": 60,
		"renotify_interval": 0,
		"escalation_message": "",
		"notify_no_data": false,
		"silenced": {}
	},
	"priority": 2,
	"restricted_roles": null
}
EOF
}

import {
  to = datadog_monitor_json.kubernetes_container_waiting
  id = 589567
}
