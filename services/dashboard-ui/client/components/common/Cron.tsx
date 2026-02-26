import cronstrue from 'cronstrue'
import { Text, type IText } from './Text'
import { Tooltip } from './Tooltip'

export interface ICron extends Omit<IText, 'role'> {
  cron?: string
  format?: 'human' | 'expression' | 'both'
  showTooltip?: boolean
}

export const Cron = ({
  cron,
  format = 'human',
  showTooltip = true,
  ...props
}: ICron) => {
  if (!cron) {
    return (
      <Text {...props} theme="neutral">
        —
      </Text>
    )
  }

  let humanReadable: string
  try {
    humanReadable = cronstrue.toString(cron, {
      throwExceptionOnParseError: true,
      verbose: false,
    })
  } catch {
    humanReadable = cron
  }

  const getContent = () => {
    switch (format) {
      case 'expression':
        return (
          <Text {...props} family="mono">
            {cron}
          </Text>
        )
      case 'both':
        return (
          <div className="flex flex-col gap-1">
            <Text {...props}>{humanReadable}</Text>
            <Text {...props} family="mono" variant="label" theme="neutral">
              {cron}
            </Text>
          </div>
        )
      case 'human':
      default:
        return <Text {...props}>{humanReadable}</Text>
    }
  }

  if (showTooltip && format === 'human') {
    return (
      <Tooltip tipContent={cron}>
        <Text {...props}>{humanReadable}</Text>
      </Tooltip>
    )
  }

  return getContent()
}
