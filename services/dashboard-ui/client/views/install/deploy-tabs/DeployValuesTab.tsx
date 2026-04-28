import { RenderedValues } from '@/components/deploys/RenderedValues'
import { EmptyState } from '@/components/common/EmptyState'
import { useDeploy } from '@/hooks/use-deploy'

export const DeployValuesTab = () => {
  const { deploy } = useDeploy()
  const values = deploy?.outputs?.rendered_values

  if (!values || (Array.isArray(values) ? values.length === 0 : Object.keys(values).length === 0)) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No values"
        emptyMessage="No rendered Helm values available for this deploy."
      />
    )
  }

  return <RenderedValues values={values} />
}
