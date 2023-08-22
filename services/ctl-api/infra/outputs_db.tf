output "db" {
  value =  {
    instance = {
      name = "ctl-api"
      host = module.primary.db_instance_address
      port = module.primary.db_instance_port
      username = "ctl-api"
    }

    admin = {
      name = module.primary.db_instance_name
      username = nonsensitive(module.primary.db_instance_username)
    }
  }
}
