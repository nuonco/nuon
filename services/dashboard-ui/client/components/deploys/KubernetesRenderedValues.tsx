import { KeyValueList } from '@/components/common/KeyValueList'

type TRenderedValue = Record<string, string>

export const KubernetesRenderedValues = ({
  values,
}: {
  values: TRenderedValue[]
}) => {
  return (
    <KeyValueList
      values={values?.map((v) => ({ key: v?.name, value: v?.value }))}
      emptyStateProps={{
        emptyTitle: 'No Kubernetes values',
        emptyMessage: 'No rendered Kubernetes values for this deployment',
      }}
    />
  )
}
