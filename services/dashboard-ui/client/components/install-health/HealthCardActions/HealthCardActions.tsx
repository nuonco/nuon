import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { RefreshClusterAccessModal } from '@/components/install-health/RefreshClusterAccess'
import { ResetHealthWindowModal } from '@/components/install-health/ResetHealthWindow'
import { useSurfaces } from '@/hooks/use-surfaces'

export const HealthCardActions = ({ installId }: { installId: string }) => {
  const { addModal } = useSurfaces()

  return (
    <Dropdown
      alignment="right"
      buttonText=""
      buttonClassName="!p-1"
      icon={<Icon variant="DotsThreeVerticalIcon" />}
      id={`install-health-actions-${installId}`}
      variant="ghost"
    >
      <Menu>
        <Text variant="label" theme="neutral">
          Controls
        </Text>
        <Button
          onClick={() =>
            addModal(<RefreshClusterAccessModal installId={installId} />)
          }
        >
          Refresh cluster access
          <Icon variant="ArrowsClockwiseIcon" />
        </Button>
        <Button
          onClick={() =>
            addModal(<ResetHealthWindowModal installId={installId} />)
          }
        >
          Reset window
          <Icon variant="ClockCounterClockwiseIcon" />
        </Button>
      </Menu>
    </Dropdown>
  )
}
