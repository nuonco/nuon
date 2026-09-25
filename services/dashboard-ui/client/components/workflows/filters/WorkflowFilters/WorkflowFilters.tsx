import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import {
  WORKFLOW_DATE_LABELS,
  WORKFLOW_PREVIEW_LABELS,
  WORKFLOW_STATUS_LABELS,
  WORKFLOW_TYPE_LABELS,
  workflowStatusOptions,
  workflowTypeOptions,
  type TWorkflowDatePreset,
  type TWorkflowOwner,
  type TWorkflowPreviewOption,
  type TWorkflowStatusOption,
  type TWorkflowTypeGroup,
} from '@/utils/workflow-filters'

export interface IWorkflowFilters {
  owner: TWorkflowOwner
  search: string
  status: Set<TWorkflowStatusOption>
  type: Set<TWorkflowTypeGroup>
  date: Set<TWorkflowDatePreset>
  preview: Set<TWorkflowPreviewOption>
  onSearchChange: (value: string) => void
  onStatusChange: (value: Set<TWorkflowStatusOption>) => void
  onTypeChange: (value: Set<TWorkflowTypeGroup>) => void
  onDateChange: (value: Set<TWorkflowDatePreset>) => void
  onPreviewChange: (value: Set<TWorkflowPreviewOption>) => void
}

export const WorkflowFilters = ({
  owner,
  search,
  status,
  type,
  date,
  preview,
  onSearchChange,
  onStatusChange,
  onTypeChange,
  onDateChange,
  onPreviewChange,
}: IWorkflowFilters) => (
  <div className="flex min-w-0 flex-wrap items-center gap-2">
    <SearchInput
      aria-label="Search activity"
      placeholder="Search by title or ID"
      value={search}
      onChange={onSearchChange}
      onClear={() => onSearchChange('')}
    />
    <CheckboxFilterDropdown
      id="workflow-filter-status"
      label="Status"
      options={workflowStatusOptions().map((value) => ({
        value,
        label: WORKFLOW_STATUS_LABELS[value],
      }))}
      selected={status}
      onChange={(value) => onStatusChange(value as Set<TWorkflowStatusOption>)}
    />
    <CheckboxFilterDropdown
      id="workflow-filter-type"
      label="Type"
      options={workflowTypeOptions(owner).map((value) => ({
        value,
        label: WORKFLOW_TYPE_LABELS[value],
      }))}
      selected={type}
      onChange={(value) => onTypeChange(value as Set<TWorkflowTypeGroup>)}
    />
    {owner === 'app' ? (
      <CheckboxFilterDropdown
        id="workflow-filter-preview"
        label="Preview"
        options={(
          Object.keys(WORKFLOW_PREVIEW_LABELS) as TWorkflowPreviewOption[]
        ).map((value) => ({
          value,
          label: WORKFLOW_PREVIEW_LABELS[value],
        }))}
        selected={preview}
        onChange={(value) =>
          onPreviewChange(value as Set<TWorkflowPreviewOption>)
        }
      />
    ) : null}
    <CheckboxFilterDropdown
      id="workflow-filter-date"
      label="Date"
      options={(Object.keys(WORKFLOW_DATE_LABELS) as TWorkflowDatePreset[]).map(
        (value) => ({
          value,
          label: WORKFLOW_DATE_LABELS[value],
        })
      )}
      selected={date}
      onChange={(value) => onDateChange(value as Set<TWorkflowDatePreset>)}
    />
  </div>
)
