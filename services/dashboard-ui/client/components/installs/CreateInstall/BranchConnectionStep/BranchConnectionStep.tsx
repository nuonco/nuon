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
import { addInstallLabels } from '@/lib'
import type { TAppBranch, TAppBranchConfig } from '@/types'

interface IBranchConnectionStep {
  branches: TAppBranch[]
  installId: string
  orgId: string
  onDone: () => void
}

const BranchGroupRow = ({
  group,
  installId,
  orgId,
}: {
  group: NonNullable<TAppBranchConfig['install_groups']>[number]
  installId: string
  orgId: string
}) => {
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const labelEntries = Object.entries(group.label_selector?.match_labels ?? {})
  const isLabels = labelEntries.length > 0
  const installIds = group.install_ids ?? []

  const { mutate: joinGroup, isPending } = useMutation({
    mutationFn: () =>
      addInstallLabels({
        installId,
        orgId,
        body: { labels: group.label_selector?.match_labels ?? {} },
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Labels added" theme="success">
          <Text>Install added to {group.name}.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['install'] })
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Label update failed" theme="error">
          <Text>{err?.error || 'Unable to add labels.'}</Text>
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
      {isLabels && (
        <Button
          variant="secondary"
          size="xs"
          onClick={() => joinGroup()}
          disabled={isPending}
        >
          {isPending ? 'Adding...' : 'Join group'}
        </Button>
      )}
    </div>
  )
}

export const BranchConnectionStep = ({
  branches,
  installId,
  orgId,
  onDone,
}: IBranchConnectionStep) => {
  if (branches.length === 0) {
    return (
      <div className="flex flex-col gap-4">
        <EmptyState
          variant="diagram"
          emptyTitle="No app branches configured"
          emptyMessage="You can connect this install to app branches later by adding labels."
        />
        <div className="flex justify-end">
          <Button variant="primary" onClick={onDone}>Done</Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <Text variant="base" weight="strong">
          Connect to app branches
        </Text>
        <Text variant="subtext" theme="neutral">
          Join a deployment group by adding its labels to this install. You can skip this and add labels later.
        </Text>
      </div>

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
                {groups.length === 0 ? (
                  <Text variant="subtext" theme="neutral">No install groups in this branch</Text>
                ) : (
                  groups.map((group, idx) => (
                    <BranchGroupRow
                      key={group.id || idx}
                      group={group}
                      installId={installId}
                      orgId={orgId}
                    />
                  ))
                )}
              </div>
            </Expand>
          )
        })}
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button variant="ghost" onClick={onDone}>Skip</Button>
        <Button variant="primary" onClick={onDone}>Done</Button>
      </div>
    </div>
  )
}
