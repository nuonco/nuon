import { ContextTooltip, type TContextTooltipItem } from '@/components/common/ContextTooltip'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TTheme } from '@/types'

interface IVCSConnectionsStatusIndicator {
  items: TContextTooltipItem[]
  theme: TTheme
}

export const VCSConnectionsStatusIndicator = ({
  items,
  theme,
}: IVCSConnectionsStatusIndicator) => (
  <ContextTooltip position="top" title="VCS connections" items={items} width="w-56">
    <Text
      theme={theme}
      family="mono"
      variant="subtext"
      className="!flex gap-1.5 items-center cursor-default"
    >
      <Icon variant="GitBranchIcon" size={14} />
    </Text>
  </ContextTooltip>
)
