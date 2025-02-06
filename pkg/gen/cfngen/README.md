## CloudFormation Install Template Generator

This directory contains the generator that takes a combination of vendor-provided configuration options for an app with internally-derived values (like IDs for a runner and installer) and generates a CloudFormation template that the end customer can use to set up a stack for an install of that app.

To test locally, use `nuonctl scripts exec generate-install-quicklink`, which will invoke the binary in the ./cmd/cfngen subdirectory and take a prototype TOML config that covers all the options for the generator. The nuonctl script will present you with a URL that can be used to test the generated CloudFormation template.