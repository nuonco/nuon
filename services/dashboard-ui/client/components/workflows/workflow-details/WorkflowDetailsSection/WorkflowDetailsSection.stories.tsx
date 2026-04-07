export default {
  title: 'Workflows/WorkflowDetails/WorkflowDetailsSection',
}

import { WorkflowDetailsSection } from './WorkflowDetailsSection'

const mockWorkflow = {
  id: 'wf-123',
  type: 'deploy',
  name: 'Deploy workflow',
  created_at: new Date(Date.now() - 60000).toISOString(),
  created_by: { email: 'user@example.com' },
  metadata: {},
} as any

const mockInstall = {
  app_id: 'app-123',
  app: { name: 'My App' },
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <WorkflowDetailsSection
      workflow={mockWorkflow}
      orgId="org-123"
      install={mockInstall}
    />
  </div>
)

export const WithoutInstall = () => (
  <div className="max-w-2xl p-4">
    <WorkflowDetailsSection
      workflow={mockWorkflow}
      orgId="org-123"
    />
  </div>
)

export const WithChangedInputs = () => (
  <div className="max-w-2xl p-4">
    <WorkflowDetailsSection
      workflow={{
        ...mockWorkflow,
        type: 'input_update',
        metadata: {
          changed_input_values: JSON.stringify({
            region: { old: 'us-east-1', new: 'us-west-2' },
            instance_type: { old: 't3.small', new: 't3.medium' },
          }),
        },
      }}
      orgId="org-123"
      install={mockInstall}
    />
  </div>
)
