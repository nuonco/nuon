import { Select, type SelectOption } from '@/components/common/form/Select'
import { SearchInput } from '@/components/common/SearchInput'
import type { TServiceAccountManagement } from '@/types'
import { humanize } from '@/utils/string-utils'

export const SERVICE_ACCOUNT_MANAGEMENT_OPTIONS: {
  label: string
  value: TServiceAccountManagement
}[] = [
  { label: 'User-managed', value: 'user' },
  { label: 'System', value: 'system' },
  { label: 'All', value: 'all' },
]

export interface IServiceAccountFilters {
  management: TServiceAccountManagement
  purpose: string
  purposes: string[]
  search: string
  onManagementChange: (management: TServiceAccountManagement) => void
  onPurposeChange: (purpose: string) => void
  onSearchChange: (search: string) => void
}

const purposeOptions = (
  purposes: string[],
  selected: string
): SelectOption[] => {
  const values =
    selected && !purposes.includes(selected)
      ? [...purposes, selected]
      : purposes

  return [
    { label: 'All purposes', value: '' },
    ...values.map((value) => ({ label: humanize(value), value })),
  ]
}

export const ServiceAccountFilters = ({
  management,
  purpose,
  purposes,
  search,
  onManagementChange,
  onPurposeChange,
  onSearchChange,
}: IServiceAccountFilters) => (
  <div
    role="search"
    aria-label="Filter service accounts"
    className="flex w-full flex-wrap items-end gap-4"
  >
    <SearchInput
      aria-label="Search service accounts"
      labelClassName="w-full md:w-fit"
      className="w-full md:w-fit dark:placeholder:!text-cool-grey-400"
      placeholder={
        management === 'user'
          ? 'Search by name or identity...'
          : 'Search by name, identity, or resource...'
      }
      value={search}
      onChange={onSearchChange}
    />
    <div className="flex flex-wrap items-end gap-4 md:ml-auto">
      <Select
        id="service-account-management"
        size="sm"
        className="min-w-40"
        labelProps={{
          labelText: 'Accounts',
          labelTextProps: { variant: 'label', theme: 'neutral' },
        }}
        options={SERVICE_ACCOUNT_MANAGEMENT_OPTIONS}
        value={management}
        onChange={(value) =>
          onManagementChange(value as TServiceAccountManagement)
        }
      />
      {management !== 'user' ? (
        <Select
          id="service-account-purpose"
          size="sm"
          className="min-w-48"
          searchable={purposes.length > 8}
          labelProps={{
            labelText: 'Purpose',
            labelTextProps: { variant: 'label', theme: 'neutral' },
          }}
          options={purposeOptions(purposes, purpose)}
          value={purpose}
          onChange={onPurposeChange}
        />
      ) : null}
    </div>
  </div>
)
