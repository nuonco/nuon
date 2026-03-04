import { KeyValueList } from '@/components/common/KeyValueList'
import { objectToKeyValueArray } from '@/utils/data-utils'

type TRenderedValue = Record<string, string>

export const TerraformRenderedVariables = ({
  values,
}: {
  values: TRenderedValue
}) => {
  return (
    <KeyValueList
      values={objectToKeyValueArray(values)}
      emptyStateProps={{
        emptyTitle: 'No Terraform variables',
        emptyMessage: 'No rendered Terraform variables for this deploy.',
      }}
    />
  )
}
