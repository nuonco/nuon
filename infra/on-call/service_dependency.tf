resource "pagerduty_service_dependency" "controlplane-deps" {
   dependency {
        dependent_service {
            id = pagerduty_business_service.controlplane.id
            type = pagerduty_business_service.controlplane.type
        }
        supporting_service {
            id = pagerduty_service.ctl-api.id
            type = pagerduty_service.ctl-api.type
        }
    }
}

resource "pagerduty_service_dependency" "dashboard-deps" {
    dependency {
        dependent_service {
            id = pagerduty_business_service.vendor-dashbaord.id
            type = pagerduty_business_service.vendor-dashbaord.type
        }
        supporting_service {
            id = pagerduty_service.dashboard-ui.id
            type = pagerduty_service.dashboard-ui.type
        }
    }
}

resource "pagerduty_service_dependency" "installer-deps" {
    dependency {
        dependent_service {
            id = pagerduty_business_service.installers.id
            type = pagerduty_business_service.installers.type
        }
        supporting_service {
            id = pagerduty_service.installer.id
            type = pagerduty_service.installer.type
        }
    }
}

resource "pagerduty_service_dependency" "runner-deps" {
    dependency {
        dependent_service {
            id = pagerduty_business_service.runners.id
            type = pagerduty_business_service.runners.type
        }
        supporting_service {
            id = pagerduty_service.runner.id
            type = pagerduty_service.runner.type
        }
    }
}


resource "pagerduty_service_dependency" "website-deps" {
    dependency {
        dependent_service {
            id = pagerduty_business_service.website.id
            type = pagerduty_business_service.website.type
        }
        supporting_service {
            id = pagerduty_service.website.id
            type = pagerduty_service.website.type
        }
    }
}


resource "pagerduty_service_dependency" "dashboard-ctl-api" {
   dependency {
        dependent_service {
            id = pagerduty_service.dashboard-ui.id
            type = pagerduty_service.dashboard-ui.type
        }
        supporting_service {
            id = pagerduty_service.ctl-api.id
            type = pagerduty_service.ctl-api.type
        }

    }
}

resource "pagerduty_service_dependency" "installer-ctl-api" {
   dependency {
        dependent_service {
            id = pagerduty_service.installer.id
            type = pagerduty_service.installer.type
        }
        supporting_service {
            id = pagerduty_service.ctl-api.id
            type = pagerduty_service.ctl-api.type
        }

    }
}


resource "pagerduty_service_dependency" "runner-ctl-api" {
   dependency {
        dependent_service {
            id = pagerduty_service.runner.id
            type = pagerduty_service.runner.type
        }
        supporting_service {
            id = pagerduty_service.ctl-api.id
            type = pagerduty_service.ctl-api.type
        }
    }
}

