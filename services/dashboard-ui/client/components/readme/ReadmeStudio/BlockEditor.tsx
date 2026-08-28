import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
import type {
  TActionBlock,
  TArraySource,
  TBlock,
  TComponentBlock,
  TConditionOp,
  TEntityOption,
  TRunbookBlock,
  TStateVariable,
  TValueKind,
} from './types'

const kindOptions = [
  { value: 'text', label: 'Text' },
  { value: 'status', label: 'Status badge' },
  { value: 'time', label: 'Relative time' },
]

const conditionOptions = [
  { value: '', label: 'Always show' },
  { value: 'exists', label: 'Show when value exists' },
  { value: 'not-exists', label: 'Show when value is missing' },
  { value: 'eq', label: 'Show when value equals' },
  { value: 'ne', label: 'Show when value differs' },
]

const themeOptions = [
  { value: 'info', label: 'Info' },
  { value: 'warn', label: 'Warning' },
  { value: 'error', label: 'Error' },
  { value: 'success', label: 'Success' },
]

interface IBlockEditor {
  block: TBlock
  variables: TStateVariable[]
  arraySources: TArraySource[]
  runbooks?: TEntityOption[]
  actions?: TEntityOption[]
  components?: TEntityOption[]
  onChange: (block: TBlock) => void
}

const entityEditorMeta = {
  runbook: { label: 'Runbook', empty: 'No runbooks in this app yet' },
  action: { label: 'Action', empty: 'No actions in this app yet' },
  component: { label: 'Component', empty: 'No components in this app yet' },
} as const

function EntityPicker({
  type,
  block,
  options,
  onChange,
}: {
  type: keyof typeof entityEditorMeta
  block: TRunbookBlock | TActionBlock | TComponentBlock
  options: TEntityOption[]
  onChange: (block: TBlock) => void
}) {
  const { label, empty } = entityEditorMeta[type]
  const selectedId =
    block.id || (options.find((option) => option.name === block.name)?.id ?? '')
  if (!options.length) {
    return (
      <Input
        size="sm"
        labelProps={{ labelText: `${label} name` }}
        value={block.name}
        placeholder={empty}
        onChange={(event) =>
          onChange({ ...block, name: event.target.value, id: '' })
        }
      />
    )
  }
  return (
    <div className="max-w-96">
      <Select
        size="sm"
        searchable
        labelProps={{ labelText: label }}
        options={options.map((option) => ({
          value: option.id,
          label: option.name,
        }))}
        value={selectedId}
        placeholder={`Select a ${type}`}
        onChange={(value) => {
          const selected = options.find((option) => option.id === value)
          onChange({
            ...block,
            id: selected?.id ?? '',
            name: selected?.name ?? '',
          })
        }}
      />
    </div>
  )
}

const pathOptions = (variables: TStateVariable[]) =>
  variables.map((variable) => {
    const path = variable.template.replace(/^{{\.nuon\./, '').replace(/}}$/, '')
    return {
      value: path,
      label: variable.value ? `${path} · ${variable.value}` : path,
    }
  })

function VariableInsert({
  variables,
  onInsert,
}: {
  variables: TStateVariable[]
  onInsert: (template: string) => void
}) {
  return (
    <Select
      size="sm"
      searchable
      options={variables.map((variable) => ({
        value: variable.template,
        label: variable.value
          ? `${variable.template} · ${variable.value}`
          : variable.template,
      }))}
      value=""
      placeholder="Insert variable"
      onChange={(value) => {
        if (value) onInsert(value)
      }}
    />
  )
}

