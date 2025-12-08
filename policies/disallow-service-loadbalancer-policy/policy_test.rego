package policy

review_service_clusterip = {"review": {"object": {"spec": {"type": "ClusterIp"}}}}

review_service_loadbalancer = {"review": {"object": {"spec": {"type": "LoadBalancer"}}}}

test_accept if {
	r = review_service_clusterip
	res = violation with input as r
	count(res) = 0
}

test_reject if {
	r = review_service_loadbalancer
	res = violation with input as r
	count(res) = 1
}
