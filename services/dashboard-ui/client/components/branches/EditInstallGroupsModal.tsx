import { useEffect, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
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
  SortableContext,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Banner } from '@/components/common/Banner'
import { Skeleton } from '@/components/common/Skeleton'
import { Toast } from '@/components/surfaces/Toast'
import { createBranchConfig, getAppInstalls } from '@/lib'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TAppBranch, TAppBranchConfig, TInstall } from '@/types'
import { InstallGroupCard } from './install-groups/InstallGroupCard'
import { InstallCard } from './install-groups/InstallCard'
import { UnassignedInstallsSection } from './install-groups/UnassignedInstallsSection'
import { GroupConfigSection } from './install-groups/GroupConfigSection'

interface IInstallGroup {
  id: string
  name: string
  install_ids: string[]
  order: number
  max_parallel: number
  requires_approval: boolean
  rollback_on_failure: boolean
}

interface IEditInstallGroupsModal extends IModal {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  onSuccess?: () => void
}

export const EditInstallGroupsModal = ({
  branch,
  currentConfig,
  onSuccess,
  ...props
}: IEditInstallGroupsModal) => {
  const { app } = useApp()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const [groups, setGroups] = useState<IInstallGroup[]>(
    currentConfig?.install_groups?.map((group, idx) => ({
      id: group.id || `group-${idx}`,
      name: group.name || '',
      install_ids: group.install_ids || [],
      order: group.order || idx,
      max_parallel: group.max_parallel || 1,
      requires_approval: group.requires_approval || false,
      rollback_on_failure: group.rollback_on_failure || false,
    })) || []
  )
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null)
  const [availableInstalls, setAvailableInstalls] = useState<TInstall[]>([])
  const [loadingInstalls, setLoadingInstalls] = useState(false)
  const [activeId, setActiveId] = useState<string | null>(null)

  const formatError = (err: any): string => {
    if (!err) return ''
    if (typeof err === 'string') return err
    return (
      err.user_error ||
      err.error ||
      err.description ||
      err.message ||
      'An error occurred'
    )
  }

  const { mutate: saveMutation, isPending: isSaving } = useMutation({
    mutationFn: async () => {
      if (groups.length === 0) {
        throw new Error('At least one install group is required')
      }

      if (groups.some((g) => !g.name)) {
        throw new Error('All install groups must have a name')
      }

      const installGroupsForApi = groups.map((group, index) => ({
        name: group.name,
        install_ids: group.install_ids || [],
        order: index,
        max_parallel: group.max_parallel || 1,
        requires_approval: group.requires_approval || false,
        rollback_on_failure: group.rollback_on_failure || false,
      }))

      const request: any = { install_groups: installGroupsForApi }

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

      return createBranchConfig({
        appId: app.id,
        branchId: branch.id || '',
        orgId: org.id,
        request,
      })
    },
    onSuccess: () => {
      addToast(
        <Toast heading="Install groups saved successfully" theme="success">
          <Text>Your install group configuration has been updated.</Text>
        </Toast>
      )
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (error: Error) => {
      addToast(
        <Toast heading="Failed to save install groups" theme="error">
          <Text>{error.message || 'An unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  useEffect(() => {
    const fetchInstalls = async () => {
      setLoadingInstalls(true)
      try {
        const { data } = await getAppInstalls({
          appId: app.id,
          orgId: org.id,
          limit: 100,
        })
        setAvailableInstalls(data || [])
      } catch {
        // installs unavailable, leave list empty
      }
      setLoadingInstalls(false)
    }

    fetchInstalls()
  }, [app.id, org.id])

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

    const sourceGroup = groups.find((g) =>
      g.install_ids.includes(activeInstallId)
    )

    if (overGroupId.startsWith('group-')) {
      const targetGroup = groups.find((g) => g.id === overGroupId)
      if (!targetGroup) return

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
        setGroups(
          groups.map((g) =>
            g.id === targetGroup.id && !g.install_ids.includes(activeInstallId)
              ? { ...g, install_ids: [...g.install_ids, activeInstallId] }
              : g
          )
        )
      }
    }

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
  const assignedInstallIds = groups.flatMap((g) => g.install_ids)

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-0.5">
          <Text variant="h3" weight="strong">
            Edit install groups
          </Text>
          <Text variant="subtext" theme="neutral">
            Drag and drop installs to organize deployment groups
          </Text>
        </div>
      }
      size="full"
      primaryActionTrigger={{
        children: isSaving ? 'Saving...' : 'Save changes',
        onClick: () => saveMutation(),
        disabled: isSaving || loadingInstalls,
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        onClick: () => removeModal(props.modalId),
        disabled: isSaving,
      }}
      {...props}
    >
      {loadingInstalls ? (
        <Skeleton lines={3} />
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
            <UnassignedInstallsSection
              installs={availableInstalls}
              assignedInstallIds={assignedInstallIds}
            />

            <div className="flex-1 flex flex-col gap-4 overflow-y-auto">
              <div className="flex items-center justify-between mb-4">
                <Text variant="h3" weight="strong">
                  Install groups ({groups.length})
                </Text>
                <Button onClick={addNewGroup} variant="secondary" size="sm">
                  <Icon variant="Plus" size={16} />
                  Add group
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
                <div className="text-center py-12 border-2 border-dashed rounded-lg dark:border-dark-grey-600">
                  <Text variant="base" theme="neutral">
                    No install groups yet. Click &quot;Add group&quot; to create
                    one.
                  </Text>
                </div>
              )}
            </div>

            <GroupConfigSection
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

export const EditInstallGroupsButton = ({
  branch,
  currentConfig,
  onSuccess,
  ...props
}: { branch: TAppBranch; currentConfig?: TAppBranchConfig; onSuccess?: () => void } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <EditInstallGroupsModal branch={branch} currentConfig={currentConfig} onSuccess={onSuccess} />
  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="SlidersHorizontalIcon" size={16} />
      Manage installs
    </Button>
  )
}
