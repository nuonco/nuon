output "db" {
  value =  {
    instance = {
      name = "ctl-api"
      host = module.primary.db_instance_address
      port = module.primary.db_instance_port
      username = "ctl_api"
    }

    admin = {
      name = module.primary.db_instance_name
      username = nonsensitive(module.primary.db_instance_username)
      # NOTE: this password is only used once, and then removed.
      password = nonsensitive(module.primary.db_instance_password)
    }
  }
}
