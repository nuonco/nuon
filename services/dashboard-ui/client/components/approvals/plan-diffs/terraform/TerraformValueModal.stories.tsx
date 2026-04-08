export default {
  title: 'Approvals/PlanDiffs/TerraformValueModal',
}

import { TerraformValueModal } from './TerraformValueModal'

export const After = () => (
    <TerraformValueModal
      valueKey="instance_type"
      value={JSON.stringify({ type: 't3.small', vcpus: 2, memory_gb: 2 }, null, 2)}
    />
  )

export const Before = () => (
    <TerraformValueModal
      isBefore
      valueKey="instance_type"
      value={JSON.stringify({ type: 't3.micro', vcpus: 1, memory_gb: 1 }, null, 2)}
    />
  )
