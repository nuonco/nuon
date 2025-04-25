resource "pagerduty_service_dependency" "controlplane-deps" {
  dependency {
    dependent_service {
      id   = pagerduty_business_service.controlplane.id
      type = "business_service"
    }
    supporting_service {
      id   = pagerduty_service.ctl-api.id
      type = "service"
    }
  }
}

resource "pagerduty_service_dependency" "dashboard-deps" {
  dependency {
    dependent_service {
      id   = pagerduty_business_service.vendor-dashboard.id
      type = "business_service"
    }
    supporting_service {
      id   = pagerduty_service.dashboard-ui.id
      type = "service"
    }
  }
}

resource "pagerduty_service_dependency" "installer-deps" {
  dependency {
    dependent_service {
      id   = pagerduty_business_service.installers.id
      type = "business_service"
    }
    supporting_service {
      id   = pagerduty_service.installer.id
      type = "service"
    }
  }
}

resource "pagerduty_service_dependency" "runner-deps" {
  dependency {
    dependent_service {
      id   = pagerduty_business_service.runners.id
      type = "business_service"
    }
    supporting_service {
      id   = pagerduty_service.runner.id
      type = "service"
    }
  }
}


resource "pagerduty_service_dependency" "website-deps" {
  dependency {
    dependent_service {
      id   = pagerduty_business_service.website.id
      type = "business_service"
    }
    supporting_service {
      id   = pagerduty_service.website.id
      type = "service"
    }
  }
}


resource "pagerduty_service_dependency" "dashboard-ctl-api" {
  dependency {
    dependent_service {
      id   = pagerduty_service.dashboard-ui.id
      type = "service"
    }
    supporting_service {
      id   = pagerduty_service.ctl-api.id
      type = "service"
    }

  }
}

resource "pagerduty_service_dependency" "installer-ctl-api" {
  dependency {
    dependent_service {
      id   = pagerduty_service.installer.id
      type = pagerduty_service.installer.type
    }
    supporting_service {
      id   = pagerduty_service.ctl-api.id
      type = pagerduty_service.ctl-api.type
    }

  }
}


resource "pagerduty_service_dependency" "runner-ctl-api" {
  dependency {
    dependent_service {
      id   = pagerduty_service.runner.id
      type = pagerduty_service.runner.type
    }
    supporting_service {
      id   = pagerduty_service.ctl-api.id
      type = pagerduty_service.ctl-api.type
    }
  }
}

