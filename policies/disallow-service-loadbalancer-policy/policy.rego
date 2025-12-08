package policy

violation = [ {"msg": "Service of type LoadBalancer are not allowed"} |
    input.review.object.spec.type == "LoadBalancer"
]

