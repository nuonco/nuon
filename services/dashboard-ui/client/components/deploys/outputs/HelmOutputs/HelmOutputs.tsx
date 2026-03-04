import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
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
      <DeploymentsDetails deployments={outputs?.deployments} />
      <ServicesDetails services={outputs?.services} />
      <IngressesDetails ingresses={outputs?.ingresses} />
      <ResourcesDetails resources={outputs?.resources} />
      <HelmManifest manifest={outputs?.manifest} />
    </div>
  )
}

const OUTPUT_SKELETON_SECTIONS = [
  'Deployments detials',
  'Services details',
  'Ingresses details',
  'Resources',
]

export const HelmOutputsSkeleton = () => {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <span className="flex items-center gap-4">
          <Skeleton height="20px" width="100px" />
          <Skeleton height="23px" width="72px" />
        </span>
        <span className="flex items-center gap-6">
          <Skeleton height="17px" width="85px" />
          <Skeleton height="17px" width="110px" />
          <Skeleton height="17px" width="175px" />
        </span>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
        <Skeleton height="110px" width="100%" />
        <Skeleton height="110px" width="100%" />
        <Skeleton height="110px" width="100%" />
        <Skeleton height="110px" width="100%" />
      </div>

      {OUTPUT_SKELETON_SECTIONS.map((sec) => (
        <div key={sec} className="flex flex-col gap-2">
          <Text weight="strong">{sec}</Text>
          <Skeleton width="100%" height="300px" />
        </div>
      ))}

      <div className="flex flex-col gap-2">
        <Text weight="strong">Helm manifest</Text>
        <Skeleton width="100%" height="800px" />
      </div>
    </div>
  )
}
