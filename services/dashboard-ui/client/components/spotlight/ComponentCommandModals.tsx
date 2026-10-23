import type { IModal } from '@/components/surfaces/Modal'
import { InstallProvider } from '@/providers/install-provider'
import { AppProvider } from '@/providers/app-provider'
import { TeardownComponentModal } from '@/components/install-components/management/TeardownComponent'
import { BuildComponentModal } from '@/components/components/management/BuildComponent'
import type { TComponent } from '@/types'

type IInstallComponentCommandModal = {
  installId: string
  component: TComponent
} & IModal
type IAppComponentCommandModal = {
  appId: string
  component: TComponent
} & IModal

const withInstallProvider = (
  Modal: React.ComponentType<IModal & { component: TComponent }>
) => {
  const Wrapped = ({
    installId,
    component,
    ...modalProps
  }: IInstallComponentCommandModal) => (
    <InstallProvider installId={installId}>
      <Modal component={component} {...modalProps} />
    </InstallProvider>
  )
  Wrapped.displayName = `ComponentCommand(${Modal.displayName || Modal.name})`
  return Wrapped
}

const withAppProvider = (
  Modal: React.ComponentType<IModal & { component: TComponent }>
) => {
  const Wrapped = ({
    appId,
    component,
    ...modalProps
  }: IAppComponentCommandModal) => (
    <AppProvider appId={appId}>
      <Modal component={component} {...modalProps} />
    </AppProvider>
  )
  Wrapped.displayName = `ComponentCommand(${Modal.displayName || Modal.name})`
  return Wrapped
}

export const SpotlightTeardownComponentModal = withInstallProvider(
  TeardownComponentModal
)
export const SpotlightBuildComponentModal = withAppProvider(BuildComponentModal)
