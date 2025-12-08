package policy

test_role_reading_secrets if {
    inp := {
        "review": {
            "object": {
                "kind": "Role",
                "rules": [
                    {
                        "apiGroups": [""],
                        "resources": ["secrets"],
                        "verbs": ["get"]
                    }
                ]
            }
        }
    }
    res := violation with input as inp
    count(res) == 1
}

test_role_not_reading_secrets if {
    inp := {
        "review": {
            "object": {
                "kind": "Role",
                "rules": [
                    {
                        "apiGroups": [""],
                        "resources": ["pods"],
                        "verbs": ["get"]
                    }
                ]
            }
        }
    }
    res := violation with input as inp
    count(res) == 0
}

test_clusterrole_reading_secrets if {
    inp := {
        "review": {
            "object": {
                "kind": "ClusterRole",
                "rules": [
                    {
                        "apiGroups": [""],
                        "resources": ["secrets"],
                        "verbs": ["watch"]
                    }
                ]
            }
        }
    }
    res := violation with input as inp
    count(res) == 1
}

test_wildcard_resources if {
    inp := {
        "review": {
            "object": {
                "kind": "Role",
                "rules": [
                    {
                        "apiGroups": [""],
                        "resources": ["*"],
                        "verbs": ["get"]
                    }
                ]
            }
        }
    }
    res := violation with input as inp
    count(res) == 1
}

test_wildcard_verbs if {
    inp := {
        "review": {
            "object": {
                "kind": "Role",
                "rules": [
                    {
                        "apiGroups": [""],
                        "resources": ["secrets"],
                        "verbs": ["*"]
                    }
                ]
            }
        }
    }
    res := violation with input as inp
    count(res) == 1
}

test_mixed_permissions if {
    inp := {
        "review": {
            "object": {
                "kind": "Role",
                "rules": [
                    {
                        "apiGroups": [""],
                        "resources": ["pods", "secrets"],
                        "verbs": ["*"]
                    }
                ]
            }
        }
    }
    res := violation with input as inp
    count(res) == 1
}
