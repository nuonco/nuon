import { useInstallOverview } from '@/hooks/use-install-overview'
import { useDriftCardProps } from '../DriftedResources'
import { DriftResourceCard } from '../DriftResourceCard'
import { ResourceAlignment } from './ResourceAlignment'

export const ResourceAlignmentContainer = ({
  kind,
}: {
  kind: 'sandbox' | 'stack'
}) => {
  const { data, error, isLoading, refetch } = useInstallOverview()
  const cardProps = useDriftCardProps()
  const resource = data?.alignment?.resources?.find(
    (item) => item.kind === kind
  )

  return (
    <ResourceAlignment
      label={kind}
      resource={resource}
      loading={isLoading}
      error={error}
      onRetry={() => void refetch()}
      card={resource ? <DriftResourceCard {...cardProps(resource)} /> : null}
    />
  )
}
