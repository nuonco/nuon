'use client'

import { useEffect, useState } from 'react'
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
import { Modal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Banner } from '@/components/common/Banner'
import { createBranchConfig, getAppInstalls } from '@/lib'
import type { TAppBranch, TAppBranchConfig, TInstall } from '@/types'
import { InstallGroupCard } from './install-groups/install-group-card'
import { InstallCard } from './install-groups/install-card'
import { UnassignedInstallsPanel } from './install-groups/unassigned-installs-panel'
import { GroupConfigPanel } from './install-groups/group-config-panel'

interface IInstallGroup {
  id: string
  name: string
  install_ids: string[]
  order: number
  max_parallel: number
  requires_approval: boolean
  rollback_on_failure: boolean
}

interface IEditInstallGroupsPanel {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  orgId: string
  appId: string
  isVisible: boolean
  onClose: () => void
}

export const EditInstallGroupsPanel = ({
  branch,
  currentConfig,
  orgId,
  appId,
  isVisible,
  onClose,
}: IEditInstallGroupsPanel) => {
  const router = useRouter()
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Install Groups State
  const [groups, setGroups] = useState<IInstallGroup[]>([])
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null)
  const [availableInstalls, setAvailableInstalls] = useState<TInstall[]>([])
  const [loadingInstalls, setLoadingInstalls] = useState(false)
  const [activeId, setActiveId] = useState<string | null>(null)

  // Initialize from current config
  useEffect(() => {
    if (isVisible) {
      if (currentConfig?.install_groups) {
        setGroups(
          currentConfig.install_groups.map((group, idx) => ({
            id: group.id || `group-${idx}`,
            name: group.name || '',
            install_ids: group.install_ids || [],
            order: group.order || idx,
            max_parallel: group.max_parallel || 1,
            requires_approval: group.requires_approval || false,
            rollback_on_failure: group.rollback_on_failure || false,
          }))
        )
      } else {
        setGroups([])
      }
    }
  }, [isVisible, currentConfig])

  // Fetch available installs
  useEffect(() => {
    if (!isVisible) return

    const fetchInstalls = async () => {
      setLoadingInstalls(true)
      const { data, error: installsError } = await getAppInstalls({
        appId,
        orgId,
        limit: 100,
      })

      if (installsError) {
        console.error('Failed to fetch installs:', installsError)
      } else {
        setAvailableInstalls(data || [])
      }
      setLoadingInstalls(false)
    }

    fetchInstalls()
  }, [isVisible, appId, orgId])

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

    // Find if install is currently in a group
    const sourceGroup = groups.find((g) =>
      g.install_ids.includes(activeInstallId)
    )

    // Dragging to a group
    if (overGroupId.startsWith('group-')) {
      const targetGroup = groups.find((g) => g.id === overGroupId)
      if (!targetGroup) return

      // Remove from source group if exists
      if (sourceGroup) {
        setGroups(
          groups.map((g) => {
            if (g.id === sourceGroup.id) {
              return {
                ...g,
                install_ids: g.install_ids.filter(
                  (id) => id !== activeInstallId
                ),
              }
            }
            if (
              g.id === targetGroup.id &&
              !g.install_ids.includes(activeInstallId)
            ) {
              return {
                ...g,
                install_ids: [...g.install_ids, activeInstallId],
              }
            }
            return g
          })
        )
      } else {
        // Add to target group from unassigned
        setGroups(
          groups.map((g) =>
            g.id === targetGroup.id && !g.install_ids.includes(activeInstallId)
              ? { ...g, install_ids: [...g.install_ids, activeInstallId] }
              : g
          )
        )
      }
    }

    // Dragging to unassigned area
    if (overGroupId === 'unassigned' && sourceGroup) {
      setGroups(
        groups.map((g) =>
          g.id === sourceGroup.id
            ? {
                ...g,
                install_ids: g.install_ids.filter(
                  (id) => id !== activeInstallId
                ),
              }
            : g
        )
      )
    }
  }

  const addNewGroup = () => {
    const newGroup: IInstallGroup = {
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

  const selectedGroup = groups.find((g) => g.id === selectedGroupId)

  const handleSave = async () => {
    setError(null)

    if (groups.length === 0) {
      setError('At least one install group is required')
      return
    }

    if (groups.some((g) => !g.name)) {
      setError('All install groups must have a name')
      return
    }

    setIsSaving(true)

    try {
      const installGroupsForApi = groups.map((group, index) => ({
        name: group.name,
        install_ids: group.install_ids || [],
        order: index,
        max_parallel: group.max_parallel || 1,
        requires_approval: group.requires_approval || false,
        rollback_on_failure: group.rollback_on_failure || false,
      }))

      const request: any = { install_groups: installGroupsForApi }

      // Preserve VCS config if it exists - ALWAYS include VCS config from current config
      if (currentConfig?.connected_github_vcs_config) {
        request.connected_github_vcs_config = {
          vcs_connection_id:
            currentConfig.connected_github_vcs_config.vcs_connection_id || '',
          repo: currentConfig.connected_github_vcs_config.repo || '',
          branch: currentConfig.connected_github_vcs_config.branch || '',
          directory: currentConfig.connected_github_vcs_config.directory,
          path_filter: currentConfig.connected_github_vcs_config.path_filter,
        }
      } else if (currentConfig?.public_git_vcs_config) {
        request.public_git_vcs_config = {
          repo: currentConfig.public_git_vcs_config.repo || '',
          branch: currentConfig.public_git_vcs_config.branch || '',
          directory: currentConfig.public_git_vcs_config.directory,
          path_filter: currentConfig.public_git_vcs_config.path_filter,
        }
      }

      const response = await createBranchConfig({
        appId,
        branchId: branch.id || '',
        orgId,
        request,
      })

      if (response.error) {
        setError(
          typeof response.error === 'string'
            ? response.error
            : response.error.user_error ||
                response.error.error ||
                response.error.description ||
                'Failed to save install groups'
        )
        setIsSaving(false)
      } else {
        router.refresh()
        onClose()
      }
    } catch (err) {
      // eslint-disable-next-line no-console
      console.error('Failed to save install groups:', err)
      setError('Failed to save install groups. Please try again.')
      setIsSaving(false)
    }
  }

  const assignedInstallIds = groups.flatMap((g) => g.install_ids)

  return (
    <Modal
      isVisible={isVisible}
      onClose={onClose}
      heading={
        <div>
          <Text variant="h3" weight="strong">
            Edit Install Groups
          </Text>
          <Text variant="subtext" theme="neutral">
            Drag and drop installs to organize deployment groups
          </Text>
        </div>
      }
      size="full"
      primaryActionTrigger={{
        children: isSaving ? 'Saving...' : 'Save Changes',
        onClick: handleSave,
        disabled: isSaving || loadingInstalls,
      }}
    >
      {error && (
        <Banner theme="error" className="mb-4">
          {error}
        </Banner>
      )}

      {loadingInstalls ? (
        <div className="flex items-center justify-center py-12">
          <Text variant="base" theme="neutral">
            Loading installs...
          </Text>
        </div>
      ) : availableInstalls.length === 0 ? (
        <Banner theme="info">
          No installs found for this app. Create installs first to configure
          deployment groups.
        </Banner>
      ) : (
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
        >
          <div className="flex gap-6 h-[600px]">
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
                items={groups.map((g) => g.id)}
                strategy={verticalListSortingStrategy}
              >
                {groups.map((group, index) => (
                  <InstallGroupCard
                    key={group.id}
                    group={group}
                    installs={availableInstalls.filter((i) =>
                      group.install_ids.includes(i.id)
                    )}
                    isSelected={selectedGroupId === group.id}
                    onClick={() => setSelectedGroupId(group.id)}
                    onRemoveInstall={(installId) => {
                      setGroups(
                        groups.map((g) =>
                          g.id === group.id
                            ? {
                                ...g,
                                install_ids: g.install_ids.filter(
                                  (id) => id !== installId
                                ),
                              }
                            : g
                        )
                      )
                    }}
                    index={index}
                  />
                ))}
              </SortableContext>

              {groups.length === 0 && (
                <div className="text-center py-12 border-2 border-dashed rounded-lg">
                  <Text variant="base" theme="neutral">
                    No install groups yet. Click &quot;Add Group&quot; to create
                    one.
                  </Text>
                </div>
              )}
            </div>

            {/* Right Panel: Group Configuration */}
            <GroupConfigPanel
              group={selectedGroup}
              availableInstalls={availableInstalls}
              onUpdate={(updates) => {
                setGroups(
                  groups.map((g) =>
                    g.id === selectedGroupId ? { ...g, ...updates } : g
                  )
                )
              }}
              onDelete={() => {
                setGroups(groups.filter((g) => g.id !== selectedGroupId))
                setSelectedGroupId(null)
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
      )}
    </Modal>
  )
}
