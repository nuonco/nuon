import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getAppConfigs } from '@/lib'
import type { TApp } from '@/types'
import { LoadAppConfigs } from './LoadAppConfigs'
import { CreateInstallFromAppContainer } from './CreateInstallFromAppContainer'

interface LoadAppConfigsContainerProps {
  app: TApp
  onSelectApp: (app: TApp | undefined) => void
  onClose: () => void
  formRef?: React.RefObject<HTMLFormElement>
  modalId?: string
  onLoadingChange?: (loading: boolean) => void
  onRegisterClearDraft?: (clearFn: () => void) => void
}

export const LoadAppConfigsContainer = ({
  app,
  onSelectApp,
  onClose,
  formRef,
  modalId,
  onLoadingChange,
  onRegisterClearDraft,
}: LoadAppConfigsContainerProps) => {
  const { org } = useOrg()
  const {
    data: configs,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['app-configs', org?.id, app.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id,
  })

  return (
    <LoadAppConfigs
      app={app}
      configs={configs}
      isLoading={isLoading}
      error={error}
      onSelectApp={onSelectApp}
    >
      {configs && configs.length > 0 && (
        <CreateInstallFromAppContainer
          app={app}
          configId={configs[0].id}
          onSelectApp={onSelectApp}
          onClose={onClose}
          formRef={formRef}
          modalId={modalId}
          onLoadingChange={onLoadingChange}
          onRegisterClearDraft={onRegisterClearDraft}
        />
      )}
    </LoadAppConfigs>
  )
}
