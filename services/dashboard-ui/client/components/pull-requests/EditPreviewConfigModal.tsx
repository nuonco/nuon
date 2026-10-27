import { useQuery } from '@tanstack/react-query'
import { PullRequestConfigure } from '@/components/playground/pull-request/PullRequestConfigure'
import type {
  TPlaygroundInstallCandidate,
  TPlaygroundPullRequest,
} from '@/components/playground/pull-request/fixtures'
import type { IModal } from '@/components/surfaces/Modal'
import { getPreviewInstallCandidates } from '@/lib'
import type { TAppBranchPreviewConfig, TInstall } from '@/types'

const candidateFromInstall = (
  install: TInstall
): TPlaygroundInstallCandidate => {
  const platform = install.aws_account
    ? 'aws'
    : install.gcp_account
      ? 'gcp'
      : install.azure_account
        ? 'azure'
        : 'cloud'
  const region =
    install.aws_account?.region ||
    install.gcp_account?.region ||
    install.azure_account?.location ||
    ''
  return {
    id: install.id,
    name: install.name,
    labels: install.labels ?? {},
    cloud_platform: platform,
    region,
  }
}

export const EditPreviewConfigModal = ({
  pullRequest,
  orgId,
  appId,
  branchId,
  configId,
  onSave,
  ...props
}: {
  pullRequest: TPlaygroundPullRequest
  orgId: string
  appId: string
  branchId: string
  configId?: string
  onSave: (override: TAppBranchPreviewConfig) => void | Promise<unknown>
} & IModal) => {
  const installs = useQuery({
    queryKey: [
      'preview-install-candidates',
      orgId,
      appId,
      branchId,
      configId,
    ],
    queryFn: () =>
      getPreviewInstallCandidates({ orgId, appId, branchId, configId }),
    enabled: !!orgId && !!appId && !!branchId,
  })

  return (
    <PullRequestConfigure
      presentation="modal"
      pullRequest={pullRequest}
      installs={(installs.data?.installs ?? []).map(candidateFromInstall)}
      githubLocked
      modeDescription="Used for the preview run this starts. Later pushes still follow the branch preview config."
      onSave={onSave}
      {...props}
    />
  )
}
