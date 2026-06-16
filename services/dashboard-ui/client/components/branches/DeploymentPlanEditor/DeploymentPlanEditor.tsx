import { useMemo, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TInstall } from '@/types'
import { GroupEditor } from './GroupEditor'
import { newGroup } from './lib'
import type { IInstallGroup } from './types'

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

  const assignedInstallIds = useMemo(() => {
    const assigned = new Set<string>()
    groups.forEach((g) => {
      if (!g.use_for_previews) {
        g.install_ids.forEach((id) => assigned.add(id))
      }
    })
    return assigned
  }, [groups])

  const unassignedInstalls = useMemo(
    () => availableInstalls.filter((i) => !assignedInstallIds.has(i.id)),
    [availableInstalls, assignedInstallIds]
  )

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
    setGroups((curr) =>
      curr.map((g) => {
        if (g.id === groupId) {
          const merged = [...g.install_ids]
          installIds.forEach((id) => {
            if (!merged.includes(id)) merged.push(id)
          })
          return { ...g, install_ids: merged }
        }
        return g
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
        <div className="flex flex-col gap-4">
          {groups.length === 0 ? (
            <div className="flex items-center justify-between gap-4 px-4 py-4 border border-dashed border-cool-grey-300 dark:border-dark-grey-600 rounded-md">
              <Text variant="subtext" theme="neutral">
                No deployment groups yet. Add a group, then assign installs to it.
              </Text>
              <Button
                variant="primary"
                size="sm"
                onClick={addGroup}
                disabled={isDisabled}
                className="shrink-0"
              >
                <Icon variant="PlusIcon" size={16} />
                Add first group
              </Button>
            </div>
          ) : (
            groups.map((group, index) => {
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
                  availableInstalls={availableInstalls}
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
            })
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

          {unassignedInstalls.length > 0 && (
            <div className="border-t border-cool-grey-200 dark:border-dark-grey-700 pt-4">
              <div className="flex items-baseline gap-2 mb-2">
                <Text variant="base" weight="strong">Unassigned</Text>
                <Text variant="subtext" theme="neutral">
                  — {unassignedInstalls.length} install{unassignedInstalls.length !== 1 ? 's' : ''} won&apos;t deploy
                </Text>
              </div>
              <div className="flex flex-col gap-1.5">
                {unassignedInstalls.map((install) => (
                  <div
                    key={install.id}
                    className="flex items-center gap-1.5 px-2 py-2 rounded-md bg-cool-grey-50 dark:bg-dark-grey-900"
                  >
                    <Text variant="body" className="truncate">
                      {install.name || install.id}
                    </Text>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </Modal>
  )
}
