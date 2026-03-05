'use client'

import { useState } from 'react'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { Input } from '@/components/common/form/Input'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import type { IInstallGroup } from '../new/types'
import { getMockInstalls } from '../new/mock-data'
import type { IInstall } from '../new/install-card'
import { cn } from '@/utils/classnames'

interface IInstallGroupsConfigProps {
  orgId: string
  appId: string
}

function getDefaultGroupTemplates(): IInstallGroup[] {
  return [
    {
      id: 'group-1',
      name: 'Staging',
      installIds: ['ins3', 'ins4'],
      order: 1,
      requiresApproval: false,
      rollbackOnFailure: true,
      maxParallel: 5,
    },
    {
      id: 'group-2',
      name: 'Production',
      installIds: ['ins1', 'ins2'],
      order: 2,
      requiresApproval: true,
      rollbackOnFailure: true,
      maxParallel: 2,
    },
  ]
}

export function InstallGroupsConfig({ orgId, appId }: IInstallGroupsConfigProps) {
  const [installGroups, setInstallGroups] = useState<IInstallGroup[]>(
    getDefaultGroupTemplates()
  )
  const [availableInstalls] = useState<IInstall[]>(getMockInstalls())
  const [searchQuery, setSearchQuery] = useState('')
  const [draggedInstallId, setDraggedInstallId] = useState<string | null>(null)
  const [dragOverGroupId, setDragOverGroupId] = useState<string | null>(null)
  const [dragOverAvailable, setDragOverAvailable] = useState(false)

  // Get installs that aren't in any group
  const assignedInstallIds = new Set(
    installGroups.flatMap(group => group.installIds)
  )
  const unassignedInstalls = availableInstalls.filter(
    install => !assignedInstallIds.has(install.id)
  )

  // Filter unassigned installs by search
  const filteredInstalls = unassignedInstalls.filter(install =>
    install.name.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const addToGroup = (groupId: string, installId: string) => {
    setInstallGroups(groups =>
      groups.map(group =>
        group.id === groupId
          ? { ...group, installIds: [...group.installIds, installId] }
          : group
      )
    )
  }

  const removeFromGroup = (groupId: string, installId: string) => {
    setInstallGroups(groups =>
      groups.map(group =>
        group.id === groupId
          ? {
              ...group,
              installIds: group.installIds.filter(id => id !== installId),
            }
          : group
      )
    )
  }

  const moveGroupUp = (groupId: string) => {
    setInstallGroups(groups => {
      const index = groups.findIndex(g => g.id === groupId)
      if (index <= 0) return groups
      
      const newGroups = [...groups]
      const temp = newGroups[index - 1]
      newGroups[index - 1] = { ...newGroups[index], order: index }
      newGroups[index] = { ...temp, order: index + 1 }
      return newGroups
    })
  }

  const moveGroupDown = (groupId: string) => {
    setInstallGroups(groups => {
      const index = groups.findIndex(g => g.id === groupId)
      if (index < 0 || index >= groups.length - 1) return groups
      
      const newGroups = [...groups]
      const temp = newGroups[index + 1]
      newGroups[index + 1] = { ...newGroups[index], order: index + 2 }
      newGroups[index] = { ...temp, order: index + 1 }
      return newGroups
    })
  }

  const updateGroupSettings = (
    groupId: string,
    updates: Partial<IInstallGroup>
  ) => {
    setInstallGroups(groups =>
      groups.map(group =>
        group.id === groupId ? { ...group, ...updates } : group
      )
    )
  }

  const addNewGroup = () => {
    const newGroup: IInstallGroup = {
      id: `group-${Date.now()}`,
      name: `Group ${installGroups.length + 1}`,
      installIds: [],
      order: installGroups.length + 1,
      requiresApproval: false,
      rollbackOnFailure: false,
      maxParallel: 3,
    }
    setInstallGroups([...installGroups, newGroup])
  }

  const deleteGroup = (groupId: string) => {
    setInstallGroups(groups => groups.filter(g => g.id !== groupId))
  }

  const handleSave = () => {
    console.log('Saving install groups configuration:', installGroups)
    alert('Configuration saved to console! (Check browser console)')
  }

  // Drag and Drop Handlers
  const handleDragStart = (
    e: React.DragEvent,
    installId: string,
    sourceGroupId?: string
  ) => {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('installId', installId)
    e.dataTransfer.setData('sourceGroupId', sourceGroupId || 'available')
    setDraggedInstallId(installId)
  }

  const handleDragEnd = () => {
    setDraggedInstallId(null)
    setDragOverGroupId(null)
    setDragOverAvailable(false)
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }

  const handleDragEnterGroup = (groupId: string) => {
    setDragOverGroupId(groupId)
  }

  const handleDragLeaveGroup = () => {
    setDragOverGroupId(null)
  }

  const handleDragEnterAvailable = () => {
    setDragOverAvailable(true)
  }

  const handleDragLeaveAvailable = () => {
    setDragOverAvailable(false)
  }

  const handleDropOnGroup = (e: React.DragEvent, targetGroupId: string) => {
    e.preventDefault()
    const installId = e.dataTransfer.getData('installId')
    const sourceGroupId = e.dataTransfer.getData('sourceGroupId')

    if (!installId) return

    // If source is "available", just add to target group
    if (sourceGroupId === 'available') {
      addToGroup(targetGroupId, installId)
    }
    // If source is a group, remove from source and add to target
    else if (sourceGroupId !== targetGroupId) {
      removeFromGroup(sourceGroupId, installId)
      addToGroup(targetGroupId, installId)
    }

    setDragOverGroupId(null)
  }

  const handleDropOnAvailable = (e: React.DragEvent) => {
    e.preventDefault()
    const installId = e.dataTransfer.getData('installId')
    const sourceGroupId = e.dataTransfer.getData('sourceGroupId')

    if (!installId || sourceGroupId === 'available') return

    // Remove from source group
    removeFromGroup(sourceGroupId, installId)
    setDragOverAvailable(false)
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Two-panel layout */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Left Panel: Available Installs */}
        <Card>
          <div
            className="flex flex-col gap-4"
            onDragOver={handleDragOver}
            onDrop={handleDropOnAvailable}
            onDragEnter={handleDragEnterAvailable}
            onDragLeave={handleDragLeaveAvailable}
          >
            <div className="flex items-center justify-between">
              <Text variant="h4" weight="strong">
                Available Installs ({filteredInstalls.length})
              </Text>
            </div>

            <Input
              placeholder="Search installs..."
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
            />

            {/* Drop zone hint when dragging from a group */}
            {draggedInstallId && !filteredInstalls.some(i => i.id === draggedInstallId) && (
              <div
                className={cn(
                  'border-2 border-dashed rounded-lg p-4 text-center transition-colors',
                  dragOverAvailable
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30'
                    : 'border-blue-300 dark:border-blue-700 bg-blue-50/50 dark:bg-blue-900/10'
                )}
              >
                <Icon variant="ArrowDown" size={24} className="mx-auto mb-2 text-blue-500" />
                <Text variant="subtext" theme="neutral">
                  Drop here to unassign from group
                </Text>
              </div>
            )}

            <div className="space-y-2 max-h-[500px] overflow-y-auto">
              {filteredInstalls.map(install => (
                <div
                  key={install.id}
                  draggable
                  onDragStart={e => handleDragStart(e, install.id)}
                  onDragEnd={handleDragEnd}
                  className={cn(
                    'flex items-center justify-between p-3 rounded-md border transition-all select-none',
                    draggedInstallId === install.id
                      ? 'opacity-50 scale-95 cursor-grabbing'
                      : 'cursor-grab bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700 hover:border-blue-400 hover:shadow-md'
                  )}
                >
                  <div className="flex items-center gap-2 flex-1 min-w-0">
                    <Icon
                      variant="DotsSixVertical"
                      size={16}
                      className="text-gray-400 flex-shrink-0"
                    />
                    <div className="flex-1 min-w-0">
                      <Text variant="base" className="truncate">
                        {install.name}
                      </Text>
                      <Text variant="subtext" theme="neutral" className="text-xs">
                        {install.region} • {install.platform}
                      </Text>
                    </div>
                  </div>
                  <Badge variant="default" theme="neutral">
                    {install.status}
                  </Badge>
                </div>
              ))}

              {filteredInstalls.length === 0 && !draggedInstallId && (
                <div className="text-center py-8">
                  <Text variant="base" theme="neutral">
                    {searchQuery
                      ? 'No installs match your search'
                      : 'All installs are assigned to groups'}
                  </Text>
                </div>
              )}
            </div>
          </div>
        </Card>

        {/* Right Panel: Deployment Groups */}
        <Card>
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <Text variant="h4" weight="strong">
                Deployment Groups ({installGroups.length})
              </Text>
              <Button onClick={addNewGroup} variant="secondary" size="sm">
                + Add Group
              </Button>
            </div>

            <div className="space-y-4 max-h-[500px] overflow-y-auto">
              {installGroups.map((group, index) => {
                const groupInstalls = availableInstalls.filter(i =>
                  group.installIds.includes(i.id)
                )
                const isDropTarget = dragOverGroupId === group.id

                return (
                  <Card key={group.id}>
                    <div className="flex flex-col gap-3">
                      {/* Group Header */}
                      <div className="flex items-center gap-2">
                        <Badge variant="default" size="sm">
                          {index + 1}
                        </Badge>
                        <Input
                          value={group.name}
                          onChange={e =>
                            updateGroupSettings(group.id, { name: e.target.value })
                          }
                          className="flex-1"
                        />
                        <div className="flex gap-1">
                          <Button
                            onClick={() => moveGroupUp(group.id)}
                            disabled={index === 0}
                            variant="ghost"
                            size="sm"
                          >
                            <Icon variant="CaretUp" size={16} />
                          </Button>
                          <Button
                            onClick={() => moveGroupDown(group.id)}
                            disabled={index === installGroups.length - 1}
                            variant="ghost"
                            size="sm"
                          >
                            <Icon variant="CaretDown" size={16} />
                          </Button>
                          <Button
                            onClick={() => deleteGroup(group.id)}
                            variant="ghost"
                            size="sm"
                          >
                            <Icon variant="Trash" size={16} />
                          </Button>
                        </div>
                      </div>

                      {/* Drop zone for installs */}
                      <div
                        onDragOver={handleDragOver}
                        onDrop={e => handleDropOnGroup(e, group.id)}
                        onDragEnter={() => handleDragEnterGroup(group.id)}
                        onDragLeave={handleDragLeaveGroup}
                        className={cn(
                          'min-h-[100px] border-2 border-dashed rounded-lg p-2 transition-colors',
                          isDropTarget
                            ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30'
                            : 'border-gray-300 dark:border-gray-600 hover:border-blue-400 hover:bg-blue-50/30 dark:hover:bg-blue-900/10'
                        )}
                      >
                        {groupInstalls.length === 0 && (
                          <div className="h-full flex items-center justify-center text-center py-4">
                            <div>
                              <Icon
                                variant="Plus"
                                size={24}
                                className={cn(
                                  'mx-auto mb-2 transition-colors',
                                  isDropTarget ? 'text-blue-500' : 'text-gray-400'
                                )}
                              />
                              <Text variant="subtext" theme="neutral">
                                Drag installs here
                              </Text>
                            </div>
                          </div>
                        )}

                        {/* Installs in this group */}
                        <div className="space-y-1">
                          {groupInstalls.map(install => (
                            <div
                              key={install.id}
                              draggable
                              onDragStart={e => handleDragStart(e, install.id, group.id)}
                              onDragEnd={handleDragEnd}
                              className={cn(
                                'flex items-center justify-between p-2 rounded border transition-all select-none',
                                draggedInstallId === install.id
                                  ? 'opacity-50 scale-95 cursor-grabbing'
                                  : 'cursor-grab bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-700 hover:border-blue-400 hover:shadow-sm'
                              )}
                            >
                              <div className="flex items-center gap-2">
                                <Icon
                                  variant="DotsSixVertical"
                                  size={14}
                                  className="text-gray-400"
                                />
                                <Text variant="subtext">{install.name}</Text>
                              </div>
                              <Button
                                onClick={e => {
                                  e.stopPropagation()
                                  removeFromGroup(group.id, install.id)
                                }}
                                variant="ghost"
                                size="sm"
                              >
                                <Icon variant="X" size={14} />
                              </Button>
                            </div>
                          ))}
                        </div>
                      </div>

                      {/* Group Settings */}
                      <div className="border-t pt-3 space-y-2">
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={group.requiresApproval}
                            onChange={e =>
                              updateGroupSettings(group.id, {
                                requiresApproval: e.target.checked,
                              })
                            }
                            className="h-4 w-4 rounded border-gray-300"
                          />
                          <Text variant="subtext">Require manual approval</Text>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={group.rollbackOnFailure}
                            onChange={e =>
                              updateGroupSettings(group.id, {
                                rollbackOnFailure: e.target.checked,
                              })
                            }
                            className="h-4 w-4 rounded border-gray-300"
                          />
                          <Text variant="subtext">Rollback on failure</Text>
                        </label>
                        <div className="flex items-center gap-2">
                          <Text variant="subtext">Max parallel:</Text>
                          <Input
                            type="number"
                            min="1"
                            max="100"
                            value={group.maxParallel}
                            onChange={e =>
                              updateGroupSettings(group.id, {
                                maxParallel: parseInt(e.target.value) || 1,
                              })
                            }
                            className="w-20"
                          />
                        </div>
                      </div>
                    </div>
                  </Card>
                )
              })}
            </div>
          </div>
        </Card>
      </div>

      {/* Save Button */}
      <div className="flex justify-end gap-3">
        <Button onClick={() => console.log(installGroups)} variant="secondary">
          Log to Console
        </Button>
        <Button onClick={handleSave} variant="primary">
          Save Configuration
        </Button>
      </div>
    </div>
  )
}