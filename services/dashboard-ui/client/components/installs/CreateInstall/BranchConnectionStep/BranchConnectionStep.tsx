import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useToast } from '@/hooks/use-toast'
import { addInstallLabels, moveInstallAppBranch } from '@/lib'
import type { TAppBranch, TAppBranchConfig } from '@/types'

interface IBranchConnectionStep {
  branches: TAppBranch[]
  installId: string
  installLabels?: Record<string, string>
  orgId: string
  appId: string
  onDone: () => void
  onSkip: () => void
}

const BranchGroupRow = ({
  group,
  installId,
  installLabels,
  orgId,
  appId,
  branchId,
  onConnected,
}: {
  group: NonNullable<TAppBranchConfig['install_groups']>[number]
  installId: string
  installLabels?: Record<string, string>
  orgId: string
  appId: string
  branchId: string
  onConnected: () => void
}) => {
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const isDefault = !!group.default
  const labelEntries = Object.entries(group.label_selector?.match_labels ?? {})
  const isLabels = !isDefault && labelEntries.length > 0
  const alreadyAddedByLabels =
    isLabels && labelEntries.every(([k, v]) => installLabels?.[k] === v)
  const conflictsWithLabels =
    isLabels &&
    !alreadyAddedByLabels &&
    labelEntries.some(
      ([k, v]) => installLabels?.[k] !== undefined && installLabels[k] !== v
    )
  const invalidateBranch = () => {
    queryClient.invalidateQueries({
      queryKey: ['app-branch-with-config', orgId, appId, branchId],
    })
    queryClient.invalidateQueries({ queryKey: ['install'] })
  }

  const { mutate: joinGroup, isPending: isJoining } = useMutation({
    mutationFn: async () => {
      if (isLabels && !alreadyAddedByLabels) {
        await addInstallLabels({
          installId,
          orgId,
          body: { labels: group.label_selector?.match_labels ?? {} },
        })
      }
      return moveInstallAppBranch({
        installId,
        orgId,
        body: { app_branch_id: branchId },
      })
    },
    onSuccess: () => {
      addToast(
        <Toast heading="Connected to app branch" theme="success">
          <Text>Connected this install through {group.name}.</Text>
        </Toast>
      )
      invalidateBranch()
      onConnected()
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Add to group failed" theme="error">
          <Text>{err?.error || 'Unable to add labels.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <div className="flex items-center justify-between gap-3 px-3 py-2.5 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700">
      <div className="flex items-center gap-2 flex-wrap min-w-0">
        <Text variant="body" weight="strong">
          {group.name}
        </Text>
        {labelEntries.map(([k, v]) => (
          <LabelBadge key={k} labelKey={k} labelValue={v} size="sm" />
        ))}
        {isDefault && (
          <Text variant="subtext" theme="neutral">
            All remaining installs
          </Text>
        )}
      </div>

      {isDefault || isLabels ? (
        <Button
          variant="secondary"
          onClick={() => joinGroup()}
          disabled={isJoining || conflictsWithLabels}
          tooltipProps={
            conflictsWithLabels
              ? {
                  tipContent:
                    'Conflicts with labels already applied to this install',
                }
              : undefined
          }
        >
          {isJoining ? 'Connecting...' : 'Connect'}
        </Button>
      ) : null}
    </div>
  )
}

export const BranchConnectionStep = ({
  branches,
  installId,
  installLabels,
  orgId,
  appId,
  onDone,
  onSkip,
}: IBranchConnectionStep) => {
  if (branches.length === 0) {
    return (
      <div className="flex flex-col gap-4">
        <EmptyState
          variant="diagram"
          emptyTitle="No app branches configured"
          emptyMessage="You can connect this install to app branches later."
        />
        <div className="flex justify-end">
          <Button variant="primary" onClick={onDone}>
            Done
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Text variant="subtext" theme="neutral">
        Add this install to an install group so app branch runs deploy to it.
        You can skip this and do it later.
      </Text>

      <div className="flex flex-col gap-3">
        {branches.map((branch) => {
          const latestConfig = branch.configs?.at(0)
          const groups = latestConfig?.install_groups ?? []

          return (
            <Expand
              key={branch.id}
              id={`branch-${branch.id}`}
              heading={
                <div className="flex items-center gap-2">
                  <Icon variant="GitBranchIcon" size={14} />
                  <Text variant="body" weight="strong">
                    {branch.name}
                  </Text>
                  <Badge size="sm" theme="info">
                    {groups.length} group{groups.length !== 1 ? 's' : ''}
                  </Badge>
                </div>
              }
              headerClassName="!px-3"
              className="border rounded-md"
            >
              <div className="flex flex-col gap-2 p-3 border-t">
                {!latestConfig || groups.length === 0 ? (
                  <Text variant="subtext" theme="neutral">
                    No install groups in this branch
                  </Text>
                ) : (
                  groups.map((group, idx) => (
                    <BranchGroupRow
                      key={group.id ?? idx}
                      group={group}
                      installId={installId}
                      installLabels={installLabels}
                      orgId={orgId}
                      appId={appId}
                      branchId={branch.id || ''}
                      onConnected={onDone}
                    />
                  ))
                )}
              </div>
            </Expand>
          )
        })}
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button variant="ghost" onClick={onSkip}>
          Skip for now
        </Button>
        <Button variant="primary" onClick={onDone}>
          Go to install
        </Button>
      </div>
    </div>
  )
}
