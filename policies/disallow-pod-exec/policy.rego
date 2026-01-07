package nuon

deny = [{"msg": "Exec into pods is not allowed"} |
	input.review.kind.kind == "PodExecOptions"
]
