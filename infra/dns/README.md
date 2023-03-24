# infra-nuon-dns

Nuon DNS Configuration as Code

## About

This houses the main DNS configuration for the `nuon.co` domain.
Any top level changes should happen here. Subdomains per environment / account
are delegated to the respective account.

Philosophically, DNS changes should all be automated / tracked in VCS. There
should be no manual changes.
