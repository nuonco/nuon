import { DeploymentsDetails } from './DeploymentsDetails'
import { HelmManifest } from './HelmManifest'
import { IngressesDetails } from './IngressesDetails'
import { Overview } from './Overview'
import { ResourcesDetails } from './ResourcesDetails'
import { ServicesDetails } from './ServicesDetails'

export const HelmOutputs = ({
  createdAt,
  outputs,
}: {
  createdAt: string
  outputs: Record<string, any>
}) => {
  return (
    <div className="flex flex-col gap-6">
      <Overview createdAt={createdAt} outputs={outputs} />
      <DeploymentsDetails deployments={outputs?.deployments ?? {}} />
      <ServicesDetails services={outputs?.services ?? {}} />
      <IngressesDetails ingresses={outputs?.ingresses ?? {}} />
      <ResourcesDetails resources={outputs?.resources ?? {}} />
      {outputs?.manifest ? <HelmManifest manifest={outputs.manifest} /> : null}
    </div>
  )
}
