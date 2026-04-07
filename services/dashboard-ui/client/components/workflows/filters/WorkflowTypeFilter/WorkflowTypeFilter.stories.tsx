export default {
  title: 'Workflows/Filters/WorkflowTypeFilter',
}

import { WorkflowTypeFilter } from './WorkflowTypeFilter'

export const Default = () => (
  <div className="p-4">
    <WorkflowTypeFilter
      workflowType=""
      onTypeChange={() => {}}
      onClearFilter={() => {}}
    />
  </div>
)
