import type { IModal } from '@/components/surfaces/Modal'
import { InstallProvider } from '@/providers/install-provider'
import { RunAdhocActionModal } from '@/components/installs/management/RunAdhocAction'

export const InstallAdhocActionModal = ({ installId, ...modalProps }: { installId: string } & IModal) => (
  <InstallProvider installId={installId}>
    <RunAdhocActionModal {...modalProps} />
  </InstallProvider>
)
