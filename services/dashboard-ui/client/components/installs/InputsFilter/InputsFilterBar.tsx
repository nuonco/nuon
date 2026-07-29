import React from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'
import {
  ATTRIBUTE_LABELS,
  ATTRIBUTE_OPTIONS,
  SOURCE_LABELS,
  SOURCE_OPTIONS,
  type TAttributeFilter,
  type TSourceFilter,
} from '@/hooks/use-inputs-filter'

interface IInputsFilterBar {
  search: string
  onSearchChange: (value: string) => void
  showFilters: boolean
  attributeFilters: TAttributeFilter[]
  sourceFilters: TSourceFilter[]
  setAttributeFilters: React.Dispatch<
    React.SetStateAction<TAttributeFilter[]>
  >
  setSourceFilters: React.Dispatch<React.SetStateAction<TSourceFilter[]>>
  toggleAttribute: (filter: TAttributeFilter) => void
  toggleSource: (filter: TSourceFilter) => void
  clearAllFilters: () => void
  filterCount: number
  hasActiveFilters: boolean
}

export const InputsFilterBar = ({
  search,
  onSearchChange,
  showFilters,
  attributeFilters,
  sourceFilters,
  setAttributeFilters,
  setSourceFilters,
  toggleAttribute,
  toggleSource,
  clearAllFilters,
  filterCount,
  hasActiveFilters,
}: IInputsFilterBar) => {
  return (
    <div className="flex items-center gap-2">
      <SearchInput
        placeholder="Search inputs..."
        value={search}
        onChange={onSearchChange}
      />
      {showFilters ? (
        <Dropdown
          alignment="right"
          closeOnBlur={false}
          id="inputs-filter"
          buttonText={
            <>
              <Icon variant="FunnelIcon" size="14" />
              {filterCount > 0 ? `Filter (${filterCount})` : 'Filter'}
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
                        prev.length === 1 && prev[0] === opt ? [] : [opt]
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
                        prev.length === 1 && prev[0] === opt ? [] : [opt]
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
  )
}
