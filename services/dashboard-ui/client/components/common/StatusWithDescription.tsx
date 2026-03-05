import { Tooltip, type ITooltip } from './Tooltip'
import { Status, type IStatus } from './Status'

export interface IStatusWithDescription {
  statusProps: IStatus
  tooltipProps?: Omit<ITooltip, 'children'>
}

export const StatusWithDescription = ({
  statusProps: { variant = 'badge', ...statusProps },
  tooltipProps: { position = 'bottom', ...tooltipProps },
}: IStatusWithDescription) => {
  return (
    <Tooltip position={position} {...tooltipProps}>
      <Status variant={variant} {...statusProps} />
    </Tooltip>
  )
}
