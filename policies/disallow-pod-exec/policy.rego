package policy

violation = [{"msg": "Exec into pods is not allowed"} |
	input.review.kind.kind == "PodExecOptions"
]
