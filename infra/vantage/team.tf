resource "vantage_team" "nuon_engineering" {
  name = "Nuon Engineering"
  description = "Everyone on the nuon engineering team."
  # adding jon@nuon.co is giving us grief - contact support
  user_emails = ["nat@nuon.co", "jordan@nuon.co", "sam@nuon.co", "fred@nuon.com", "rob@nuon.co"]
  workspace_tokens = ["wrkspc_217372b8dbd3cc61"]
}
