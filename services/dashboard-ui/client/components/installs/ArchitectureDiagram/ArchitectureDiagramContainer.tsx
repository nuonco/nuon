import { useQuery } from '@tanstack/react-query'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallComponents, getInstallStack, getAppConfig, getInstallAppPermissionsConfig } from '@/lib'
import { ArchitectureDiagram } from './ArchitectureDiagram'

const ArchitectureDiagramContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const {
    data: componentsResult,
    isLoading: componentsLoading,
    isError: componentsError,
  } = useQuery({
    queryKey: ['install-components-diagram', org?.id, install?.id],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id!,
        installId: install.id!,
        limit: 100,
        offset: 0,
      }),
    enabled: !!org?.id && !!install?.id,
    refetchInterval: 20000,
  })

  const { data: stack } = useQuery({
    queryKey: ['install-stack-diagram', org?.id, install?.id],
    queryFn: () =>
      getInstallStack({ orgId: org.id!, installId: install.id! }),
    enabled: !!org?.id && !!install?.id,
  })

  const { data: appConfig } = useQuery({
    queryKey: [
      'app-config-diagram',
      org?.id,
      install?.app_id,
      install?.app_config_id,
    ],
    queryFn: () =>
      getAppConfig({
        orgId: org.id!,
        appId: install.app_id!,
        appConfigId: install.app_config_id!,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!install?.app_config_id,
  })

  const { data: permissionsConfig } = useQuery({
    queryKey: ['install-permissions-config-diagram', org?.id, install?.id],
    queryFn: () =>
      getInstallAppPermissionsConfig({
        orgId: org.id!,
        installId: install.id!,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <ArchitectureDiagram
      install={install}
      components={componentsResult?.data ?? []}
      stack={stack ?? undefined}
      appConfig={appConfig ?? undefined}
      permissionsConfig={permissionsConfig ?? undefined}
      orgId={org?.id ?? ''}
      isLoading={componentsLoading}
      isError={componentsError}
    />
  )
}

const ArchitectureDiagramModal = ({ ...props }: IModal) => (
  <Modal
    heading={
      <Text className="inline-flex gap-2 items-center" variant="h3" weight="strong">
        <Icon variant="TreeStructure" size="20" />
        Architecture
      </Text>
    }
    size="xl"
    showFooter={false}
    childrenClassName="!p-0 flex-1 min-h-0"
    className="h-[80vh]"
    {...props}
  >
    <div className="w-full h-full">
      <ArchitectureDiagramContainer />
    </div>
  </Modal>
)

export const ArchitectureDiagramButton = ({
  ...props
}: Omit<IButtonAsButton, 'onClick'>) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      variant="ghost"
      onClick={() => {
        const modal = <ArchitectureDiagramModal />
        addModal(modal)
      }}
      {...props}
    >
      Architecture
      <Icon variant="TreeStructure" />
    </Button>
  )
}

export { ArchitectureDiagramContainer }
