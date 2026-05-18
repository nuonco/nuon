import { useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  closestCorners,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from '@dnd-kit/core'
import { sortableKeyboardCoordinates } from '@dnd-kit/sortable'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TInstall } from '@/types'
import { GroupEditor } from './GroupEditor'
import { newGroup } from './lib'
import { UnassignedSection } from './UnassignedSection'
import type { IInstallGroup } from './types'

const UNASSIGNED_ID = '__unassigned__'

interface IDeploymentPlanEditor extends Omit<IModal, 'onSubmit'> {
  initialGroups: IInstallGroup[]
  availableInstalls: TInstall[]
  loadingInstalls: boolean
  isSaving: boolean
  onSave: (groups: IInstallGroup[]) => void
  onCancel: () => void
}

export const DeploymentPlanEditor = ({
  initialGroups,
  availableInstalls,
  loadingInstalls,
  isSaving,
  onSave,
  onCancel,
  ...props
}: IDeploymentPlanEditor) => {
  const [groups, setGroups] = useState<IInstallGroup[]>(initialGroups)
  const [showValidation, setShowValidation] = useState(false)
  const [activeId, setActiveId] = useState<string | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 6 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const installsById = useMemo(() => {
    const map = new Map<string, TInstall>()
    availableInstalls.forEach((i) => map.set(i.id, i))
    return map
  }, [availableInstalls])

  const assignedInstallIds = useMemo(
    () => new Set(groups.flatMap((g) => g.install_ids)),
    [groups]
  )

  const unassignedInstalls = useMemo(
    () => availableInstalls.filter((i) => !assignedInstallIds.has(i.id)),
    [availableInstalls, assignedInstallIds]
  )

  const activeInstall = activeId ? installsById.get(activeId) : null

  const hasEmptyNames = groups.some((g) => !g.name.trim())
  const canSave =
    !isSaving && !loadingInstalls && groups.length > 0 && !hasEmptyNames
  const isDisabled = isSaving || loadingInstalls

  const updateGroup = (id: string, updates: Partial<IInstallGroup>) => {
    setGroups((curr) =>
      curr.map((g) => (g.id === id ? { ...g, ...updates } : g))
    )
  }

  const addGroup = () => {
    setGroups((curr) => [...curr, newGroup(curr.length)])
  }

  const deleteGroup = (id: string) => {
    setGroups((curr) =>
      curr.filter((g) => g.id !== id).map((g, idx) => ({ ...g, order: idx }))
    )
  }

  const moveGroup = (id: string, delta: -1 | 1) => {
    setGroups((curr) => {
      const idx = curr.findIndex((g) => g.id === id)
      if (idx === -1) return curr
      const targetIdx = idx + delta
      if (targetIdx < 0 || targetIdx >= curr.length) return curr
      const next = [...curr]
      ;[next[idx], next[targetIdx]] = [next[targetIdx], next[idx]]
      return next.map((g, i) => ({ ...g, order: i }))
    })
  }

  const addInstallsToGroup = (groupId: string, installIds: string[]) => {
    const idSet = new Set(installIds)
    setGroups((curr) =>
      curr.map((g) => {
        if (g.id === groupId) {
          const merged = [...g.install_ids]
          installIds.forEach((id) => {
            if (!merged.includes(id)) merged.push(id)
          })
          return { ...g, install_ids: merged }
        }
        return {
          ...g,
          install_ids: g.install_ids.filter((i) => !idSet.has(i)),
        }
      })
    )
  }

  const removeInstallFromGroup = (groupId: string, installId: string) => {
    setGroups((curr) =>
      curr.map((g) =>
        g.id === groupId
          ? { ...g, install_ids: g.install_ids.filter((i) => i !== installId) }
          : g
      )
    )
  }

  const moveInstall = (
    installId: string,
    fromContainer: string,
    toContainer: string
  ) => {
    if (fromContainer === toContainer) return

    setGroups((curr) => {
      const next = curr.map((g) => {
        if (g.id === fromContainer) {
          return {
            ...g,
            install_ids: g.install_ids.filter((id) => id !== installId),
          }
        }
        if (g.id === toContainer) {
          const filtered = g.install_ids.filter((id) => id !== installId)
          return { ...g, install_ids: [...filtered, installId] }
        }
        return g
      })
      return next
    })
  }

  const findContainer = (id: string): string => {
    if (id === UNASSIGNED_ID) return UNASSIGNED_ID
    if (groups.some((g) => g.id === id)) return id
    for (const g of groups) {
      if (g.install_ids.includes(id)) return g.id
    }
    return UNASSIGNED_ID
  }

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string)
  }

  const handleDragEnd = (event: DragEndEvent) => {
    setActiveId(null)
    const { active, over } = event
    if (!over) return

    const activeId = active.id as string
    const overId = over.id as string

    if (activeId === overId) return

    const fromContainer =
      (active.data.current?.containerId as string | undefined) ??
      findContainer(activeId)
    const toContainer =
      (over.data.current?.containerId as string | undefined) ??
      findContainer(overId)

    if (!fromContainer || !toContainer || fromContainer === toContainer) return

    moveInstall(activeId, fromContainer, toContainer)
  }

  const handleSave = () => {
    if (hasEmptyNames || groups.length === 0) {
      setShowValidation(true)
      return
    }
    onSave(groups)
  }

  return (
    <Modal
      heading="Deployment plan"
      size="xl"
      primaryActionTrigger={{
        children: isSaving ? 'Saving...' : 'Save changes',
        onClick: handleSave,
        disabled: !canSave,
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
        <div className="flex flex-col gap-4">
          <Skeleton height="120px" />
          <Skeleton height="120px" />
        </div>
      ) : availableInstalls.length === 0 ? (
        <Banner theme="info">
          No installs found for this app. Create installs first to configure a
          deployment plan.
        </Banner>
      ) : (
        <DndContext
          sensors={sensors}
          collisionDetection={closestCorners}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
        >
          <div className="flex flex-col gap-4">
            {groups.length === 0 ? (
              <div className="flex flex-col items-center gap-4 px-6 py-10 border border-dashed border-cool-grey-300 dark:border-dark-grey-600 rounded-md">
                <EmptyState
                  variant="diagram"
                  emptyTitle="No deployment groups yet"
                  emptyMessage="Add a group, then drag installs into it."
                />
                <Button
                  variant="primary"
                  onClick={addGroup}
                  disabled={isDisabled}
                >
                  <Icon variant="PlusIcon" size={16} />
                  Add first group
                </Button>
              </div>
            ) : (
              <>
                {groups.map((group, index) => {
                  const groupInstalls = group.install_ids
                    .map((id) => installsById.get(id))
                    .filter((i): i is TInstall => !!i)

                  const nameError =
                    showValidation && !group.name.trim()
                      ? 'Group name is required'
                      : undefined

                  return (
                    <GroupEditor
                      key={group.id}
                      group={group}
                      index={index}
                      totalGroups={groups.length}
                      installs={groupInstalls}
                      unassignedInstalls={unassignedInstalls}
                      disabled={isDisabled}
                      nameError={nameError}
                      onUpdate={(updates) => updateGroup(group.id, updates)}
                      onAddInstalls={(installIds) =>
                        addInstallsToGroup(group.id, installIds)
                      }
                      onRemoveInstall={(installId) =>
                        removeInstallFromGroup(group.id, installId)
                      }
                      onMoveUp={() => moveGroup(group.id, -1)}
                      onMoveDown={() => moveGroup(group.id, 1)}
                      onDelete={() => deleteGroup(group.id)}
                    />
                  )
                })}
              </>
            )}

            <Button
              variant="secondary"
              onClick={addGroup}
              disabled={isDisabled}
              className="self-start"
            >
              <Icon variant="PlusIcon" size={16} />
              Add group
            </Button>

            <UnassignedSection
              installs={unassignedInstalls}
              containerId={UNASSIGNED_ID}
              disabled={isDisabled}
            />
          </div>

          {createPortal(
            <DragOverlay>
              {activeInstall ? (
                <div className="flex items-center gap-1.5 px-2 py-2 rounded-md bg-white dark:bg-dark-grey-800 border border-cool-grey-300 dark:border-dark-grey-600 shadow-lg cursor-grabbing">
                  <Icon
                    variant="DotsSixVerticalIcon"
                    size={16}
                    className="text-cool-grey-500"
                  />
                  <Text variant="body" className="truncate">
                    {activeInstall.name || activeInstall.id}
                  </Text>
                </div>
              ) : null}
            </DragOverlay>,
            document.body
          )}
        </DndContext>
      )}
    </Modal>
  )
}
