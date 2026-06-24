import React, { useState, useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { InputValue } from '@/components/installs/management/InputValue'
import {
  disabledToggleableDeps,
  getEnabledOverrideComponent,
  getInputDisplayName,
} from '@/utils/install-utils'

type TAttributeFilter = 'required' | 'sensitive'
type TSourceFilter = 'vendor' | 'customer'

const ATTRIBUTE_OPTIONS: TAttributeFilter[] = ['required', 'sensitive']
const SOURCE_OPTIONS: TSourceFilter[] = ['vendor', 'customer']

const ATTRIBUTE_LABELS: Record<TAttributeFilter, string> = {
  required: 'Required',
  sensitive: 'Sensitive',
}

const SOURCE_LABELS: Record<TSourceFilter, string> = {
  vendor: 'Vendor',
  customer: 'Customer',
}

type TInputGroup = {
  id: string
  display_name?: string
  description?: string
  app_inputs?: Array<{
    name?: string
    display_name?: string
    description?: string
    default?: string
    required?: boolean
    sensitive?: boolean
    source?: string
    group_id?: string
  }>
}

type TEnablementComponent = {
  component_id?: string
  component_name?: string
  component_dependency_ids?: string[]
  refs?: { type?: string; name?: string }[]
}

interface IViewCurrentInputsModal extends IModal {
  isLoading: boolean
  redactedValues: Record<string, any>
  inputGroups: TInputGroup[]
  configComponents?: TEnablementComponent[]
  effectiveEnabledByName?: Record<string, boolean | undefined>
  footerActions?: React.ReactNode
}

export const ViewCurrentInputsModal = ({
  isLoading,
  redactedValues,
  inputGroups,
  configComponents = [],
  effectiveEnabledByName = {},
  footerActions,
  ...props
}: IViewCurrentInputsModal) => {
  const [search, setSearch] = useState('')
  const [attributeFilters, setAttributeFilters] = useState<TAttributeFilter[]>([])
  const [sourceFilters, setSourceFilters] = useState<TSourceFilter[]>([])

  const redacted = redactedValues
  const hasConfig = inputGroups.length > 0
  const hasInputs = Object.keys(redacted).length > 0
  const query = search.toLowerCase()
  const hasActiveSearch = query.length > 0
  const filterCount = attributeFilters.length + sourceFilters.length
  const hasActiveFilters = filterCount > 0

  const clearAllFilters = () => {
    setAttributeFilters([])
    setSourceFilters([])
  }

  const clearAll = () => {
    setSearch('')
    clearAllFilters()
  }

  const noResultsEmpty = (
    <EmptyState
      emptyTitle="No matching inputs"
      emptyMessage={
        hasActiveSearch && hasActiveFilters
          ? `No inputs match "${search}" with the selected filters.`
          : hasActiveSearch
            ? `No inputs match "${search}".`
            : `No inputs match the selected filters.`
      }
      variant="diagram"
      size="sm"
      action={
        <div className="flex items-center gap-2">
          {hasActiveSearch ? (
            <Button size="sm" variant="ghost" onClick={() => setSearch('')}>
              Clear search
            </Button>
          ) : null}
          {hasActiveFilters ? (
            <Button
              size="sm"
              variant="ghost"
              onClick={clearAllFilters}
            >
              Reset filters
            </Button>
          ) : null}
          {hasActiveSearch && hasActiveFilters ? (
            <Button size="sm" variant="ghost" onClick={clearAll}>
              Clear all
            </Button>
          ) : null}
        </div>
      }
    />
  )

  const matchesFilters = (input: { required?: boolean; sensitive?: boolean; source?: string }) => {
    if (!hasActiveFilters) return true

    const matchesAttributes =
      attributeFilters.length === 0 ||
      attributeFilters.every((f) => {
        if (f === 'required') return input.required
        if (f === 'sensitive') return input.sensitive
        return false
      })

    const matchesSource =
      sourceFilters.length === 0 ||
      sourceFilters.some((f) => input.source === f)

    return matchesAttributes && matchesSource
  }

  const toggleAttribute = (filter: TAttributeFilter) => {
    setAttributeFilters((prev) =>
      prev.includes(filter)
        ? prev.filter((f) => f !== filter)
        : [...prev, filter]
    )
  }

  const toggleSource = (filter: TSourceFilter) => {
    setSourceFilters((prev) =>
      prev.includes(filter)
        ? prev.filter((f) => f !== filter)
        : [...prev, filter]
    )
  }

  const filteredGroups = useMemo(() => {
    return inputGroups
      .map((group) => ({
        ...group,
        app_inputs: (group.app_inputs ?? []).filter((input) => {
          if (!matchesFilters(input)) return false
          if (!query) return true
          const val = input.name ? String(redacted[input.name] ?? '') : ''
          return (
            input.name?.toLowerCase().includes(query) ||
            input.display_name?.toLowerCase().includes(query) ||
            input.description?.toLowerCase().includes(query) ||
            val.toLowerCase().includes(query)
          )
        }),
      }))
      .filter((group) => (group.app_inputs?.length ?? 0) > 0)
  }, [query, inputGroups, redacted, attributeFilters, sourceFilters])

  const filteredFlatInputs = useMemo(() => {
    if (!query) return Object.entries(redacted)
    return Object.entries(redacted).filter(
      ([key, value]) =>
        key.toLowerCase().includes(query) ||
        String(value).toLowerCase().includes(query)
    )
  }, [query, redacted])

  return (
    <Modal
      className="!m-0 !mx-auto !mt-[10vh] !h-[80vh]"
      childrenClassName="flex-1"
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
        >
          <Icon variant="ListChecksIcon" size="24" />
          Current inputs
        </Text>
      }
      size="xl"
      footerActions={footerActions}
      actions={
        !isLoading && (hasConfig || hasInputs) ? (
          <div className="flex items-center gap-2">
            <SearchInput
              placeholder="Search inputs..."
              value={search}
              onChange={setSearch}
            />
            {hasConfig ? (
              <Dropdown
                alignment="right"
                closeOnBlur={false}
                id="inputs-filter"
                buttonText={
                  <>
                    <Icon variant="FunnelIcon" size="14" />
                    {filterCount > 0
                      ? `Filter (${filterCount})`
                      : 'Filter'}
                  </>
                }
              >
                <Menu className="min-w-48">
                  <Text variant="label" theme="neutral">
                    Attributes
                  </Text>
                  {ATTRIBUTE_OPTIONS.map((opt) => (
                    <div className="flex items-center space-x-2" key={opt}>
                      <CheckboxInputWithButton
                        buttonProps={{
                          className:
                            '!p-1 flex items-center justify-between group/filter w-full',
                          children: (
                            <span className="font-semibold text-xs">
                              {ATTRIBUTE_LABELS[opt]}
                            </span>
                          ),
                          type: 'button',
                          variant: 'ghost',
                          onClick: () =>
                            setAttributeFilters((prev) =>
                              prev.length === 1 && prev[0] === opt
                                ? []
                                : [opt]
                            ),
                        }}
                        className="w-full"
                        name={opt}
                        onChange={() => toggleAttribute(opt)}
                        checked={attributeFilters.includes(opt)}
                        value={opt}
                      />
                    </div>
                  ))}
                  <hr />
                  <Text variant="label" theme="neutral">
                    Source
                  </Text>
                  {SOURCE_OPTIONS.map((opt) => (
                    <div className="flex items-center space-x-2" key={opt}>
                      <CheckboxInputWithButton
                        buttonProps={{
                          className:
                            '!p-1 flex items-center justify-between group/filter w-full',
                          children: (
                            <span className="font-semibold text-xs">
                              {SOURCE_LABELS[opt]}
                            </span>
                          ),
                          type: 'button',
                          variant: 'ghost',
                          onClick: () =>
                            setSourceFilters((prev) =>
                              prev.length === 1 && prev[0] === opt
                                ? []
                                : [opt]
                            ),
                        }}
                        className="w-full"
                        name={opt}
                        onChange={() => toggleSource(opt)}
                        checked={sourceFilters.includes(opt)}
                        value={opt}
                      />
                    </div>
                  ))}
                  {hasActiveFilters ? (
                    <>
                      <hr />
                      <Button
                        className="w-full !p-1 shrink-0"
                        type="button"
                        onClick={clearAllFilters}
                        size="sm"
                        variant="ghost"
                      >
                        Reset
                      </Button>
                    </>
                  ) : null}
                </Menu>
              </Dropdown>
            ) : null}
          </div>
        ) : null
      }
      {...props}
    >
      {isLoading ? (
        <div className="flex flex-col gap-4">
          <Skeleton width="100%" height="32px" />
          <Skeleton width="100%" height="32px" />
          <Skeleton width="100%" height="32px" />
        </div>
      ) : hasConfig ? (
        <div className="flex flex-col gap-4">
          {filteredGroups.length === 0 ? noResultsEmpty : null}
          {filteredGroups.map((group) => {
            const groupInputs = group.app_inputs ?? []
            if (groupInputs.length === 0) return null

            return (
              <Expand
                isOpen
                id={group.id}
                key={group.id}
                heading={
                  <div className="flex flex-col items-start">
                    <Text weight="strong">{group.display_name}</Text>
                    <Text variant="subtext" theme="neutral">
                      {group.description}
                    </Text>
                  </div>
                }
                className="border rounded-md"
                headerClassName="!px-4"
              >
                <div className="p-4 border-t bg-code">
                  <PropertyGrid
                    columns={[
                      { key: 'name', header: 'Name' },
                      { key: 'value', header: 'Current value' },
                      { key: 'default', header: 'Default' },
                      { key: 'description', header: 'Description' },
                      { key: 'required', header: 'Required' },
                      { key: 'sensitive', header: 'Sensitive' },
                      { key: 'source', header: 'Source' },
                    ]}
                    gridTemplate="minmax(130px, 2fr) minmax(140px, 2fr) minmax(100px, 2fr) minmax(180px, 2fr) minmax(80px, max-content) minmax(80px, max-content) minmax(80px, max-content)"
                    values={groupInputs.map((input) => ({
                      name: (
                        <span className="flex flex-col">
                          <Text variant="subtext" weight="strong">
                            {input.display_name}
                          </Text>
                          <Text variant="label" family="mono" theme="neutral">
                            {input.name ? getInputDisplayName(input.name) : null}
                          </Text>
                          {(() => {
                            const comp = input.name
                              ? getEnabledOverrideComponent(input.name)
                              : null
                            if (!comp) return null
                            const own =
                              input.name &&
                              String(redacted[input.name]) === 'true'
                            if (!own) return null
                            if (effectiveEnabledByName[comp] !== false)
                              return null
                            const blockers = disabledToggleableDeps(
                              comp,
                              configComponents,
                              effectiveEnabledByName
                            )
                            return (
                              <Text
                                variant="label"
                                theme="warn"
                                className="mt-0.5"
                              >
                                {blockers.length > 0
                                  ? `Effectively disabled — requires ${blockers.join(', ')}`
                                  : 'Effectively disabled — a required component is turned off'}
                              </Text>
                            )
                          })()}
                        </span>
                      ),
                      value: (
                        <InputValue
                          name={input.name}
                          value={input.name ? redacted[input.name] : undefined}
                        />
                      ),
                      default: (
                        <Text variant="label" family="mono" theme="neutral">
                          {input?.default}
                        </Text>
                      ),
                      description: (
                        <Text variant="subtext" theme="neutral">
                          {input?.description}
                        </Text>
                      ),
                      required: (
                        <Icon
                          variant={
                            input?.required ? 'CheckIcon' : 'MinusIcon'
                          }
                        />
                      ),
                      sensitive: (
                        <Icon
                          variant={
                            input?.sensitive ? 'CheckIcon' : 'MinusIcon'
                          }
                        />
                      ),
                      source: (
                        <Text
                          variant="label"
                          family="mono"
                          theme={input?.source === 'vendor' ? 'info' : 'brand'}
                        >
                          {input?.source}
                        </Text>
                      ),
                    }))}
                  />
                </div>
              </Expand>
            )
          })}
        </div>
      ) : hasInputs ? (
        filteredFlatInputs.length === 0 ? (
          noResultsEmpty
        ) : (
          <PropertyGrid
            values={filteredFlatInputs.map(([key, value]) => ({
              key: getInputDisplayName(key),
              value: <InputValue name={key} value={String(value)} />,
            }))}
          />
        )
      ) : (
        <EmptyState
          emptyTitle="No inputs configured"
          emptyMessage="This install doesn't have any inputs yet. Use the manage menu to edit inputs."
          variant="diagram"
          size="sm"
        />
      )}
    </Modal>
  )
}
