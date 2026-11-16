import { useQuery } from '@tanstack/react-query'
import { Link } from '@/components/common/Link'
import {
  getInstallDeploymentSourceDiff,
  type TInstallDeploymentSourceDiffTarget,
} from '@/lib/ctl-api/installs/get-install-deployment-source-diff'
import type { TAPIError, TInstallDeploymentSourceDiff } from '@/types'

export const InstallDeploymentSourceCompareLink = ({
  target,
}: {
  target: TInstallDeploymentSourceDiffTarget
}) => {
  const { data } = useQuery<TInstallDeploymentSourceDiff, TAPIError>({
    queryKey: [
      'install-deployment-source-diff',
      target.orgId,
      target.installId,
      target.workflowId,
      target.kind,
      target.name,
    ],
    queryFn: () => getInstallDeploymentSourceDiff(target),
    enabled: !!target.orgId && !!target.installId && !!target.workflowId,
    staleTime: Infinity,
    retry: false,
    refetchOnWindowFocus: false,
  })

  if (!data?.compare_url) return null
  return (
    <Link
      href={data.compare_url}
      isExternal
      aria-label={`View changes on GitHub for ${target.name}`}
    >
      View changes on GitHub
    </Link>
  )
}
