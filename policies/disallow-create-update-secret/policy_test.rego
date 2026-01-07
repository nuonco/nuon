package policy

review_secret = {"review": {"object": {"kind": "Secret"}}}

review_pod = {"review": {"object": {"kind": "Pod"}}}

test_accept if {
	r = review_pod
	res = violation with input as r
	count(res) = 0
}

test_reject if {
	r = review_secret
	res = violation with input as r
	count(res) = 1
}
