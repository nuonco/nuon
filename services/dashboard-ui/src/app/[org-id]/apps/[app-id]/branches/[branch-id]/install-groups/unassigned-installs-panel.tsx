'use client'

import { useState } from 'react'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Icon } from '@/components/common/Icon'
import type { TInstall } from '@/types'
import { InstallCard } from './install-card'

interface IUnassignedInstallsPanel {
  installs: TInstall[]
  assignedInstallIds: string[]
}

export const UnassignedInstallsPanel = ({
  installs,
  assignedInstallIds,
}: IUnassignedInstallsPanel) => {
  const [searchQuery, setSearchQuery] = useState('')

  const unassignedInstalls = installs.filter(
    (install) => !assignedInstallIds.includes(install.id)
  )

  const filteredInstalls = unassignedInstalls.filter((install) =>
    install.name.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="w-80 flex flex-col border-r dark:border-gray-700 pr-6">
      <div className="mb-4">
        <Text variant="h4" weight="strong" className="mb-2">
          Available Installs
        </Text>
        <Text variant="subtext" theme="neutral">
          Drag installs to groups
        </Text>
      </div>

      {/* Search */}
      <div className="mb-4">
        <Input
          type="text"
          placeholder="Search installs..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          leftIcon={<Icon variant="Search" size={16} />}
        />
      </div>

      {/* Unassigned installs list */}
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

      {/* Stats */}
      <div className="mt-4 pt-4 border-t dark:border-gray-700">
        <Text variant="subtext" theme="neutral">
          {unassignedInstalls.length} of {installs.length} unassigned
        </Text>
      </div>
    </div>
  )
}