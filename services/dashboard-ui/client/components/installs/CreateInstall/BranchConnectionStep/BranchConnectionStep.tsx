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
import { addInstallLabels, createBranchConfig } from '@/lib'
import type { TCreateBranchConfigRequest } from '@/lib/ctl-api/apps/branches/create-branch-config'
import type { TAppBranch, TAppBranchConfig } from '@/types'

interface IBranchConnectionStep {
  branches: TAppBranch[]
  installId: string
  orgId: string
  appId: string
  onDone: () => void
  onSkip: () => void
}

const buildConfigRequest = (
  config: TAppBranchConfig,
  targetGroupIndex: number,
  installId: string
): TCreateBranchConfigRequest => {
  const install_groups = (config.install_groups ?? []).map((g, index) => {
    const matchLabels = g.label_selector?.match_labels
    const isLabelGroup = !!matchLabels && Object.keys(matchLabels).length > 0
    const install_ids = isLabelGroup
      ? []
      : index === targetGroupIndex
        ? Array.from(new Set([...(g.install_ids ?? []), installId]))
        : g.install_ids ?? []

    return {
      name: g.name ?? '',
      install_ids,
      label_selector: isLabelGroup ? g.label_selector : undefined,
      order: index,
      max_parallel: g.max_parallel || 1,
      use_for_previews: g.use_for_previews || false,
    }
  })

  const request: TCreateBranchConfigRequest = { install_groups }

  if (config.connected_github_vcs_config) {
    request.connected_github_vcs_config = {
      vcs_connection_id: config.connected_github_vcs_config.vcs_connection_id || '',
      repo: config.connected_github_vcs_config.repo || '',
      branch: config.connected_github_vcs_config.branch || '',
      directory: config.connected_github_vcs_config.directory,
      path_filter: config.connected_github_vcs_config.path_filter,
    }
  } else if (config.public_git_vcs_config) {
    request.public_git_vcs_config = {
      repo: config.public_git_vcs_config.repo || '',
      branch: config.public_git_vcs_config.branch || '',
      directory: config.public_git_vcs_config.directory,
      path_filter: config.public_git_vcs_config.path_filter,
    }
  }

  return request
}

const BranchGroupRow = ({
  group,
  groupIndex,
  config,
  installId,
  orgId,
  appId,
  branchId,
}: {
  group: NonNullable<TAppBranchConfig['install_groups']>[number]
  groupIndex: number
  config: TAppBranchConfig
  installId: string
  orgId: string
  appId: string
  branchId: string
}) => {
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const labelEntries = Object.entries(group.label_selector?.match_labels ?? {})
  const isLabels = labelEntries.length > 0
  const installIds = group.install_ids ?? []
  const alreadyAdded = installIds.includes(installId)

  const invalidateBranch = () => {
    queryClient.invalidateQueries({ queryKey: ['app-branch-with-config', orgId, appId, branchId] })
    queryClient.invalidateQueries({ queryKey: ['install'] })
  }

  const { mutate: joinGroup, isPending: isJoining } = useMutation({
    mutationFn: () =>
      addInstallLabels({
        installId,
        orgId,
        body: { labels: group.label_selector?.match_labels ?? {} },
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Added to group" theme="success">
          <Text>Added this install to {group.name}.</Text>
        </Toast>
      )
      invalidateBranch()
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Add to group failed" theme="error">
          <Text>{err?.error || 'Unable to add labels.'}</Text>
        </Toast>
      )
    },
  })

  const { mutate: addToGroup, isPending: isAdding } = useMutation({
    mutationFn: () =>
      createBranchConfig({
        appId,
        branchId,
        orgId,
        request: buildConfigRequest(config, groupIndex, installId),
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Added to group" theme="success">
          <Text>Added this install to {group.name}.</Text>
        </Toast>
      )
      invalidateBranch()
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Add to group failed" theme="error">
          <Text>{err?.error || 'Unable to add this install to the group.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <div className="flex items-center justify-between gap-3 px-3 py-2.5 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700">
      <div className="flex items-center gap-2 flex-wrap min-w-0">
        <Text variant="body" weight="strong">{group.name}</Text>
        {labelEntries.map(([k, v]) => (
          <LabelBadge key={k} labelKey={k} labelValue={v} size="sm" variant="code" />
        ))}
        {!isLabels && installIds.length > 0 && (
          <Text variant="subtext" theme="neutral">
            {installIds.length} install{installIds.length !== 1 ? 's' : ''} by ID
          </Text>
        )}
      </div>

      {alreadyAdded ? (
        <span className="flex shrink-0 items-center gap-1 text-xs text-green-600 dark:text-green-400">
          <Icon variant="CheckIcon" size={14} />
          Added
        </span>
      ) : isLabels ? (
        <Button variant="secondary" onClick={() => joinGroup()} disabled={isJoining}>
          {isJoining ? 'Adding...' : 'Join group'}
        </Button>
      ) : (
        <Button variant="secondary" onClick={() => addToGroup()} disabled={isAdding}>
          {isAdding ? 'Adding...' : 'Add to group'}
        </Button>
      )}
    </div>
  )
}

export const BranchConnectionStep = ({
  branches,
  installId,
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
          <Button variant="primary" onClick={onDone}>Done</Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Text variant="subtext" theme="neutral">
        Add this install to a deployment group so app branch runs deploy to it. You can skip this and do it later.
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
                  <Text variant="body" weight="strong">{branch.name}</Text>
                  <Badge size="sm" theme="info">{groups.length} group{groups.length !== 1 ? 's' : ''}</Badge>
                </div>
              }
              headerClassName="!px-3"
              className="border rounded-md"
            >
              <div className="flex flex-col gap-2 p-3 border-t">
                {!latestConfig || groups.length === 0 ? (
                  <Text variant="subtext" theme="neutral">No install groups in this branch</Text>
                ) : (
                  groups.map((group, idx) => (
                    <BranchGroupRow
                      key={group.id || idx}
                      group={group}
                      groupIndex={idx}
                      config={latestConfig}
                      installId={installId}
                      orgId={orgId}
                      appId={appId}
                      branchId={branch.id || ''}
                    />
                  ))
                )}
              </div>
            </Expand>
          )
        })}
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button variant="ghost" onClick={onSkip}>Skip for now</Button>
        <Button variant="primary" onClick={onDone}>Go to install</Button>
      </div>
    </div>
  )
}
