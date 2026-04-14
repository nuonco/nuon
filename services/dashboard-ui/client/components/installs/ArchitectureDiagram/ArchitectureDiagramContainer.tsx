import { useQuery } from '@tanstack/react-query'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallComponents, getInstallStack, getAppConfig, getInstallAppPermissionsConfig } from '@/lib'
import type { TCloudPlatform } from '@/types'
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

const InstallDetailsPanel = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  if (!install) return null

  const platform = install.gcp_account
    ? 'gcp'
    : install.aws_account
      ? 'aws'
      : install.azure_account
        ? 'azure'
        : undefined

  const region =
    install.gcp_account?.region || install.aws_account?.region
  const location = install.azure_account?.location

  const isManagedByConfig =
    install.metadata?.managed_by === 'nuon/cli/install-config'

  return (
    <div className="flex flex-col gap-5 p-5 overflow-y-auto">
      <LabeledValue label="Install ID">
        <ID>{install.id}</ID>
      </LabeledValue>

      {install.install_number != null && (
        <LabeledValue label="Install number">
          <Text variant="subtext">#{install.install_number}</Text>
        </LabeledValue>
      )}

      {install.created_by?.email && (
        <LabeledValue label="Created by">
          <Text variant="subtext">{install.created_by.email}</Text>
        </LabeledValue>
      )}

      <LabeledValue label="App">
        <Text variant="subtext">
          <Link href={`/${org?.id}/apps/${install.app_id}`}>
            {install.app?.name || install.app_id}
          </Link>
        </Text>
      </LabeledValue>

      {isManagedByConfig && (
        <LabeledValue label="Managed by">
          <Text variant="subtext">
            <span className="flex items-center gap-1">
              <Icon variant="FileCodeIcon" size={14} /> Install config
            </span>
          </Text>
        </LabeledValue>
      )}

      <LabeledValue label="Created">
        <Time variant="subtext" time={install.created_at} format="long-datetime" />
      </LabeledValue>

      {install.cloud_platform && (
        <LabeledValue label="Platform">
          <CloudPlatform
            platform={(install.cloud_platform as TCloudPlatform) || 'unknown'}
            variant="subtext"
            colorVariant="color"
          />
        </LabeledValue>
      )}

      {(region || location) && platform && (
        <LabeledValue label="Region">
          <CloudRegion
            variant="subtext"
            platform={platform}
            region={region}
            location={location}
          />
        </LabeledValue>
      )}
    </div>
  )
}

const InstallDetailsModal = ({ ...props }: IModal) => (
  <Modal
    heading={
      <Text className="inline-flex gap-2 items-center" variant="h3" weight="strong">
        <Icon variant="Info" size="20" />
        Install details
      </Text>
    }
    size="xl"
    showFooter={false}
    childrenClassName="!p-0 flex-1 min-h-0"
    className="h-[80vh]"
    {...props}
  >
    <div className="flex w-full h-full">
      <div className="w-[30%] border-r border-cool-grey-300 dark:border-dark-grey-300 shrink-0">
        <InstallDetailsPanel />
      </div>
      <div className="w-[70%] min-w-0">
        <ArchitectureDiagramContainer />
      </div>
    </div>
  </Modal>
)

export const InstallDetailsButton = ({
  ...props
}: Omit<IButtonAsButton, 'onClick'>) => {
  const { addModal } = useSurfaces()

  return (
    <Button
      variant="ghost"
      onClick={() => {
        const modal = <InstallDetailsModal />
        addModal(modal)
      }}
      {...props}
    >
      <Icon variant="Info" />
      Install details
    </Button>
  )
}

export { ArchitectureDiagramContainer }
