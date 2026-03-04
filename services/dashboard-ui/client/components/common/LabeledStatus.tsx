import { LabeledValue, type ILabeledValue } from './LabeledValue'
import {
  StatusWithDescription,
  type IStatusWithDescription,
} from './StatusWithDescription'

interface ILabeledStatus
  extends Omit<ILabeledValue, 'children'>,
    IStatusWithDescription {}

export const LabeledStatus = ({
  statusProps,
  tooltipProps,
  ...props
}: ILabeledStatus) => {
  return (
    <LabeledValue {...props}>
      <StatusWithDescription
        statusProps={statusProps}
        tooltipProps={tooltipProps}
      />
    </LabeledValue>
  )
}
