export default {
  title: 'Workflows/StepDetails/AwaitAWSDetails',
}

import { AwaitAWSDetails, AwaitAWSDetailsSkeleton } from './AwaitAWSDetails'

const mockStack = {
  versions: [
    {
      template_url: 'https://s3.amazonaws.com/bucket/template.json',
      quick_link_url: 'https://console.aws.amazon.com/cloudformation/home?stackName=nuon-stack&region=us-east-1',
      region: 'us-east-1',
    },
  ],
} as any

const mockInputsTfvars = `nuon_install_id        = "inl4xabsyaqxp0cb2oy5l8urvf"
nuon_org_id            = "orgnwi4odoca7y0z9wddc1767e"
nuon_app_id            = "appk2o58477kw8jbounuxpkaqr"
aws_region             = "us-east-1"
runner_api_url         = "https://api.nuon.co/runner"
runner_id              = "run4dbg9i5fzwdlq7zk1llbout"
phone_home_url         = "https://api.nuon.co/v1/installs/inl4xabsyaqxp0cb2oy5l8urvf/phone-home/aws3no0qz8sxsbqa13dgs2pfb3"
install_inputs = {
  "cluster_name" = ""
}
`

const mockSecretsTfvars = `secrets = {
  "stripe_key" = {
    description = "Your Stripe API key"
    required    = true
    value       = ""
  }
}

auto_generate_secrets = ["db_password", ]
`

const mockStackWithBoth = {
  versions: [
    {
      ...mockStack.versions[0],
      terraform_contents: JSON.stringify({
        inputs_tfvars: mockInputsTfvars,
        provider_tfvars:
          'api_url       = "https://api.nuon.co"\nphone_home_id = "ph-1"\n\naws = {\n  region = "us-east-1"\n}\n\ninstall_inputs = {\n  "cluster_name" = ""\n}\n',
        secrets_tfvars: mockSecretsTfvars,
      }),
      terraform_checksum: 'sha256-abc',
    },
  ],
} as any

const mockStackLegacy = {
  versions: [
    {
      ...mockStack.versions[0],
      terraform_contents: JSON.stringify({
        tfvars: `nuon_install_id        = "inl4xabsyaqxp0cb2oy5l8urvf"
nuon_org_id            = "orgnwi4odoca7y0z9wddc1767e"
nuon_app_id            = "appk2o58477kw8jbounuxpkaqr"
aws_region             = "us-east-1"
install_inputs = {
}
auto_generate_secrets = []
secrets = {
  "stripe_key" = {
    description = "Your Stripe API key"
    required    = true
    value       = ""
  }
}
`,
      }),
      terraform_checksum: 'sha256-legacy',
    },
  ],
} as any

const mockStep = {
  id: 'step-1',
  status: { status: 'active' },
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <AwaitAWSDetails
      stack={mockStack}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
    />
  </div>
)

export const WithBothOptions = () => (
  <div className="max-w-2xl p-4">
    <AwaitAWSDetails
      stack={mockStackWithBoth}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
    />
  </div>
)

const mockStackNoBucket = {
  versions: [
    {
      terraform_contents: mockStackWithBoth.versions[0].terraform_contents,
      terraform_checksum: 'sha256-abc',
    },
  ],
} as any

export const NoTemplateBucket = () => (
  <div className="max-w-2xl p-4">
    <AwaitAWSDetails
      stack={mockStackNoBucket}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
    />
  </div>
)

export const TerraformProvider = () => (
  <div className="max-w-2xl p-4">
    <AwaitAWSDetails
      stack={mockStackWithBoth}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
      tfProvider
    />
  </div>
)

export const LegacyContents = () => (
  <div className="max-w-2xl p-4">
    <AwaitAWSDetails
      stack={mockStackLegacy}
      step={mockStep}
      orgId="org-1"
      installId="install-1"
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitAWSDetailsSkeleton />
  </div>
)
