package policy

violation = [res |
    msg := get_violation_msg(input.review.object)
    res := {"msg": msg}
]

get_violation_msg(obj) = "Role reading secrets is not allowed" if {
    obj.kind == "Role"
    rule := obj.rules[_]
    has_secret_access(rule)
}

get_violation_msg(obj) = "ClusterRole reading secrets is not allowed" if {
    obj.kind == "ClusterRole"
    rule := obj.rules[_]
    has_secret_access(rule)
}

has_secret_access(rule) if {
  resource_match(rule.resources)
  verb_match(rule.verbs)
}

resource_match(resources) if {
  resources[_] == "secrets"
}

resource_match(resources) if {
  resources[_] == "*"
}

verb_match(verbs) if {
  verbs[_] == "get"
}

verb_match(verbs) if {
  verbs[_] == "list"
}

verb_match(verbs) if {
  verbs[_] == "watch"
}

verb_match(verbs) if {
  verbs[_] == "*"
}
