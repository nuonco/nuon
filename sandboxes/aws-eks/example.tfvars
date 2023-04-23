nuon_id = "28g232g0an6vh29t6mu7kv3a96"
region  = "us-east-2"
# assume_role_arn  = "arn:aws:iam::949309607565:role/nuon-demo-install-access"
assume_role_arn  = "" # This needs to be set for use from our services but should not be when running locally.
install_role_arn = "arn:aws:iam::618886478608:role/install-k8s-admin-stage"
tags = {
  nuon_id              = "28g232g0an6vh29t6mu7kv3a96"
  nuon_install_id      = "28g232g0an6vh29t6mu7kv3a96"
  nuon_app_id          = "1mqyl4egjsebw2dwap2s66r69x"
  nuon_sandbox_name    = "aws-eks"
  nuon_sandbox_version = "0.10.4"
}
waypoint_odr_namespace            = "02w56dgii30zc23fb4kvqz77yi"
waypoint_odr_service_account_name = "waypoint-odr-02w56dgii30zc23fb4kvqz77yi"
