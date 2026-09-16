import { useState } from 'react'
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { getAppBranches, moveInstallAppBranch } from '@/lib'
import type { TAppBranch, TInstall } from '@/types'
import { ChangeAppBranchModal } from './ChangeAppBranchModal'

interface IChangeAppBranchContainer extends IModal {
  install: TInstall
  onSuccess?: () => void
}

export const ChangeAppBranchContainer = ({
  install,
  onSuccess,
  ...props
}: IChangeAppBranchContainer) => {
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()
  const appId = install.app_id ?? ''
  const [targetBranch, setTargetBranch] = useState<TAppBranch | null>(null)

  const { data: branchList } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-branches', org?.id, appId],
    queryFn: () => getAppBranches({ appId, orgId: org!.id }),
    enabled: !!org?.id && !!appId,
  })

  const { mutate: moveInstall, isPending } = useMutation({
    mutationFn: () =>
      moveInstallAppBranch({
        installId: install.id,
        orgId: org!.id,
        body: { app_branch_id: targetBranch!.id },
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Branch changed" theme="success">
          <Text>
            Moved {install.name} to {targetBranch?.name}. A reconciliation
            workflow has started.
          </Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['install', org?.id, install.id] })
      queryClient.invalidateQueries({ queryKey: ['installs'] })
      removeModal(props.modalId)
      onSuccess?.()
    },
    onError: (err: any) => {
      addToast(
        <Toast heading="Branch change failed" theme="error">
          <Text>
            {err?.error || err?.description || 'Unable to move install to the selected branch.'}
          </Text>
        </Toast>
      )
    },
  })

  return (
    <ChangeAppBranchModal
      install={install}
      targetBranch={targetBranch}
      branches={branchList?.data ?? []}
      isPending={isPending}
      onSelectBranch={setTargetBranch}
      onConfirm={() => targetBranch && moveInstall()}
      {...props}
    />
  )
}

interface IChangeAppBranchButton {
  install: TInstall
  onSuccess?: () => void
}

export const ChangeAppBranchButton = ({
  install,
  onSuccess,
}: IChangeAppBranchButton) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      variant="secondary"
      onClick={() =>
        addModal(
          <ChangeAppBranchContainer
            install={install}
            onSuccess={onSuccess}
          />
        )
      }
    >
      <Icon variant="GitBranchIcon" size={16} />
      Change branch
    </Button>
  )
}
