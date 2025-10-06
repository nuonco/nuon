import { Text, type IText } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { getSeverityBorderClasses } from '@/utils/log-stream-utils'

interface ILogSeverity extends IText {
  severityNumber: number
  severityText: string
}

export const LogSeverity = ({
  className,
  family = 'mono',
  severityNumber,
  severityText,
  ...props
}: ILogSeverity) => {
  return (
    <Text
      className={cn('flex items-center gap-1', className)}
      family={family}
      {...props}
    >
      <span
        className={cn(
          'flex border-l-2 h-5',
          getSeverityBorderClasses(severityNumber)
        )}
      />
      {severityText.toUpperCase()}
    </Text>
  )
}
