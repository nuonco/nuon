export default {
  title: 'Workflows/WorkflowDetails/WorkflowStatusSection',
}

import { WorkflowStatusSection } from './WorkflowStatusSection'

const makeWorkflow = (status: string, description?: string, type = 'deploy') =>
  ({
    id: 'wf-123',
    type,
    status: {
      status,
      status_human_description: description ?? status,
    },
  }) as any

export const Running = () => (
  <div className="max-w-2xl p-4">
    <WorkflowStatusSection workflow={makeWorkflow('in-progress', 'In progress')} />
  </div>
)

export const Succeeded = () => (
  <div className="max-w-2xl p-4">
    <WorkflowStatusSection workflow={makeWorkflow('success', 'Succeeded')} />
  </div>
)

export const Failed = () => (
  <div className="max-w-2xl p-4">
    <WorkflowStatusSection workflow={makeWorkflow('error', 'Failed')} />
  </div>
)

export const InputUpdate = () => (
  <div className="max-w-2xl p-4">
    <WorkflowStatusSection
      workflow={makeWorkflow('success', 'Succeeded', 'input_update')}
    />
  </div>
)
