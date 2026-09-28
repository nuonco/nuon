export default {
  title: 'Features / Workflows / Filters / Type filter',
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
