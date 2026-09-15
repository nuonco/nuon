import { useState } from 'react'
import { WorkflowFilters } from './WorkflowFilters'
import type {
  TWorkflowDatePreset,
  TWorkflowOwner,
  TWorkflowPreviewOption,
  TWorkflowStatusOption,
  TWorkflowTypeGroup,
} from '@/utils/workflow-filters'

export default {
  title: 'Workflows/Filters/WorkflowFilters',
}

const Example = ({ owner }: { owner: TWorkflowOwner }) => {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState(new Set<TWorkflowStatusOption>())
  const [type, setType] = useState(new Set<TWorkflowTypeGroup>())
  const [date, setDate] = useState(new Set<TWorkflowDatePreset>())
  const [preview, setPreview] = useState(new Set<TWorkflowPreviewOption>())

  return (
    <div className="p-4">
      <WorkflowFilters
        owner={owner}
        search={search}
        status={status}
        type={type}
        date={date}
        preview={preview}
        onSearchChange={setSearch}
        onStatusChange={setStatus}
        onTypeChange={setType}
        onDateChange={setDate}
        onPreviewChange={setPreview}
      />
    </div>
  )
}

export const InstallActivity = () => <Example owner="install" />

export const AppBranchActivity = () => <Example owner="app" />

export const SelectedFilters = () => {
  const [status, setStatus] = useState<Set<TWorkflowStatusOption>>(
    new Set(['running', 'failed'])
  )
  const [type, setType] = useState<Set<TWorkflowTypeGroup>>(
    new Set(['branch-manual'])
  )
  const [date, setDate] = useState<Set<TWorkflowDatePreset>>(new Set(['7d']))
  const [preview, setPreview] = useState<Set<TWorkflowPreviewOption>>(
    new Set(['preview'])
  )

  return (
    <div className="p-4">
      <WorkflowFilters
        owner="app"
        search="deploy"
        status={status}
        type={type}
        date={date}
        preview={preview}
        onSearchChange={() => {}}
        onStatusChange={setStatus}
        onTypeChange={setType}
        onDateChange={setDate}
        onPreviewChange={setPreview}
      />
    </div>
  )
}
