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
import { resolveInstallGroupMembership } from '@/components/branches/install-group-membership'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks/get-runbooks'
import type { TInstall, TAppBranchConfig } from '@/types'
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
  const [newGroupId, setNewGroupId] = useState<string | null>(null)
  const newGroupRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!newGroupId) return
    newGroupRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, [newGroupId])

  const installsById = useMemo(() => {
    const map: Record<string, TInstall> = {}
    for (const i of availableInstalls) map[i.id] = i
    return map
  }, [availableInstalls])

  const previewGroups = useMemo(
    () =>
      groups.map((g) => ({
        id: g.id,
        name: g.name || `Group ${g.order + 1}`,
        label_selector:
          g.selection_mode === 'labels' ? g.label_selector : undefined,
        default: g.selection_mode === 'default',
        max_parallel: g.max_parallel,
      })),
    [groups]
  )

  const previewConfig = useMemo<TAppBranchConfig>(
    () => ({
      install_groups: previewGroups,
    }),
    [previewGroups]
  )

  const membership = useMemo(() => {
    return resolveInstallGroupMembership(availableInstalls, previewGroups)
  }, [previewGroups, availableInstalls])

  const { unassignedInstalls, overlappingInstalls } = membership
  const duplicateGroupNames = useMemo(() => {
    const seen = new Set<string>()
    const duplicates = new Set<string>()
    groups.forEach((group) => {
      const name = group.name.trim()
      if (!name) return
      if (seen.has(name)) duplicates.add(name)
      seen.add(name)
    })
    return duplicates
  }, [groups])

  const groupContentError = (g: IInstallGroup): string | undefined => {
    if (g.selection_mode === 'default') return undefined
    if (
      !g.label_selector?.match_labels ||
      Object.keys(g.label_selector.match_labels).length === 0
    ) {
      return 'Add at least one label to match installs.'
    }
    return undefined
  }

  const hasErrors =
    groups.some((g) => !g.name.trim() || !!groupContentError(g)) ||
    duplicateGroupNames.size > 0 ||
    groups.filter((g) => g.selection_mode === 'default').length > 1 ||
    overlappingInstalls.length > 0
  const canSave = !isSaving && !loadingInstalls && !hasErrors
  const isDisabled = isSaving || loadingInstalls

  const saveDisabledReason = (() => {
    if (canSave || isSaving || loadingInstalls) return undefined
    if (overlappingInstalls.length > 0)
      return 'Each install must match exactly one install group.'
    if (groups.filter((g) => g.selection_mode === 'default').length > 1)
      return 'Only one default group is allowed.'
    if (duplicateGroupNames.size > 0) return 'Every group needs a unique name.'
    const needsName = groups.some((g) => !g.name.trim())
    const needsLabels = groups.some((g) => !!groupContentError(g))
    if (needsName && needsLabels)
      return 'Every group needs a name and labels to target installs.'
    if (needsLabels) return 'Every labels group needs at least one label.'
    if (needsName) return 'Every group needs a name.'
    return undefined
  })()

  const updateGroup = (id: string, updates: Partial<IInstallGroup>) => {
    setGroups((curr) =>
      curr.map((g) => (g.id === id ? { ...g, ...updates } : g))
    )
  }

  const addGroup = () => {
    const group = newGroup(groups.length, 'labels')
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
            <DeploymentPlanGraph
              config={previewConfig}
              installsById={installsById}
              orgId={orgId}
            />
          )}

          {groups.length === 0 ? (
            <EmptyState
              variant="table"
              emptyTitle="No install groups yet"
              emptyMessage="Use Add group below to create your first group."
            />
          ) : (
            <>
              {groups.map((group, index) => {
                const nameError =
                  showValidation && !group.name.trim()
                    ? 'Group name is required'
                    : duplicateGroupNames.has(group.name.trim())
                      ? 'Group name must be unique'
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
                      resolvedInstalls={membership.installsByGroup[index]}
                      labelColors={labelColors}
                      disabled={isDisabled}
                      nameError={nameError}
                      contentError={contentError}
                      onUpdate={(updates) => updateGroup(group.id, updates)}
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
                <Text variant="base" weight="strong">
                  Orphaned
                </Text>
                <Text variant="subtext" theme="neutral">
                  — {unassignedInstalls.length} install
                  {unassignedInstalls.length !== 1 ? 's' : ''} won&apos;t
                  receive updates
                </Text>
              </div>
              <Text variant="subtext" theme="neutral" className="mb-2">
                These installs belong to this branch but don&apos;t match any
                group. They will not be updated when this branch runs.
              </Text>
              <div className="flex flex-col gap-1.5">
                {unassignedInstalls.map((install) => (
                  <InstallRow
                    key={install.id}
                    install={install}
                    labelColors={labelColors}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </Modal>
  )
}
