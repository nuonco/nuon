import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Text } from '@/components/common/Text'
import {
  InputsFilterBar,
  InputsNoResults,
} from '@/components/installs/InputsFilter'
import {
  useInputsFilter,
  type TInputsFilterGroup,
} from '@/hooks/use-inputs-filter'
import { normalizeAppInputGroups } from '@/utils/app-utils'
import type { TAppConfig } from '@/types'

export interface IAppInputs {
  appConfig: TAppConfig
}

export const AppInputs = ({ appConfig }: IAppInputs) => {
  const inputGroups = normalizeAppInputGroups(
    appConfig.input.input_groups,
    appConfig.input.inputs
  )

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
  } = useInputsFilter({
    inputGroups: inputGroups as TInputsFilterGroup[],
    redacted: {},
  })

  return (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <InputsFilterBar
          search={search}
          onSearchChange={setSearch}
          showFilters
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
      </div>

      {filteredGroups.length === 0 ? (
        <InputsNoResults
          search={search}
          hasActiveSearch={hasActiveSearch}
          hasActiveFilters={hasActiveFilters}
          onClearSearch={() => setSearch('')}
          onClearFilters={clearAllFilters}
          onClearAll={clearAll}
        />
      ) : (
        filteredGroups.map((inputGroup) => (
          <Expand
            isOpen
            id={inputGroup?.id}
            key={inputGroup.id}
            heading={
              <div className="flex flex-col items-start">
                <Text weight="strong">{inputGroup?.display_name}</Text>
                <Text variant="subtext" theme="neutral">
                  {inputGroup?.description}
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
                  { key: 'description', header: 'Description' },
                  { key: 'default', header: 'Default' },
                  { key: 'required', header: 'Required' },
                  { key: 'sensitive', header: 'Sensitive' },
                  { key: 'source', header: 'Source' },
                ]}
                gridTemplate="minmax(150px, 2fr) minmax(200px, 3fr) minmax(120px, 2fr) minmax(80px, max-content) minmax(80px, max-content) minmax(80px, max-content)"
                values={inputGroup?.app_inputs?.map((input) => ({
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
                  description: (
                    <Text variant="subtext">{input?.description}</Text>
                  ),
                  default: (
                    <Text variant="label" family="mono" theme="neutral">
                      {input?.default}
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
        ))
      )}
    </div>
  )
}
