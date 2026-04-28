import { TerraformRenderedVariables } from '@/components/deploys/TerraformRenderedVariables'
import { EmptyState } from '@/components/common/EmptyState'
import { useDeploy } from '@/hooks/use-deploy'

export const DeployVariablesTab = () => {
  const { deploy } = useDeploy()
  const values = deploy?.outputs?.rendered_variables

  if (!values || Object.keys(values).length === 0) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No variables"
        emptyMessage="No rendered Terraform variables available for this deploy."
      />
    )
  }

  return <TerraformRenderedVariables values={values} />
}
