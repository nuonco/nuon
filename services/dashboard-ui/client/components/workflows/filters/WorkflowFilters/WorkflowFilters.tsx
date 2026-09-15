import type { ChangeEvent, MouseEvent } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
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

interface IFilterOption {
  label: string
  value: string
}

interface IWorkflowFilterDropdown {
  id: string
  label: string
  options: IFilterOption[]
  selected: Set<string>
  onChange: (selected: Set<string>) => void
}

const WorkflowFilterDropdown = ({
  id,
  label,
  options,
  selected,
  onChange,
}: IWorkflowFilterDropdown) => {
  const handleToggle = (event: ChangeEvent<HTMLInputElement>) => {
    const next = new Set(selected)
    if (event.target.checked) next.add(event.target.value)
    else next.delete(event.target.value)
    onChange(next)
  }

  const handleOnly = (event: MouseEvent<HTMLButtonElement>) => {
    const value = event.currentTarget.value
    onChange(
      selected.size === 1 && selected.has(value) ? new Set() : new Set([value])
    )
  }

  return (
    <Dropdown
      alignment="right"
      closeOnBlur={false}
      id={`workflow-filter-${id}`}
      buttonText={
        <>
          <Icon variant="FunnelIcon" size={14} />
          {label}
          {selected.size ? ` (${selected.size})` : ''}
        </>
      }
    >
      <Menu className="min-w-56 max-h-80 overflow-y-auto">
        {options.map((option) => (
          <CheckboxInputWithButton
            key={option.value}
            buttonProps={{
              children: (
                <>
                  <span className="font-semibold text-xs">{option.label}</span>
                  <span className="ml-2 text-xs opacity-0 group-hover:opacity-100">
                    {selected.size === 1 && selected.has(option.value)
                      ? 'Reset'
                      : 'Only'}
                  </span>
                </>
              ),
              onClick: handleOnly,
              type: 'button',
              value: option.value,
            }}
            checked={selected.has(option.value)}
            className="w-full"
            name={`${id}-${option.value}`}
            onChange={handleToggle}
            value={option.value}
          />
        ))}
        <hr />
        <Button
          disabled={selected.size === 0}
          isMenuButton
          onClick={() => onChange(new Set())}
          type="button"
          variant="ghost"
        >
          Clear
          <Icon variant="XIcon" />
        </Button>
      </Menu>
    </Dropdown>
  )
}

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
    <WorkflowFilterDropdown
      id="status"
      label="Status"
      options={workflowStatusOptions().map((value) => ({
        value,
        label: WORKFLOW_STATUS_LABELS[value],
      }))}
      selected={status}
      onChange={(value) => onStatusChange(value as Set<TWorkflowStatusOption>)}
    />
    <WorkflowFilterDropdown
      id="type"
      label="Type"
      options={workflowTypeOptions(owner).map((value) => ({
        value,
        label: WORKFLOW_TYPE_LABELS[value],
      }))}
      selected={type}
      onChange={(value) => onTypeChange(value as Set<TWorkflowTypeGroup>)}
    />
    {owner === 'app' ? (
      <WorkflowFilterDropdown
        id="preview"
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
    <WorkflowFilterDropdown
      id="date"
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
