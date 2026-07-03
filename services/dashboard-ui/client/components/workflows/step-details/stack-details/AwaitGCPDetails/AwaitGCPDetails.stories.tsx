export default {
  title: 'Workflows/StepDetails/AwaitGCPDetails',
}

import { AwaitGCPDetails, AwaitGCPDetailsSkeleton } from './AwaitGCPDetails'

const mockStack = {
  versions: [
    {
      contents: JSON.stringify({
        inputs_tfvars:
          'nuon_install_id = "install-1"\ninstall_inputs = {\n  "cluster_name" = ""\n}\n',
        secrets_tfvars:
          'auto_generate_secrets = ["db_password", ]\nsecrets = {\n  "stripe_key" = {\n    description = "Your Stripe API key"\n    required    = true\n    value       = ""\n  }\n}\n',
      }),
    },
  ],
} as any

const mockStep = {
  id: 'step-1',
  status: { status: 'active' },
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <AwaitGCPDetails stack={mockStack} step={mockStep} installId="install-1" />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitGCPDetailsSkeleton />
  </div>
)
