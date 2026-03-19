import { useState } from 'react'
import { Text } from '@/components/common/Text'
import { SearchInput } from '@/components/common/SearchInput'
import type { TInstall } from '@/types'
import { InstallCard } from './InstallCard'

interface IUnassignedInstallsSection {
  installs: TInstall[]
  assignedInstallIds: string[]
}

export const UnassignedInstallsSection = ({
  installs,
  assignedInstallIds,
}: IUnassignedInstallsSection) => {
  const [searchQuery, setSearchQuery] = useState('')

  const unassignedInstalls = installs.filter(
    (install) => !assignedInstallIds.includes(install.id)
  )

  const filteredInstalls = unassignedInstalls.filter((install) =>
    install.name.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="w-80 flex flex-col border-r dark:border-dark-grey-700 pr-6">
      <div className="mb-4 flex flex-col">
        <Text variant="h2" weight="strong" className="mb-2">
          Available installs
        </Text>
        <Text variant="subtext" theme="neutral">
          Drag installs to groups
        </Text>
      </div>

      <div className="mb-4">
        <SearchInput
          placeholder="Search installs..."
          value={searchQuery}
          onChange={setSearchQuery}
          className="md:!min-w-full"
          labelClassName="w-full"
        />
      </div>

      <div className="flex-1 space-y-2 overflow-y-auto">
        {filteredInstalls.length > 0 ? (
          filteredInstalls.map((install) => (
            <InstallCard key={install.id} install={install} />
          ))
        ) : (
          <div className="text-center py-8">
            <Text variant="subtext" theme="neutral">
              {searchQuery
                ? 'No installs match your search'
                : 'All installs are assigned'}
            </Text>
          </div>
        )}
      </div>

      <div className="mt-4 pt-4 border-t dark:border-dark-grey-700">
        <Text variant="subtext" theme="neutral">
          {unassignedInstalls.length} of {installs.length} unassigned
        </Text>
      </div>
    </div>
  )
}
