import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'

import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallComponents, getInstallStack, getAppConfig, getInstallAppPermissionsConfig, getInstallAuditLog } from '@/lib'
import type { TCloudPlatform } from '@/types'
import { downloadFileOnClick } from '@/utils/file-download'
import { slugify } from '@/utils/string-utils'
import { cn } from '@/utils/classnames'
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

type TabId = 'architecture' | 'details' | 'auditLogs'

const TAB_CONFIG: { id: TabId; label: string; icon: TIconVariant }[] = [
  { id: 'architecture', label: 'Architecture', icon: 'TreeStructure' },
  { id: 'details', label: 'Install details', icon: 'Info' },
  { id: 'auditLogs', label: 'Audit logs', icon: 'ClockClockwise' },
]

const InstallDetailsTab = () => {
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
    <div className="flex flex-col gap-5 p-5 overflow-y-auto h-full">
      <div className="flex flex-col gap-1">
        <Text variant="body" weight="strong">Install details</Text>
        <Text variant="subtext" theme="neutral">
          Metadata and configuration for this install.
        </Text>
      </div>

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

      <LabeledValue label="Updated">
        <Time variant="subtext" time={install.updated_at} format="long-datetime" />
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

const AuditLogsTab = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const [dateRange, setDateRange] = useState({
    start: new Date(new Date().getTime() - 7 * 24 * 60 * 60 * 1000),
    end: new Date(),
  })

  const {
    data: auditLog,
    error,
    isLoading,
  } = useQuery({
    queryKey: ['install-audit-log', org.id, install.id, dateRange.start.toISOString(), dateRange.end.toISOString()],
    queryFn: () =>
      getInstallAuditLog({
        orgId: org.id,
        installId: install.id,
        start: dateRange.start.toISOString(),
        end: dateRange.end.toISOString(),
      }),
  })

  const handleDateChange = (hoursAgo: number) => {
    const end = new Date()
    const start = new Date(end.getTime() - hoursAgo * 60 * 60 * 1000)
    setDateRange({ start, end })
  }

  const handleDownload = () => {
    if (auditLog?.content) {
      downloadFileOnClick({
        ...auditLog,
        filename: `${slugify(install.name)}-audit-log.csv`,
        fileType: 'csv',
        mimeType: 'text/csv',
      })
    }
  }

  return (
    <div className="flex flex-col gap-5 p-5 overflow-y-auto h-full">
      <div className="flex flex-col gap-1">
        <Text variant="body" weight="strong">Audit logs</Text>
        <Text variant="subtext" theme="neutral">
          See a complete record of all activities performed in this install.
        </Text>
      </div>

      {error ? (
        <Banner theme="error">
          {error?.error || 'Unable to load audit logs for the selected date range'}
        </Banner>
      ) : null}

      <div className="flex flex-col gap-3">
        <RadioInput
          name="date-range"
          value="1"
          onChange={() => handleDateChange(1)}
          labelProps={{ labelText: 'Last one hour' }}
        />
        <RadioInput
          name="date-range"
          value="24"
          onChange={() => handleDateChange(24)}
          labelProps={{ labelText: 'Last 24 hours' }}
        />
        <RadioInput
          name="date-range"
          value="168"
          onChange={() => handleDateChange(7 * 24)}
          defaultChecked={true}
          labelProps={{ labelText: 'Last week' }}
        />
        <RadioInput
          name="date-range"
          value="720"
          onChange={() => handleDateChange(30 * 24)}
          labelProps={{ labelText: 'Last 30 days' }}
        />
        <RadioInput
          name="date-range"
          value="1440"
          onChange={() => handleDateChange(60 * 24)}
          labelProps={{ labelText: 'Last 60 days' }}
        />
      </div>

      <Button
        variant="primary"
        disabled={isLoading || !auditLog?.content}
        onClick={handleDownload}
        className="w-fit"
      >
        {isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Download CSV
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="DownloadSimple" size="18" /> Download CSV
          </span>
        )}
      </Button>
    </div>
  )
}

const InstallDetailsModal = ({ ...props }: IModal) => {
  const [activeTab, setActiveTab] = useState<TabId>('architecture')

  return (
    <Modal
      heading={
        <Text className="inline-flex gap-2 items-center" variant="h3" weight="strong">
          <Icon variant="Info" size="20" />
          Install details
        </Text>
      }
      size="full"
      showFooter={false}
      childrenClassName="!p-0 flex-1 min-h-0"
      className="h-[80vh]"
      {...props}
    >
      <div className="flex w-full h-full">
        <nav className="w-[200px] shrink-0 border-r border-cool-grey-300 dark:border-dark-grey-300 flex flex-col gap-1 p-2">
          {TAB_CONFIG.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex items-center gap-2 px-3 py-2.5 rounded-md text-left text-[14px] leading-[21px] tracking-[-0.2px] transition-colors w-full cursor-pointer',
                activeTab === tab.id
                  ? 'text-primary-800 dark:text-primary-400 bg-primary-200 dark:bg-primary-600/25'
                  : 'text-cool-grey-800 dark:text-cool-grey-400 hover:bg-black/5 dark:hover:bg-white/10'
              )}
            >
              <Icon variant={tab.icon} size={16} />
              {tab.label}
            </button>
          ))}
        </nav>
        <div className="flex-1 min-w-0 min-h-0">
          {activeTab === 'architecture' && <ArchitectureDiagramContainer />}
          {activeTab === 'details' && <InstallDetailsTab />}
          {activeTab === 'auditLogs' && <AuditLogsTab />}
        </div>
      </div>
    </Modal>
  )
}

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
