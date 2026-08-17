import { LabeledValue, type ILabeledValue } from './LabeledValue'
import {
  StatusWithDescription,
  type IStatusWithDescription,
} from './StatusWithDescription'

interface ILabeledStatus
  extends Omit<ILabeledValue, 'children'>,
    Omit<IStatusWithDescription, 'statusProps'> {
  statusProps?: IStatusWithDescription['statusProps']
}

export const LabeledStatus = ({
  statusProps,
  tooltipProps,
  loading,
  loadingWidth,
  ...props
}: ILabeledStatus) => {
  return (
    <LabeledValue loading={loading} loadingWidth={loadingWidth} {...props}>
      {statusProps ? (
        <StatusWithDescription
          statusProps={statusProps}
          tooltipProps={tooltipProps}
        />
      ) : null}
    </LabeledValue>
  )
}
