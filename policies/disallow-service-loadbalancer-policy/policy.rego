package nuon

deny = [{"msg": "Service of type LoadBalancer are not allowed"} |
	input.review.object.spec.type == "LoadBalancer"
]
