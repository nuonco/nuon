import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import {
  getInstallDeploymentSourceDiff,
  type TInstallDeploymentSourceDiffTarget,
} from '@/lib/ctl-api/installs/get-install-deployment-source-diff'
import { InstallDeploymentSourceDiff } from './InstallDeploymentSourceDiff'

const SourceDiffContent = ({
  target,
}: {
  target: TInstallDeploymentSourceDiffTarget
}) => {
  const [selectedPath, setSelectedPath] = useState<string>()
  const queryKey = [
    'install-deployment-source-diff',
    target.orgId,
    target.installId,
    target.workflowId,
    target.kind,
    target.name,
  ]
  const enabled = !!target.orgId && !!target.installId && !!target.workflowId
  const manifest = useQuery({
    queryKey,
    queryFn: () => getInstallDeploymentSourceDiff(target),
    enabled,
  })
  const file = useQuery({
    queryKey: [...queryKey, selectedPath],
    queryFn: () =>
      getInstallDeploymentSourceDiff({ ...target, path: selectedPath! }),
    enabled:
      enabled &&
      manifest.data?.status === 'available' &&
      !!selectedPath &&
      !!manifest.data.files?.some(({ path }) => path === selectedPath),
  })

  return (
    <InstallDeploymentSourceDiff
      comparison={manifest.data}
      isLoading={manifest.isPending}
      isError={manifest.isError}
      onRetry={() => manifest.refetch()}
      selectedPath={selectedPath}
      onSelectPath={setSelectedPath}
      fileComparison={file.data}
      isLoadingFile={file.isPending}
      isFileError={file.isError}
      onRetryFile={() => file.refetch()}
    />
  )
}

export const InstallDeploymentSourceDiffPanel = ({
  target,
  ...props
}: {
  target: TInstallDeploymentSourceDiffTarget
} & Omit<IPanel, 'heading' | 'children'>) => (
  <Panel
    heading="Source changes"
    size="3/4"
    aria-label="Source changes"
    {...props}
  >
    <SourceDiffContent key={JSON.stringify(target)} target={target} />
  </Panel>
)
