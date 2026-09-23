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
  // installs this branch owns: the population `all installs` and label
  // selectors resolve to
  availableInstalls: TInstall[]
  // every install on the app; naming one by ID moves it onto this branch when
  // the plan is saved. Defaults to the branch's own installs.
  appInstalls?: TInstall[]
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
  appInstalls,
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
  const [newGroupId, setNewGroupId] = useState<string | null>(null)
  const newGroupRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!newGroupId) return
    newGroupRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, [newGroupId])

  const selectableInstalls = appInstalls ?? availableInstalls

  const installsById = useMemo(() => {
    const map: Record<string, TInstall> = {}
    for (const i of selectableInstalls) map[i.id] = i
    return map
  }, [selectableInstalls])

  const previewConfig = useMemo<TAppBranchConfig>(() => ({
    install_groups: groups.map((g) => ({
      id: g.id,
      name: g.name || `Group ${g.order + 1}`,
      install_ids: g.selection_mode === 'manual' ? g.install_ids : [],
      label_selector: g.selection_mode === 'labels' ? g.label_selector : undefined,
      all_installs: g.selection_mode === 'all',
      max_parallel: g.max_parallel,
    })),
  } as TAppBranchConfig), [groups])

  const assignedInstallIds = useMemo(() => {
    const assigned = new Set<string>()
    groups.forEach((g) => {
      if (g.selection_mode === 'all') {
        availableInstalls.forEach((i) => assigned.add(i.id))
      } else if (g.selection_mode === 'labels') {
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

  const pickableInstalls = useMemo(
    () => selectableInstalls.filter((i) => !assignedInstallIds.has(i.id)),
    [selectableInstalls, assignedInstallIds]
  )

  // Installs this plan will act on once saved: the ones the branch owns, plus
  // the ones it claims by naming them.
  const planInstalls = useMemo(() => {
    const owned = new Set(availableInstalls.map((i) => i.id))
    const claimed = selectableInstalls.filter(
      (i) =>
        !owned.has(i.id) &&
        groups.some(
          (g) => g.selection_mode === 'manual' && g.install_ids.includes(i.id)
        )
    )
    return [...availableInstalls, ...claimed]
  }, [availableInstalls, selectableInstalls, groups])

  const overlappingInstalls = useMemo(() => {
    const matches = new Map<string, number>()
    groups.forEach((g) => {
      const matchedIds =
        g.selection_mode === 'all'
          ? planInstalls.map((i) => i.id)
          : g.selection_mode === 'labels'
            ? planInstalls
                .filter((i) => matchesSelector(i.labels, g.label_selector))
                .map((i) => i.id)
            : g.install_ids.filter((id) =>
                planInstalls.some((i) => i.id === id)
              )
      matchedIds.forEach((id) => matches.set(id, (matches.get(id) ?? 0) + 1))
    })
    return planInstalls.filter((i) => (matches.get(i.id) ?? 0) > 1)
  }, [groups, planInstalls])

  const groupContentError = (g: IInstallGroup): string | undefined => {
    if (g.selection_mode === 'all') return undefined
    if (g.selection_mode === 'labels') {
      if (!g.label_selector?.match_labels || Object.keys(g.label_selector.match_labels).length === 0) {
        return 'Add at least one label to match installs.'
      }
    } else if (g.install_ids.length === 0) {
      return 'Add at least one install.'
    }
    return undefined
  }

  const hasErrors =
    groups.some((g) => !g.name.trim() || !!groupContentError(g)) ||
    overlappingInstalls.length > 0
  const canSave = !isSaving && !loadingInstalls && !hasErrors
  const isDisabled = isSaving || loadingInstalls

  const saveDisabledReason = (() => {
    if (canSave || isSaving || loadingInstalls) return undefined
    if (overlappingInstalls.length > 0)
      return 'Each install must match exactly one install group.'
    const needsName = groups.some((g) => !g.name.trim())
    const needsInstalls = groups.some((g) => !!groupContentError(g))
    if (needsName && needsInstalls)
      return 'Every group needs a name and installs to target.'
    if (needsInstalls) return 'Every group needs installs or matching labels.'
    if (needsName) return 'Every group needs a name.'
    return undefined
  })()

  const updateGroup = (id: string, updates: Partial<IInstallGroup>) => {
    setGroups((curr) =>
      curr.map((g) => (g.id === id ? { ...g, ...updates } : g))
    )
  }

  const addGroup = () => {
    const group = newGroup(
      groups.length,
      selectableInstalls.length === 0 ? 'all' : 'manual'
    )
    setGroups((curr) => [...curr, group])
    setNewGroupId(group.id)
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
    if (hasErrors) {
      setShowValidation(true)
      return
    }
    onSave(groups, postDeployRunbookIds)
  }

  const canAddGroup = !loadingInstalls

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
      primaryActionTooltip={
        saveDisabledReason ? (
          <Text variant="subtext">{saveDisabledReason}</Text>
        ) : undefined
      }
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
      ) : (
        <div className="flex flex-col gap-6">
          <Text variant="subtext" theme="neutral">
            Groups deploy top to bottom. Installs in a group deploy together, up
            to its max parallel. Any install left unassigned is skipped.
          </Text>

          {selectableInstalls.length === 0 && (
            <Banner theme="info">
              This app has no installs yet. Groups that match on labels or take
              all installs pick them up as they are created — a group with a
              hand-picked list needs installs to exist first.
            </Banner>
          )}

          {overlappingInstalls.length > 0 && (
            <Banner theme="error">
              {overlappingInstalls.length === 1
                ? `${overlappingInstalls[0].name} matches more than one install group.`
                : `${overlappingInstalls.length} installs match more than one install group.`}{' '}
              Each install must match exactly one group before this deployment
              plan can be saved.
            </Banner>
          )}

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
                const contentError = groupContentError(group)

                return (
                  <div
                    key={group.id}
                    ref={group.id === newGroupId ? newGroupRef : null}
                  >
                    <GroupEditor
                      group={group}
                      index={index}
                      totalGroups={groups.length}
                      autoFocusName={group.id === newGroupId}
                      availableInstalls={availableInstalls}
                      pickableInstalls={pickableInstalls}
                      installsById={installsById}
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
              <div className="flex items-baseline gap-2 mb-1">
                <Text variant="base" weight="strong">Orphaned</Text>
                <Text variant="subtext" theme="neutral">
                  — {unassignedInstalls.length} install{unassignedInstalls.length !== 1 ? 's' : ''} won&apos;t receive updates
                </Text>
              </div>
              <Text variant="subtext" theme="neutral" className="mb-2">
                These installs belong to this branch but don&apos;t match any
                group. They will not be updated when this branch runs.
              </Text>
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
