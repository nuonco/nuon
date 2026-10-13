import { useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
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
}

const unavailable = (
  <Text variant="subtext" theme="neutral">
    This deployment could not be loaded.
  </Text>
)

export const InstallDeploymentPanel = ({
  orgId,
  appId,
  installId,
  workflowId,
  title,
  repo,
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
      {...panelProps}
    >
      <InstallProvider
        installId={installId}
        loadingElement={<DeploymentDetailLoading />}
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
