import { useState } from 'react'
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
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Banner } from '@/components/common/Banner'
import { Skeleton } from '@/components/common/Skeleton'
import type { TInstall } from '@/types'
import { InstallGroupCard } from '../install-groups/InstallGroupCard'
import { InstallCard } from '../install-groups/InstallCard'
import { UnassignedInstallsSection } from '../install-groups/UnassignedInstallsSection'
import { GroupConfigSection } from '../install-groups/GroupConfigSection'

export interface IInstallGroup {
  id: string
  name: string
  install_ids: string[]
  order: number
  max_parallel: number
  requires_approval: boolean
  rollback_on_failure: boolean
}

interface IEditInstallGroupsModal extends IModal {
  initialGroups: IInstallGroup[]
  availableInstalls: TInstall[]
  loadingInstalls: boolean
  isSaving: boolean
  onSave: (groups: IInstallGroup[]) => void
  onCancel: () => void
}

export const EditInstallGroupsModal = ({
  initialGroups,
  availableInstalls,
  loadingInstalls,
  isSaving,
  onSave,
  onCancel,
  ...props
}: IEditInstallGroupsModal) => {
  const [groups, setGroups] = useState<IInstallGroup[]>(initialGroups)
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null)
  const [activeId, setActiveId] = useState<string | null>(null)

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
      size="lg"
      primaryActionTrigger={{
        children: isSaving ? 'Saving...' : 'Save changes',
        onClick: () => onSave(groups),
        disabled: isSaving || loadingInstalls,
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        onClick: onCancel,
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
                  <Icon variant="PlusIcon" size={16} />
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
                <EmptyState
                  variant="diagram"
                  emptyTitle="No install groups yet"
                  emptyMessage={`Click "Add group" above to create your first deployment group.`}
                />
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
