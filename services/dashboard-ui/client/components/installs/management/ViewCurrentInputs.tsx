import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
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
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfig, getInstallCurrentInputs } from '@/lib'
import { normalizeAppInputGroups } from '@/utils/app-utils'

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

export const ViewCurrentInputsModal = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [search, setSearch] = useState('')
  const [attributeFilters, setAttributeFilters] = useState<TAttributeFilter[]>([])
  const [sourceFilters, setSourceFilters] = useState<TSourceFilter[]>([])

  const { data: inputs, isLoading: inputsLoading } = useQuery({
    queryKey: ['install-inputs', org?.id, install?.id],
    queryFn: () =>
      getInstallCurrentInputs({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  const { data: config, isLoading: configLoading } = useQuery({
    queryKey: ['app-config', org?.id, install?.app_id, install?.app_config_id],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!install?.app_config_id,
  })

  const isLoading = inputsLoading || configLoading
  const redacted = inputs?.redacted_values ?? {}
  const inputGroups = config
    ? normalizeAppInputGroups(
        config.input?.input_groups ?? [],
        config.input?.inputs ?? []
      )
    : []

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
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
          <Icon variant="ListChecks" size="24" />
          Current inputs
        </Text>
      }
      size="full"
      className="!mt-12 !mb-auto"
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
                className="!p-2"
                closeOnBlur={false}
                id="inputs-filter"
                variant="ghost"
                buttonClassName="!p-1"
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
                            {input.name}
                          </Text>
                        </span>
                      ),
                      value:
                        input.name && redacted[input.name] != null ? (
                          String(redacted[input.name]) === '' ? (
                            <Text variant="subtext" family="mono" theme="neutral">
                              ""
                            </Text>
                          ) : (
                            <Text variant="subtext" family="mono" weight="strong">
                              {String(redacted[input.name])}
                            </Text>
                          )
                        ) : (
                          <Text variant="subtext" family="mono" theme="neutral">
                            —
                          </Text>
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
              key,
              value,
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

export const ViewCurrentInputsButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
      variant="ghost"
      onClick={() => {
        const modal = <ViewCurrentInputsModal />
        addModal(modal)
      }}
      {...props}
    >
      Current inputs
      <Icon variant="ListChecks" />
    </Button>
  )
}
