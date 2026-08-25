import { useParams } from 'react-router'
import { useOutletContext } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { HelmOutputs } from '@/components/deploys/outputs/HelmOutputs/HelmOutputs'
import { EmptyState } from '@/components/common/EmptyState'
import { Loading } from '@/components/common/Loading'
import { TerraformOutputs } from '@/components/terraform-outputs/TerraformOutputs'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'
import { getInstallComponentOutputs } from '@/lib'
import type { TDeployOutletContext } from './types'

export const DeployOutputsTab = () => {
  const { componentId, installId } = useParams()
  const { component } = useOutletContext<TDeployOutletContext>()
  const { org } = useOrg()
  const { deploy } = useDeploy()
  const { install } = useInstall()

  const { data: outputs, isLoading, error } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-component-outputs', org?.id, installId, componentId],
    queryFn: () =>
      getInstallComponentOutputs({
        orgId: org.id,
        installId: installId!,
        componentId: componentId!,
      }),
    enabled: !!org?.id && !!installId && !!componentId,
    retry: false,
  })

  return (
    <>
      <PageTitle segments={['Deploy outputs', install?.name]} />
      {isLoading ? (
        <div className="flex justify-center py-10">
          <Loading variant="large" />
        </div>
      ) : error || !outputs ? (
        <EmptyState
          variant="history"
          emptyTitle="No outputs"
          emptyMessage="No outputs available for this component yet."
        />
      ) : component?.type === 'helm_chart' ? (
        <HelmOutputs createdAt={deploy?.created_at} outputs={outputs} />
      ) : (
        <TerraformOutputs
          heading={`${component?.name ?? 'Component'} outputs`}
          outputs={outputs}
        />
      )}
    </>
  )
}
