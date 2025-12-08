package policy

review_role = {"review": {"object": {"kind": "Role"}}}
review_clusterrole = {"review": {"object": {"kind": "ClusterRole"}}}
review_pod = {"review": {"object": {"kind": "Pod"}}}

test_reject_role if {
	r = review_role
	res = violation with input as r
	count(res) = 1
}

test_reject_clusterrole if {
	r = review_clusterrole
	res = violation with input as r
	count(res) = 1
}

test_accept_pod if {
	r = review_pod
	res = violation with input as r
	count(res) = 0
}
