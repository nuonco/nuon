package policy

violation = [ {"msg": "Creating Roles or ClusterRoles is not allowed"} |
    kinds := {"Role", "ClusterRole"}
    kinds[input.review.object.kind]
]
