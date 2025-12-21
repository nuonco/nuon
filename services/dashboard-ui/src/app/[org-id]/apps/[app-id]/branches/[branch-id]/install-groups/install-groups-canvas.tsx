'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import {
  DndContext,
  DragOverlay,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
  DragStartEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Banner } from '@/components/common/Banner'
import { getAppInstalls, getAppBranch, createBranchConfig } from '@/lib'
import type { TInstall, TAppBranchInstallGroup, TAppBranchConfig } from '@/types'
import { UnassignedInstallsPanel } from './unassigned-installs-panel'
import { InstallGroupCard } from './install-group-card'
import { GroupConfigPanel } from './group-config-panel'
import { InstallCard } from './install-card'

interface IInstallGroupsCanvas {
  appId: string
  branchId: string
  orgId: string
}

export const InstallGroupsCanvas = ({
  appId,
  branchId,
  orgId,
}: IInstallGroupsCanvas) => {
  const router = useRouter()
  const [groups, setGroups] = useState<TAppBranchInstallGroup[]>([])
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null)
  const [availableInstalls, setAvailableInstalls] = useState<TInstall[]>([])
  const [activeId, setActiveId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [currentConfig, setCurrentConfig] = useState<TAppBranchConfig | null>(null)

  // Fetch all installs for this app and current branch config
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        // Fetch all installs for this app
        const installsResponse = await getAppInstalls({ appId, orgId, limit: 100 })
        setAvailableInstalls(installsResponse.data || [])

        // Fetch current branch config to get existing install groups
        const branchResponse = await getAppBranch({ appId, branchId, orgId })
        const config = branchResponse.data?.configs?.[0]
        setCurrentConfig(config || null)
        if (config?.install_groups) {
          setGroups(config.install_groups)
        }
      } catch (error) {
        console.error('Failed to fetch installs or branch config:', error)
        setError('Failed to load install groups')
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [appId, branchId, orgId])

  // Drag sensors
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  )

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string)
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    setActiveId(null)

    if (!over) return

    const activeInstallId = active.id as string
    const overGroupId = over.id as string

    // Check if dragging onto a group
    const targetGroup = groups.find((g) => g.id === overGroupId)
    if (targetGroup) {
      // Add install to group if not already present
      if (!targetGroup.install_ids?.includes(activeInstallId)) {
        setGroups(
          groups.map((g) =>
            g.id === targetGroup.id
              ? {
                  ...g,
                  install_ids: [...(g.install_ids || []), activeInstallId],
                }
              : g
          )
        )
      }
    }
  }

  const addNewGroup = () => {
    const newGroup: TAppBranchInstallGroup = {
      id: `group-${Date.now()}`,
      name: `Group ${groups.length + 1}`,
      install_ids: [],
      order: groups.length,
      max_parallel: 1,
      requires_approval: false,
      rollback_on_failure: false,
    }
    setGroups([...groups, newGroup])
    setSelectedGroupId(newGroup.id)
  }

  const updateGroup = (groupId: string, updates: Partial<TAppBranchInstallGroup>) => {
    setGroups(groups.map((g) => (g.id === groupId ? { ...g, ...updates } : g)))
  }

  const deleteGroup = (groupId: string) => {
    setGroups(groups.filter((g) => g.id !== groupId))
    if (selectedGroupId === groupId) {
      setSelectedGroupId(null)
    }
  }

  const removeInstallFromGroup = (groupId: string, installId: string) => {
    setGroups(
      groups.map((g) =>
        g.id === groupId
          ? {
              ...g,
              install_ids: (g.install_ids || []).filter((id) => id !== installId),
            }
          : g
      )
    )
  }

  const handleSave = async () => {
    setSaving(true)
    setError(null)

    try {
      // Prepare install groups for API
      const installGroupsForApi = groups.map((group, index) => ({
        name: group.name || `Group ${index + 1}`,
        install_ids: group.install_ids || [],
        order: index,
        max_parallel: group.max_parallel || 1,
        requires_approval: group.requires_approval || false,
        rollback_on_failure: group.rollback_on_failure || false,
      }))

      // Create new config with updated install groups
      const request: any = {
        install_groups: installGroupsForApi,
      }

      // Preserve VCS config if it exists
      if (currentConfig?.connected_github_vcs_config) {
        request.connected_github_vcs_config = {
          vcs_connection_id: currentConfig.connected_github_vcs_config.vcs_connection_id || '',
          repo: currentConfig.connected_github_vcs_config.repo || '',
          branch: currentConfig.connected_github_vcs_config.branch || '',
          directory: currentConfig.connected_github_vcs_config.directory,
          path_filter: currentConfig.connected_github_vcs_config.path_filter,
        }
      } else if (currentConfig?.public_git_vcs_config) {
        request.public_git_vcs_config = {
          url: currentConfig.public_git_vcs_config.url || '',
          branch: currentConfig.public_git_vcs_config.branch || '',
          directory: currentConfig.public_git_vcs_config.directory,
        }
      }

      await createBranchConfig({
        appId,
        branchId,
        orgId,
        request,
      })

      // Navigate back to branch detail page
      router.push(`/${orgId}/apps/${appId}/branches/${branchId}`)
      router.refresh()
    } catch (err) {
      console.error('Failed to save install groups:', err)
      setError('Failed to save install groups. Please try again.')
      setSaving(false)
    }
  }

  const selectedGroup = groups.find((g) => g.id === selectedGroupId)
  const assignedInstallIds = groups.flatMap((g) => g.install_ids || [])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Text variant="base" theme="neutral">
          Loading install groups...
        </Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header with Save/Cancel */}
      <div className="flex items-center justify-between mb-6 pb-4 border-b">
        <div>
          <Text variant="h3" weight="strong">
            Manage Install Groups
          </Text>
          <Text variant="subtext" theme="neutral">
            Configure deployment groups for this branch
          </Text>
        </div>
        <div className="flex items-center gap-3">
          <Button
            onClick={() => router.push(`/${orgId}/apps/${appId}/branches/${branchId}`)}
            variant="ghost"
            size="sm"
            disabled={saving}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            variant="primary"
            size="sm"
            disabled={saving || groups.length === 0}
          >
            {saving ? (
              <>
                <Icon variant="Loader" size={16} className="animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Icon variant="Save" size={16} />
                Save Groups
              </>
            )}
          </Button>
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <Banner theme="danger" className="mb-4">
          {error}
        </Banner>
      )}

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
      >
        <div className="flex gap-6 flex-1 overflow-hidden">
          {/* Left Panel: Unassigned Installs */}
          <UnassignedInstallsPanel
            installs={availableInstalls}
            assignedInstallIds={assignedInstallIds}
          />

          {/* Center Canvas: Install Groups */}
          <div className="flex-1 flex flex-col gap-4 overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <Text variant="h4" weight="strong">
                Install Groups ({groups.length})
              </Text>
              <Button onClick={addNewGroup} variant="secondary" size="sm">
                <Icon variant="Plus" size={16} />
                Add Group
              </Button>
            </div>

          <SortableContext
            items={groups.map((g) => g.id || '')}
            strategy={verticalListSortingStrategy}
          >
            {groups.map((group, index) => (
              <InstallGroupCard
                key={group.id}
                group={group}
                installs={availableInstalls.filter((i) =>
                  group.install_ids?.includes(i.id)
                )}
                isSelected={selectedGroupId === group.id}
                onClick={() => setSelectedGroupId(group.id || null)}
                onRemoveInstall={(installId) =>
                  removeInstallFromGroup(group.id || '', installId)
                }
                index={index}
              />
            ))}
          </SortableContext>

          {groups.length === 0 && (
            <div className="text-center py-12 border-2 border-dashed rounded-lg">
              <Text variant="base" theme="neutral">
                No install groups yet. Click &quot;Add Group&quot; to create one.
              </Text>
            </div>
          )}
        </div>

          {/* Right Panel: Group Configuration */}
          <GroupConfigPanel
            group={selectedGroup}
            availableInstalls={availableInstalls}
            assignedInstallIds={assignedInstallIds}
            onUpdate={(updates) => {
              if (selectedGroupId) {
                updateGroup(selectedGroupId, updates)
              }
            }}
            onDelete={() => {
              if (selectedGroupId) {
                deleteGroup(selectedGroupId)
              }
            }}
          />
        </div>

        {/* Drag Overlay */}
        <DragOverlay>
          {activeId ? (
            <InstallCard
              install={availableInstalls.find((i) => i.id === activeId)}
              isDragging
            />
          ) : null}
        </DragOverlay>
      </DndContext>
    </div>
  )
}