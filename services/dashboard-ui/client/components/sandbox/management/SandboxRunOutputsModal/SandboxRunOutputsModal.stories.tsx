export default {
  title: 'Sandbox/Management/SandboxRunOutputsModal',
}

import { SandboxRunOutputsButton } from './SandboxRunOutputsModal'

const mockSandboxRun = {
  id: 'run-1',
  outputs: {
    vpc_id: 'vpc-12345',
    subnet_ids: ['subnet-1', 'subnet-2'],
  },
} as any

export const Default = () => (
  <div className="p-4">
    <SandboxRunOutputsButton
      sandboxRun={mockSandboxRun}
      onOpen={() => {}}
    />
  </div>
)
