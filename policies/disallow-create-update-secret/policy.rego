package policy

violation = [ {"msg": "Creating or Updating Secrets is not allowed"} |
    input.review.object.kind == "Secret"
]
