provider "nuon" {
  alias  = "sandbox"
  org_id = var.sandbox_org_id
}

provider "nuon" {
  org_id = var.org_id
}
