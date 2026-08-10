import { useEffect, useMemo, useRef, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { DeploymentPlanGraph } from '@/components/branches/DeploymentPlanGraph'
import { PostDeployRunbooksPicker } from '@/components/branches/PostDeployRunbooksPicker'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks/get-runbooks'
import type { TInstall, TAppBranchConfig } from '@/types'
import { matchesSelector } from '@/components/match/matches'
import { GroupEditor } from './GroupEditor'
import { InstallRow } from './InstallRow'
import { newGroup } from './lib'
import type { IInstallGroup } from './types'

interface IDeploymentPlanEditor extends Omit<IModal, 'onSubmit'> {
  initialGroups: IInstallGroup[]
  availableInstalls: TInstall[]
  loadingInstalls: boolean
  isSaving: boolean
  labelColors?: Record<string, string>
  orgId: string
  runbooks: TRunbook[]
  loadingRunbooks: boolean
  initialPostDeployRunbookIds: string[]
  onSave: (groups: IInstallGroup[], postDeployRunbookIds: string[]) => void
  onCancel: () => void
}

export const DeploymentPlanEditor = ({
  initialGroups,
  availableInstalls,
  loadingInstalls,
  isSaving,
  labelColors,
  orgId,
  runbooks,
  loadingRunbooks,
  initialPostDeployRunbookIds,
  onSave,
  onCancel,
  ...props
}: IDeploymentPlanEditor) => {
  const [groups, setGroups] = useState<IInstallGroup[]>(initialGroups)
  const [postDeployRunbookIds, setPostDeployRunbookIds] = useState<string[]>(
    initialPostDeployRunbookIds
  )
  const [showValidation, setShowValidation] = useState(false)
  const [scrollToId, setScrollToId] = useState<string | null>(null)
  const newGroupRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!scrollToId) return
    newGroupRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' })
    setScrollToId(null)
  }, [scrollToId])

  const installsById = useMemo(() => {
    const map: Record<string, TInstall> = {}
    for (const i of availableInstalls) map[i.id] = i
    return map
  }, [availableInstalls])

  const previewConfig = useMemo<TAppBranchConfig>(() => ({
    install_groups: groups.map((g) => ({
      id: g.id,
      name: g.name || `Group ${g.order + 1}`,
      install_ids: g.selection_mode === 'manual' ? g.install_ids : [],
      label_selector: g.selection_mode === 'labels' ? g.label_selector : undefined,
      max_parallel: g.max_parallel,
      use_for_previews: g.use_for_previews,
    })),
  } as TAppBranchConfig), [groups])

  const assignedInstallIds = useMemo(() => {
    const assigned = new Set<string>()
    groups.forEach((g) => {
      if (g.use_for_previews) return
      if (g.selection_mode === 'labels') {
        const matchLabels = g.label_selector?.match_labels
        if (matchLabels && Object.keys(matchLabels).length > 0) {
          availableInstalls.forEach((i) => {
            if (matchesSelector(i.labels, g.label_selector)) assigned.add(i.id)
          })
        }
      } else {
        g.install_ids.forEach((id) => assigned.add(id))
      }
    })
    return assigned
  }, [groups, availableInstalls])

  const unassignedInstalls = useMemo(
    () => availableInstalls.filter((i) => !assignedInstallIds.has(i.id)),
    [availableInstalls, assignedInstallIds]
  )

  const groupContentError = (g: IInstallGroup): string | undefined => {
    if (g.selection_mode === 'labels') {
      if (!g.label_selector?.match_labels || Object.keys(g.label_selector.match_labels).length === 0) {
        return 'Add at least one label to match installs.'
      }
    } else if (g.install_ids.length === 0) {
      return 'Add at least one install.'
    }
    return undefined
  }

  const hasErrors = groups.some((g) => !g.name.trim() || !!groupContentError(g))
  const canSave = !isSaving && !loadingInstalls && groups.length > 0 && !hasErrors
  const isDisabled = isSaving || loadingInstalls

  const updateGroup = (id: string, updates: Partial<IInstallGroup>) => {
    setGroups((curr) =>
      curr.map((g) => (g.id === id ? { ...g, ...updates } : g))
    )
  }

  const addGroup = () => {
    const group = newGroup(groups.length)
    setGroups((curr) => [...curr, group])
    setScrollToId(group.id)
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
    if (hasErrors || groups.length === 0) {
      setShowValidation(true)
      return
    }
    onSave(groups, postDeployRunbookIds)
  }

  const canAddGroup = !loadingInstalls && availableInstalls.length > 0

  return (
    <Modal
      heading="Deployment plan"
      size="xl"
      className="!max-w-[1200px]"
      footerActions={
        canAddGroup ? (
          <Button variant="secondary" onClick={addGroup} disabled={isDisabled}>
            <Icon variant="PlusIcon" size={16} />
            Add group
          </Button>
        ) : undefined
      }
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
        <div className="flex flex-col gap-6">
          <Text variant="subtext" theme="neutral">
            Groups deploy top to bottom. Installs in a group deploy together, up
            to its max parallel. Any install left unassigned is skipped.
          </Text>

          {groups.length >= 2 && (
            <DeploymentPlanGraph config={previewConfig} installsById={installsById} orgId={orgId} />
          )}

          {groups.length === 0 ? (
            <EmptyState
              variant="table"
              emptyTitle="No install groups yet"
              emptyMessage="Use Add group below to create your first group, then assign installs to it."
            />
          ) : (
            <>
              {groups.map((group, index) => {
                const nameError =
                  showValidation && !group.name.trim()
                    ? 'Group name is required'
                    : undefined
                const contentError = showValidation ? groupContentError(group) : undefined

                return (
                  <div
                    key={group.id}
                    ref={group.id === scrollToId ? newGroupRef : null}
                  >
                    <GroupEditor
                      group={group}
                      index={index}
                      totalGroups={groups.length}
                      availableInstalls={availableInstalls}
                      unassignedInstalls={unassignedInstalls}
                      labelColors={labelColors}
                      disabled={isDisabled}
                      nameError={nameError}
                      contentError={contentError}
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
                  </div>
                )
              })}
            </>
          )}

          {groups.length > 0 && (
            <div className="border-t pt-4">
              <PostDeployRunbooksPicker
                runbooks={runbooks}
                loadingRunbooks={loadingRunbooks}
                selectedRunbookIds={postDeployRunbookIds}
                onChange={setPostDeployRunbookIds}
                disabled={isDisabled}
              />
            </div>
          )}

          {groups.length > 0 && unassignedInstalls.length > 0 && (
            <div className="border-t pt-4">
              <div className="flex items-baseline gap-2 mb-2">
                <Text variant="base" weight="strong">Unassigned</Text>
                <Text variant="subtext" theme="neutral">
                  — {unassignedInstalls.length} install{unassignedInstalls.length !== 1 ? 's' : ''} won&apos;t deploy
                </Text>
              </div>
              <div className="flex flex-col gap-1.5">
                {unassignedInstalls.map((install) => (
                  <InstallRow key={install.id} install={install} labelColors={labelColors} />
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </Modal>
  )
}
