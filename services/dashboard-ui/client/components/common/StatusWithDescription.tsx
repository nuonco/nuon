import { Tooltip, type ITooltip } from './Tooltip'
import { Status, type IStatus } from './Status'
import { Text } from './Text'

export interface IStatusWithDescription {
  statusProps: IStatus
  tooltipProps?: Omit<ITooltip, 'children'>
  maxWidth?: string
}

export const StatusWithDescription = ({
  statusProps: { variant = 'badge', ...statusProps },
  tooltipProps: {
    position = 'bottom',
    tipContent,
    tipContentClassName,
    ...tooltipProps
  } = { tipContent: undefined },
  maxWidth = 'max-w-xs',
}: IStatusWithDescription) => {
  if (!tipContent) {
    return <Status variant={variant} {...statusProps} />
  }

  const content =
    typeof tipContent === 'string' ? (
      <Text variant="subtext">{tipContent}</Text>
    ) : (
      tipContent
    )

  return (
    <Tooltip
      position={position}
      tipContent={content}
      tipContentClassName={`${maxWidth} whitespace-normal ${tipContentClassName || ''}`}
      {...tooltipProps}
    >
      <Status variant={variant} {...statusProps} />
    </Tooltip>
  )
}
