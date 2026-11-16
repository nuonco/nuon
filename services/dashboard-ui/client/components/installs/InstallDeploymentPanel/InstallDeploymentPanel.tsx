import { useQuery } from '@tanstack/react-query'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TDeploymentTab } from '@/components/installs/DeploymentDetail/deployment-progress'
import {
  DeploymentDetailBody,
  DeploymentDetailLoading,
} from '@/components/installs/DeploymentsList'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { getInstallDeployment } from '@/lib'
import { InstallProvider } from '@/providers/install-provider'

export interface IInstallDeploymentPanel extends IPanel {
  orgId: string
  appId: string
  installId: string
  workflowId: string
  title: string
  repo?: string
  initialTab?: TDeploymentTab
}

export const deploymentPanelKey = (installId: string, workflowId: string) =>
  `deployment-${installId}-${workflowId}`

export const parseDeploymentPanelKey = (key: string) => {
  const match = key.match(/^deployment-([a-z0-9]+)-([a-z0-9]+)$/)
  return match ? { installId: match[1], workflowId: match[2] } : null
}

const unavailable = (
  <Text variant="subtext" theme="neutral">
    This deployment could not be loaded.
  </Text>
)

export const PendingInstallPanel = ({
  title,
  status,
}: {
  title: string
  status?: string
}) => (
  <Panel heading={title} size="3/4" aria-label="Install details">
    <div className="flex flex-col gap-3">
      {status ? <Status status={status} /> : null}
      <Text variant="subtext" theme="neutral">
        This install has not started in this rollout.
      </Text>
    </div>
  </Panel>
)

export const InstallDeploymentPanel = ({
  orgId,
  appId,
  installId,
  workflowId,
  title,
  repo,
  initialTab,
  ...panelProps
}: IInstallDeploymentPanel) => {
  const { data: record, isError } = useQuery({
    queryKey: ['install-deployment', orgId, installId, workflowId],
    queryFn: () => getInstallDeployment({ orgId, installId, workflowId }),
    enabled: !!orgId && !!installId && !!workflowId,
  })

  return (
    <Panel
      heading={title}
      size="3/4"
      aria-label="Deployment details"
      panelKey={deploymentPanelKey(installId, workflowId)}
      {...panelProps}
    >
      <InstallProvider
        installId={installId}
        loadingElement={<DeploymentDetailLoading initialTab={initialTab} />}
        errorElement={unavailable}
      >
        {record ? (
          <DeploymentDetailBody
            deployment={{
              id: record.id,
              type: record.type,
              title: record.title || title,
              created_at: record.created_at,
              status: record.status,
              steps: [],
            }}
            orgId={orgId}
            appId={appId}
            installId={installId}
            repo={repo}
            initialTab={initialTab}
          />
        ) : isError ? (
          unavailable
        ) : (
          <DeploymentDetailLoading />
        )}
      </InstallProvider>
    </Panel>
  )
}
