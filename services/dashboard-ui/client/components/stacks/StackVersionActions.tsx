import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { useSurfaces } from '@/hooks/use-surfaces'
import { StackVersionDetails } from './StackVersionDetails'
import { SendStackOutputsModal } from './SendStackOutputsModal'
import type { TInstallStack } from '@/types'

type TStackVersion = TInstallStack['versions'][number]

const StackVersionActionsMenu = ({ version }: { version: TStackVersion }) => {
  const { addModal, addPanel } = useSurfaces()

  const panel = <StackVersionDetails version={version} />
  const phoneHomeId = version?.phone_home_id

  return (
    <Menu>
      <Button onClick={() => addPanel(panel)}>
        View details
        <Icon variant="InfoIcon" />
      </Button>
      {phoneHomeId && (
        <Button
          onClick={() =>
            addModal(
              <SendStackOutputsModal
                phoneHomeId={phoneHomeId}
                versionId={version?.id}
              />
            )
          }
        >
          Trigger phone home
          <Icon variant="PhoneTransferIcon" />
        </Button>
      )}
    </Menu>
  )
}

export const StackVersionActions = ({
  version,
}: {
  version: TStackVersion
}) => {
  return (
    <Dropdown
      alignment="right"
      buttonText=""
      buttonClassName="!p-1"
      icon={<Icon variant="DotsThreeVerticalIcon" />}
      id={`stack-version-${version?.id}`}
      variant="ghost"
    >
      <StackVersionActionsMenu version={version} />
    </Dropdown>
  )
}
