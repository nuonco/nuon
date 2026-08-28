import React from 'react'
import { Expand } from '@/components/common/Expand'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { InputValue } from '@/components/installs/management/InputValue'
import {
  InputsFilterBar,
  InputsNoResults,
} from '@/components/installs/InputsFilter'
import {
  useInputsFilter,
  type TInputsFilterGroup,
} from '@/hooks/use-inputs-filter'
import { getInputDisplayName } from '@/utils/install-utils'

interface IViewCurrentInputsModal extends IModal {
  isLoading: boolean
  redactedValues: Record<string, any>
  inputGroups: TInputsFilterGroup[]
  footerActions?: React.ReactNode
}

export const ViewCurrentInputsModal = ({
  isLoading,
  redactedValues,
  inputGroups,
  footerActions,
  ...props
}: IViewCurrentInputsModal) => {
  const redacted = redactedValues
  const hasConfig = inputGroups.length > 0
  const hasInputs = Object.keys(redacted).length > 0

  const {
    search,
    setSearch,
    attributeFilters,
    sourceFilters,
    setAttributeFilters,
    setSourceFilters,
    toggleAttribute,
    toggleSource,
    clearAllFilters,
    clearAll,
    filterCount,
    hasActiveSearch,
    hasActiveFilters,
    filteredGroups,
    filteredFlatInputs,
  } = useInputsFilter({ inputGroups, redacted })

  const noResultsEmpty = (
    <InputsNoResults
      search={search}
      hasActiveSearch={hasActiveSearch}
      hasActiveFilters={hasActiveFilters}
      onClearSearch={() => setSearch('')}
      onClearFilters={clearAllFilters}
      onClearAll={clearAll}
    />
  )

  return (
    <Modal
      className="!m-0 !mx-auto !mt-[10vh] !h-[80vh]"
      childrenClassName="flex-1"
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="ListChecksIcon" size="24" />
          Current inputs
        </Text>
      }
      size="xl"
      footerActions={footerActions}
      actions={
        !isLoading && (hasConfig || hasInputs) ? (
          <InputsFilterBar
            search={search}
            onSearchChange={setSearch}
            showFilters={hasConfig}
            attributeFilters={attributeFilters}
            sourceFilters={sourceFilters}
            setAttributeFilters={setAttributeFilters}
            setSourceFilters={setSourceFilters}
            toggleAttribute={toggleAttribute}
            toggleSource={toggleSource}
            clearAllFilters={clearAllFilters}
            filterCount={filterCount}
            hasActiveFilters={hasActiveFilters}
          />
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
                <div className="p-4 border-t bg-black/[0.0075] dark:bg-white/[0.0075]">
                  <PropertyGrid
                    align="start"
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
                            {input.name
                              ? getInputDisplayName(input.name)
                              : null}
                          </Text>
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
                          variant={input?.required ? 'CheckIcon' : 'MinusIcon'}
                        />
                      ),
                      sensitive: (
                        <Icon
                          variant={input?.sensitive ? 'CheckIcon' : 'MinusIcon'}
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
