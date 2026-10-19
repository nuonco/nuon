import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import {
  installAppBranchGroup,
  installAppBranchId,
} from '@/components/branches/active-app-branch-connection'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { getAppBranch, moveInstallAppBranch } from '@/lib'
import type { TInstall } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { ChangeInstallGroupModal } from './ChangeInstallGroupModal'
import { labelChangesForGroupMove } from './label-changes'

interface IChangeInstallGroupContainer extends IModal {
  install: TInstall
  onSuccess?: () => void
}

export const ChangeInstallGroupContainer = ({
  install,
  onSuccess,
  ...props
}: IChangeInstallGroupContainer) => {
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()
  const appId = install.app_id ?? ''
  const branchId = installAppBranchId(install) ?? ''
  const branchName = install.app_branch?.name ?? branchId
  const currentGroup = installAppBranchGroup(install)
  const [targetGroup, setTargetGroup] = useState('')

  const {
    data: branch,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ['install-header-branch', org?.id, appId, branchId],
    queryFn: () =>
      getAppBranch({
        appId,
        branchId,
        orgId: org!.id,
        latestConfig: true,
      }),
    enabled: !!org?.id && !!appId && !!branchId,
  })

  const groups = branch
    ? (latestBranchConfig(branch)?.install_groups ?? [])
    : []
  const selectedGroup = groups.find((group) => group.name === targetGroup)
  const labelChanges = labelChangesForGroupMove(install.labels, selectedGroup)

  const { mutate: moveGroup, isPending } = useMutation({
    mutationFn: () =>
      moveInstallAppBranch({
        installId: install.id,
        orgId: org!.id,
        body: {
          app_branch_id: branchId,
          app_branch_group: targetGroup,
          ...(labelChanges.length > 0
            ? {
                labels: Object.fromEntries(
                  labelChanges.map((change) => [change.key, change.to])
                ),
              }
            : {}),
        },
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Install group changed" theme="success">
          <Text>
            Moved {install.name} to {targetGroup}.{' '}
            {labelChanges.length > 0 ? 'Group labels were updated. ' : ''}
            A reconciliation workflow has started.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({
        queryKey: ['install', org?.id, install.id],
      })
      queryClient.invalidateQueries({ queryKey: ['installs'] })
      queryClient.invalidateQueries({
        queryKey: ['install-current-app-branch-run', org?.id, install.id],
      })
      removeModal(props.modalId)
      onSuccess?.()
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Install group change failed" theme="error">
          <Text>
            {err?.error ||
              err?.description ||
              'Unable to move this install to the selected group.'}
          </Text>
        </Toast>
      )
    },
  })

  return (
    <ChangeInstallGroupModal
      install={install}
      branchName={branchName}
      currentGroup={currentGroup}
      groups={groups}
      targetGroup={targetGroup}
      isPending={isPending}
      isLoading={isLoading}
      loadFailed={isError}
      onSelectGroup={setTargetGroup}
      onConfirm={() => targetGroup && moveGroup()}
      {...props}
    />
  )
}

interface IChangeInstallGroupButton {
  install: TInstall
  onSuccess?: () => void
  compact?: boolean
  iconOnly?: boolean
}

export const ChangeInstallGroupButton = ({
  install,
  onSuccess,
  compact = false,
  iconOnly = false,
}: IChangeInstallGroupButton) => {
  const { addModal } = useSurfaces()

  if (!installAppBranchId(install)) return null

  return (
    <Button
      variant={iconOnly ? 'icon' : 'secondary'}
      size={iconOnly ? 'xs' : compact ? 'sm' : 'md'}
      aria-label={iconOnly ? 'Change install group' : undefined}
      tooltipProps={
        iconOnly ? { tipContent: 'Change install group' } : undefined
      }
      onClick={() =>
        addModal(
          <ChangeInstallGroupContainer
            install={install}
            onSuccess={onSuccess}
          />
        )
      }
    >
      <Icon
        variant={iconOnly ? 'PencilSimpleLineIcon' : 'UsersThreeIcon'}
        size={iconOnly ? 12 : 16}
      />
      {iconOnly ? null : 'Change group'}
    </Button>
  )
}
