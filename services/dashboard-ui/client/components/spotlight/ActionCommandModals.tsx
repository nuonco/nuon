import type { IModal } from '@/components/surfaces/Modal'
import { InstallProvider } from '@/providers/install-provider'
import { InstallActionManualRunModal } from '@/components/actions/InstallActionManualRun'
import type { TAction } from '@/types'

type IActionCommandModal = {
  installId: string
  action: TAction
  actionConfigId: string
} & IModal

export const SpotlightRunActionModal = ({
  installId,
  action,
  actionConfigId,
  ...modalProps
}: IActionCommandModal) => (
  <InstallProvider installId={installId}>
    <InstallActionManualRunModal
      action={action}
      actionConfigId={actionConfigId}
      {...modalProps}
    />
  </InstallProvider>
)