export function BlockEditor({
  block,
  variables,
  arraySources,
  runbooks = [],
  actions = [],
  components = [],
  onChange,
}: IBlockEditor) {
  switch (block.type) {
    case 'runbook':
      return (
        <EntityPicker
          type="runbook"
          block={block}
          options={runbooks}
          onChange={onChange}
        />
      )
    case 'action':
      return (
        <EntityPicker
          type="action"
          block={block}
          options={actions}
          onChange={onChange}
        />
      )
    case 'component':
      return (
        <EntityPicker
          type="component"
          block={block}
          options={components}
          onChange={onChange}
        />
      )
    case 'markdown':
      return (
        <div className="flex flex-col gap-2">
          <Textarea
            value={block.content}
            rows={Math.min(14, Math.max(3, block.content.split('\n').length))}
            className="font-mono text-sm"
            placeholder="Write markdown. Use {{.nuon.*}} variables anywhere."
            onChange={(event) =>
              onChange({ ...block, content: event.target.value })
            }
          />
          <div className="max-w-72">
            <VariableInsert
              variables={variables}
              onInsert={(template) =>
                onChange({ ...block, content: `${block.content}${template}` })
              }
            />
          </div>
        </div>
      )
    case 'raw':
      return (
        <Textarea
          value={block.content}
          rows={Math.min(14, Math.max(3, block.content.split('\n').length))}
          className="font-mono text-sm"
          placeholder="Verbatim Go template. Copied into the output untouched."
          onChange={(event) =>
            onChange({ ...block, content: event.target.value })
          }
        />
      )
    case 'banner':
      return (
        <div className="flex flex-col gap-3">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <Select
              size="sm"
              labelProps={{ labelText: 'Theme' }}
              options={themeOptions}
              value={block.theme}
              onChange={(value) =>
                onChange({
                  ...block,
                  theme: value as typeof block.theme,
                })
              }
            />
            <Select
              size="sm"
              labelProps={{ labelText: 'Visibility' }}
              options={conditionOptions}
              value={block.condition?.op ?? ''}
              onChange={(value) => {
                const op = value as TConditionOp | ''
                onChange({
                  ...block,
                  condition: op
                    ? {
                        path: block.condition?.path ?? '',
                        op,
                        value: block.condition?.value,
                      }
                    : undefined,
                })
              }}
            />
          </div>
          {block.condition ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <Select
                size="sm"
                searchable
                labelProps={{ labelText: 'State value' }}
                options={pathOptions(variables)}
                value={block.condition.path}
                placeholder="Select a .nuon value"
                onChange={(value) =>
                  onChange({
                    ...block,
                    condition: { ...block.condition!, path: value },
                  })
                }
              />
              {['eq', 'ne'].includes(block.condition.op) ? (
                <Input
                  size="sm"
                  labelProps={{ labelText: 'Compare to' }}
                  value={block.condition.value ?? ''}
                  onChange={(event) =>
                    onChange({
                      ...block,
                      condition: {
                        ...block.condition!,
                        value: event.target.value,
                      },
                    })
                  }
                />
              ) : null}
            </div>
          ) : null}
          <Textarea
            value={block.content}
            rows={3}
            placeholder="Banner message. Supports markdown and {{.nuon.*}} variables."
            onChange={(event) =>
              onChange({ ...block, content: event.target.value })
            }
          />
        </div>
      )
    case 'status-row':
      return (
        <div className="flex flex-col gap-2">
          {block.items.map((item, index) => (
            <div
              key={item.key}
              className="grid grid-cols-[1fr_1fr_1.4fr_auto] gap-2 items-end"
            >
              <Input
                size="sm"
                labelProps={index === 0 ? { labelText: 'Label' } : undefined}
                value={item.label}
                placeholder="Label"
                onChange={(event) =>
                  onChange({
                    ...block,
                    items: block.items.map((current) =>
                      current.key === item.key
                        ? { ...current, label: event.target.value }
                        : current
                    ),
                  })
                }
              />
              <Select
                size="sm"
                labelProps={
                  index === 0 ? { labelText: 'Render as' } : undefined
                }
                options={kindOptions}
                value={item.kind}
                onChange={(value) =>
                  onChange({
                    ...block,
                    items: block.items.map((current) =>
                      current.key === item.key
                        ? { ...current, kind: value as TValueKind }
                        : current
                    ),
                  })
                }
              />
              <Select
                size="sm"
                searchable
                labelProps={
                  index === 0 ? { labelText: 'State value' } : undefined
                }
                options={pathOptions(variables)}
                value={item.path}
                placeholder="Select a .nuon value"
                onChange={(value) =>
                  onChange({
                    ...block,
                    items: block.items.map((current) =>
                      current.key === item.key
                        ? { ...current, path: value }
                        : current
                    ),
                  })
                }
              />
              <Button
                variant="ghost"
                size="xs"
                aria-label="Remove item"
                disabled={block.items.length === 1}
                onClick={() =>
                  onChange({
                    ...block,
                    items: block.items.filter(
                      (current) => current.key !== item.key
                    ),
                  })
                }
              >
                <Icon variant="TrashIcon" size="14" />
              </Button>
            </div>
          ))}
          <div>
            <Button
              variant="secondary"
              size="xs"
              onClick={() =>
                onChange({
                  ...block,
                  items: [
                    ...block.items,
                    {
                      key: crypto.randomUUID(),
                      label: '',
                      kind: 'text',
                      path: '',
                    },
                  ],
                })
              }
            >
              <Icon variant="PlusIcon" size="14" />
              Add item
            </Button>
          </div>
        </div>
      )
    case 'table':
      return (
        <div className="flex flex-col gap-3">
          <div className="grid grid-cols-1 md:grid-cols-[2fr_1fr_1fr] gap-3">
            <Select
              size="sm"
              searchable
              labelProps={{ labelText: 'Data source (list)' }}
              options={arraySources.map((source) => ({
                value: source.path,
                label: `${source.path} · ${source.length} rows`,
              }))}
              value={block.sourcePath}
              placeholder={
                arraySources.length
                  ? 'Select a list from install state'
                  : 'Pick an install to browse lists'
              }
              onChange={(value) => onChange({ ...block, sourcePath: value })}
            />
            <Input
              size="sm"
              type="number"
              labelProps={{ labelText: 'Row limit' }}
              value={block.limit ?? ''}
              placeholder="All"
              onChange={(event) =>
                onChange({
                  ...block,
                  limit: event.target.value
                    ? Number(event.target.value)
                    : undefined,
                })
              }
            />
            <Input
              size="sm"
              labelProps={{ labelText: 'Empty state text' }}
              value={block.emptyText}
              onChange={(event) =>
                onChange({ ...block, emptyText: event.target.value })
              }
            />
          </div>
          {!arraySources.some((source) => source.path === block.sourcePath) &&
          block.sourcePath ? (
            <Input
              size="sm"
              labelProps={{ labelText: 'Custom source path' }}
              value={block.sourcePath}
              onChange={(event) =>
                onChange({ ...block, sourcePath: event.target.value })
              }
            />
          ) : null}
          {block.columns.map((column, index) => {
            const rowKeys =
              arraySources.find((source) => source.path === block.sourcePath)
                ?.keys ?? []
            return (
              <div
                key={column.key}
                className="grid grid-cols-[1fr_1fr_1.4fr_auto] gap-2 items-end"
              >
                <Input
                  size="sm"
                  labelProps={index === 0 ? { labelText: 'Column' } : undefined}
                  value={column.header}
                  placeholder="Header"
                  onChange={(event) =>
                    onChange({
                      ...block,
                      columns: block.columns.map((current) =>
                        current.key === column.key
                          ? { ...current, header: event.target.value }
                          : current
                      ),
                    })
                  }
                />
                <Select
                  size="sm"
                  labelProps={
                    index === 0 ? { labelText: 'Render as' } : undefined
                  }
                  options={kindOptions}
                  value={column.kind}
                  onChange={(value) =>
                    onChange({
                      ...block,
                      columns: block.columns.map((current) =>
                        current.key === column.key
                          ? {
                              ...current,
                              kind: value as TValueKind,
                            }
                          : current
                      ),
                    })
                  }
                />
                {rowKeys.length ? (
                  <Select
                    size="sm"
                    labelProps={
                      index === 0 ? { labelText: 'Row field' } : undefined
                    }
                    options={rowKeys.map((key) => ({ value: key, label: key }))}
                    value={column.path}
                    placeholder="Select field"
                    onChange={(value) =>
                      onChange({
                        ...block,
                        columns: block.columns.map((current) =>
                          current.key === column.key
                            ? { ...current, path: value }
                            : current
                        ),
                      })
                    }
                  />
                ) : (
                  <Input
                    size="sm"
                    labelProps={
                      index === 0 ? { labelText: 'Row field' } : undefined
                    }
                    value={column.path}
                    placeholder="field.path"
                    onChange={(event) =>
                      onChange({
                        ...block,
                        columns: block.columns.map((current) =>
                          current.key === column.key
                            ? { ...current, path: event.target.value }
                            : current
                        ),
                      })
                    }
                  />
                )}
                <Button
                  variant="ghost"
                  size="xs"
                  aria-label="Remove column"
                  disabled={block.columns.length === 1}
                  onClick={() =>
                    onChange({
                      ...block,
                      columns: block.columns.filter(
                        (current) => current.key !== column.key
                      ),
                    })
                  }
                >
                  <Icon variant="TrashIcon" size="14" />
                </Button>
              </div>
            )
          })}
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="xs"
              onClick={() =>
                onChange({
                  ...block,
                  columns: [
                    ...block.columns,
                    {
                      key: crypto.randomUUID(),
                      header: '',
                      kind: 'text',
                      path: '',
                    },
                  ],
                })
              }
            >
              <Icon variant="PlusIcon" size="14" />
              Add column
            </Button>
            <Text variant="subtext" theme="neutral">
              Rows repeat over the selected list at render time.
            </Text>
          </div>
        </div>
      )
  }
}
