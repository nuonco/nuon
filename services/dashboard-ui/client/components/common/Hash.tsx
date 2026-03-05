import { ClickToCopyButton } from './ClickToCopy'
import { Text, type IText } from './Text'

export interface IHash extends Omit<IText, 'children'> {
  hash: string
  length?: number
}

export const Hash = ({
  hash,
  length = 12,
  variant = 'subtext',
  family = 'mono',
  ...props
}: IHash) => {
  if (!hash) {
    return (
      <Text variant={variant} theme="neutral" {...props}>
        —
      </Text>
    )
  }

  return (
    <Text
      className="!flex gap-1 items-center"
      variant={variant}
      family={family}
      {...props}
    >
      {hash.slice(0, length)}...
      <ClickToCopyButton textToCopy={hash} className="border-none" />
    </Text>
  )
}
