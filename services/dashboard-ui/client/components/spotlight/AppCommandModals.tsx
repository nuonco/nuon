import type { IModal } from '@/components/surfaces/Modal'
import { AppProvider } from '@/providers/app-provider'
import { BuildAllComponentsModal } from '@/components/components/management/BuildAllComponents'

type IAppCommandModal = { appId: string } & IModal

const withAppProvider = (Modal: React.ComponentType<IModal>) => {
  const Wrapped = ({ appId, ...modalProps }: IAppCommandModal) => (
    <AppProvider appId={appId}>
      <Modal {...modalProps} />
    </AppProvider>
  )
  Wrapped.displayName = `AppCommand(${Modal.displayName || Modal.name})`
  return Wrapped
}

export const AppBuildAllComponentsModal = withAppProvider(BuildAllComponentsModal)
