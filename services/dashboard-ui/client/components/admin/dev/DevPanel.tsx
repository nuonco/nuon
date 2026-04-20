import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { DevControls } from './DevControls'

export const DevPanel = ({ size = '3/4', ...props }: IPanel) => {
  return (
    <Panel
      heading={
        <div className="flex items-center gap-3">
          <Icon
            variant="Terminal"
            size="24"
            className="text-teal-600 dark:text-teal-400"
          />
          <Text weight="strong" variant="h2">
            Dev tools
          </Text>
        </div>
      }
      size={size}
      {...props}
    >
      <DevControls />
    </Panel>
  )
}
