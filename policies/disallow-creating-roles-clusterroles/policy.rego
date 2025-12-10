package nuon

deny = [{"msg": "Creating Roles or ClusterRoles is not allowed"} |
	kinds := {"Role", "ClusterRole"}
	kinds[input.review.object.kind]
]
