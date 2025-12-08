package policy

test_reject if {
	r = {"review": {"kind": {"kind": "PodExecOptions"}}}
	res = violation with input as r
	count(res) = 1
}

test_accept if {
	r = {"review": {"kind": {"kind": "Pod"}}}
	res = violation with input as r
	count(res) = 0
}
