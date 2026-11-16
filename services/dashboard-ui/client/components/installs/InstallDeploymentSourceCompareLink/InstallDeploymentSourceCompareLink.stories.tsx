import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import type { TInstallDeploymentSourceDiff } from '@/types'
import { InstallDeploymentSourceCompareLink } from './InstallDeploymentSourceCompareLink'

export default { title: 'Features / Installs / Source compare link' }

const target = {
  orgId: 'org-acme',
  installId: 'ins-acme',
  workflowId: 'wf-deploy',
  kind: 'component' as const,
  name: 'acme-api',
}

const CompareLinkStory = ({
  comparison,
}: {
  comparison: TInstallDeploymentSourceDiff
}) => {
  const client = useQueryClient()
  useEffect(() => {
    client.setQueryData(
      [
        'install-deployment-source-diff',
        target.orgId,
        target.installId,
        target.workflowId,
        target.kind,
        target.name,
      ],
      comparison
    )
  }, [client, comparison])
  return <InstallDeploymentSourceCompareLink target={target} />
}

export const Available = () => (
  <CompareLinkStory
    comparison={{
      compare_url:
        'https://github.com/acme/example/compare/1111111111111111111111111111111111111111..2222222222222222222222222222222222222222',
    }}
  />
)

export const NoCompare = () => <CompareLinkStory comparison={{}} />
NoCompare.storyName = 'No compare link'
