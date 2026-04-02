import type { IModal } from '@/components/surfaces/Modal'
import { InstallProvider } from '@/providers/install-provider'
import { RunAdhocActionModal } from '@/components/installs/management/RunAdhocAction'
import { EditInputsFormModal } from '@/components/installs/management/EditInputs'
import { SyncSecretsModal } from '@/components/installs/management/SyncSecrets'
import { ReprovisionModal } from '@/components/installs/management/Reprovision'
import { ReprovisionSandboxModal } from '@/components/sandbox/management/ReprovisionSandbox'
import { DeployAllComponentsModal } from '@/components/install-components/management/DeployAllComponents'

type IInstallCommandModal = { installId: string } & IModal

const withInstallProvider = (Modal: React.ComponentType<IModal>) => {
  const Wrapped = ({ installId, ...modalProps }: IInstallCommandModal) => (
    <InstallProvider installId={installId}>
      <Modal {...modalProps} />
    </InstallProvider>
  )
  Wrapped.displayName = `InstallCommand(${Modal.displayName || Modal.name})`
  return Wrapped
}

export const InstallAdhocActionModal = withInstallProvider(RunAdhocActionModal)
export const InstallEditInputsModal = withInstallProvider(EditInputsFormModal)
export const InstallSyncSecretsModal = withInstallProvider(SyncSecretsModal)
export const InstallReprovisionModal = withInstallProvider(ReprovisionModal)
export const InstallReprovisionSandboxModal = withInstallProvider(ReprovisionSandboxModal)
export const InstallDeployAllComponentsModal = withInstallProvider(DeployAllComponentsModal)
