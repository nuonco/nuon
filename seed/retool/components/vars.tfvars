nested = {
  name   = "{{.nuon.install.id}}"

  region = "{{.nuon.install_stack.outputs.region}}"
  tags = {
    Environment = "{{.nuon.install.id}}"
    Team        = "platform"
    CostCenter  = "12345"
  }
  resources = {
    instances = [
      {
        name      = "app-server-1"
        size      = "t3.medium"
        disk_size = 100
      },
      {
        name      = "app-server-2"
        size      = "t3.medium"
        disk_size = 100
      }
    ]
    networking = {
      vpc_cidr    = "10.0.0.0/16"
      subnets = {
        public-1  = "10.0.1.0/24"
        public-2  = "10.0.2.0/24"
        private-1 = "10.0.3.0/24"
        private-2 = "10.0.4.0/24"
      }
      enable_nat = true
    }
  }
}

