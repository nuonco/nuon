import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useAuth } from '@/hooks/use-auth'
import { AdminControls } from './AdminControls'

export const AdminPanel = ({ size = '3/4', ...props }: IPanel) => {
  const { demoMode, toggleDemoMode } = useAuth()

  return (
    <Panel
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="Gear" size="24" />
          <Text weight="strong" variant="h2">
            Admin Controls
          </Text>
        </div>
      }
      size={size}
      {...props}
    >
      <CheckboxInput
        checked={demoMode}
        onChange={toggleDemoMode}
        labelProps={{ className: 'w-fit', labelText: 'Demo mode' }}
      />
      <AdminControls />
    </Panel>
  )
}
