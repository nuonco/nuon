import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { EnvAccentSettings } from '@/components/orgs/EnvAccentSettings'

interface IAdminEnvAccentsPanel extends IPanel {
  orgId: string
}

export const AdminEnvAccentsPanel = ({
  orgId,
  size = 'half',
  ...props
}: IAdminEnvAccentsPanel) => {
  return (
    <Panel
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="PaletteIcon" size="24" />
          <Text weight="strong" variant="h2">
            Environment accents
          </Text>
        </div>
      }
      size={size}
      {...props}
    >
      <div className="flex flex-col gap-6">
        <Text variant="body" className="text-gray-600 dark:text-gray-300">
          Map install labels to accent colors so prod/non-prod installs are
          easy to tell apart at a glance. Applies to organization:{' '}
          <span className="font-mono">{orgId}</span>
        </Text>
        <EnvAccentSettings />
      </div>
    </Panel>
  )
}
