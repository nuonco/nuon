import { useState } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Icon } from '../atoms/Icon'
import { Text } from '../atoms/Text'
import { useFilterSelection } from '../../hooks/use-filter-selection'
import {
  FilterDropdown,
  FilterMenu,
  type IFilterMenuOption,
} from './FilterMenu'

export default { title: 'lite/molecules/FilterMenu' }

const OPERATIONS = [
  { value: 'create', label: 'Create', rail: 'bg-diff-add' },
  { value: 'update', label: 'Update', rail: 'bg-diff-change' },
  { value: 'delete', label: 'Delete', rail: 'bg-diff-remove' },
] as const satisfies readonly IFilterMenuOption<string>[]
const OPERATION_VALUES = OPERATIONS.map(({ value }) => value)

const SEVERITIES = [
  { value: 'trace', label: 'Trace', rail: 'bg-brand-tint' },
  { value: 'debug', label: 'Debug', rail: 'bg-tertiary' },
  { value: 'info', label: 'Info', rail: 'bg-status-info' },
  { value: 'warn', label: 'Warn', rail: 'bg-status-warn' },
  { value: 'error', label: 'Error', rail: 'bg-status-error' },
  { value: 'fatal', label: 'Fatal', rail: 'bg-field-invalid' },
] as const satisfies readonly IFilterMenuOption<string>[]
const SEVERITY_VALUES = SEVERITIES.map(({ value }) => value)
const DEFAULT_SEVERITIES = ['info', 'warn', 'error', 'fatal']

const COMPONENT_TYPES = [
  {
    value: 'manifest',
    label: 'Kubernetes manifest',
    description: 'Rendered YAML resources',
    leading: <Icon variant="FileIcon" size={14} />,
  },
  {
    value: 'terraform',
    label: 'Terraform module',
    description: 'Managed infrastructure',
    leading: <Icon variant="FileIcon" size={14} />,
  },
] as const
const COMPONENT_TYPE_VALUES = COMPONENT_TYPES.map(({ value }) => value)

const LONG_OPTIONS = [
  {
    value: 'long',
    label:
      'A very long component type label that must remain inside the filter',
  },
  { value: 'short', label: 'Short option' },
] as const
const LONG_VALUES = LONG_OPTIONS.map(({ value }) => value)

export const Overview = () => (
  <ComponentDocs
    name="FilterMenu"
    tier="molecule"
    summary="A Datadog-style checkable filter: checkbox toggles, row isolates."
    use={[
      'Use FilterDropdown for ordinary multi-select filters.',
      'Use FilterMenu directly when a custom trigger is required.',
      'Pass rail for severity or operation colour.',
    ]}
    avoid={[
      'Do not use MenuItem for filter choices.',
      'Do not hide isolate behind a submenu.',
    ]}
    rules={[
      'Space toggles the focused option; Enter isolates it.',
      'Only and Reset appear on hover and keyboard focus.',
      'Reset restores the declared defaults, not every option.',
      'The trigger count stays hidden while the selection matches the defaults.',
    ]}
    props={[
      {
        name: 'options',
        type: 'IFilterMenuOption[]',
        description:
          'Value, label, description, leading content, rail and disabled state.',
      },
      {
        name: 'selected',
        type: 'Set<string>',
        description: 'Currently included values.',
      },
      {
        name: 'onToggle',
        type: '(value) => void',
        description: 'Checkbox and Space behavior.',
      },
      {
        name: 'onIsolate',
        type: '(value) => void',
        description: 'Row click and Enter behavior.',
      },
      {
        name: 'onReset',
        type: '() => void',
        description: 'Restores the default selection.',
      },
      {
        name: 'isConstrained',
        type: 'boolean',
        description:
          'Overrides the trigger count when defaults are not every option.',
      },
    ]}
  />
)

const OperationDropdown = () => {
  const filter = useFilterSelection(OPERATION_VALUES)
  return (
    <FilterDropdown
      label="Changes"
      options={OPERATIONS}
      selected={filter.selected}
      onToggle={filter.toggle}
      onIsolate={filter.isolate}
      onReset={filter.reset}
    />
  )
}

export const Default = () => (
  <div className="p-20">
    <OperationDropdown />
  </div>
)

const MixedDemo = () => {
  const filter = useFilterSelection(SEVERITY_VALUES, DEFAULT_SEVERITIES)
  return (
    <FilterDropdown
      label="Severity"
      options={SEVERITIES}
      selected={filter.selected}
      onToggle={filter.toggle}
      onIsolate={filter.isolate}
      onReset={filter.reset}
      isConstrained={filter.isConstrained}
    />
  )
}

export const SeverityRails = () => (
  <div className="flex flex-col items-start gap-4 p-20">
    <Text variant="caption" color="tertiary">
      Trace and Debug start off. Reset returns to those defaults, not to every
      option.
    </Text>
    <MixedDemo />
  </div>
)

const IsolatedDemo = () => {
  const [selected, setSelected] = useState(new Set(['update']))
  return (
    <FilterMenu
      label="Change filters"
      options={OPERATIONS}
      selected={selected}
      onToggle={(value) =>
        setSelected((current) => {
          const next = new Set(current)
          if (next.has(value)) next.delete(value)
          else next.add(value)
          return next
        })
      }
      onIsolate={(value) =>
        setSelected((current) =>
          current.size === 1 && current.has(value)
            ? new Set(OPERATIONS.map((option) => option.value))
            : new Set([value])
        )
      }
      onReset={() =>
        setSelected(new Set(OPERATIONS.map((option) => option.value)))
      }
      closeOnReset={false}
    />
  )
}

export const Isolated = () => (
  <div className="max-w-xs p-8">
    <Text variant="caption" color="tertiary" className="mb-3">
      Update is isolated. Hover or focus it to see Reset.
    </Text>
    <IsolatedDemo />
  </div>
)

export const DescriptionsAndLeading = () => {
  const filter = useFilterSelection(COMPONENT_TYPE_VALUES)
  return (
    <div className="p-20">
      <FilterDropdown
        label="Types"
        options={COMPONENT_TYPES}
        selected={filter.selected}
        onToggle={filter.toggle}
        onIsolate={filter.isolate}
        onReset={filter.reset}
      />
    </div>
  )
}

export const Keyboard = () => (
  <div className="flex flex-col items-start gap-3 p-20">
    <Text variant="caption" color="tertiary">
      Open with ArrowDown. Move with arrows, Space toggles, Enter isolates.
    </Text>
    <OperationDropdown />
  </div>
)

export const LongLabels = () => {
  const filter = useFilterSelection(LONG_VALUES)
  return (
    <div className="p-20">
      <FilterDropdown
        label="Types"
        options={LONG_OPTIONS}
        selected={filter.selected}
        onToggle={filter.toggle}
        onIsolate={filter.isolate}
        onReset={filter.reset}
      />
    </div>
  )
}

export const EmptyAndSingle = () => (
  <div className="flex items-start gap-8 p-20">
    <FilterDropdown
      label="Empty"
      options={[]}
      selected={new Set<string>()}
      onToggle={() => {}}
      onIsolate={() => {}}
      onReset={() => {}}
    />
    <FilterDropdown
      label="Provider"
      options={[{ value: 'aws', label: 'AWS' }]}
      selected={new Set(['aws'])}
      onToggle={() => {}}
      onIsolate={() => {}}
      onReset={() => {}}
    />
  </div>
)
