import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import type { TVCSConnection } from '@/types'
import { ConnectionDetailsButton } from './ConnectionDetails'
import { RemoveConnectionButton } from './RemoveConnection'

export const VCSManagementDropdown = ({
  vcs_connection,
}: {
  vcs_connection: TVCSConnection
}) => {
  return (
    <Dropdown
      buttonClassName="!p-1"
      buttonText=""
      id={vcs_connection?.id}
      icon={<Icon variant="DotsThreeVerticalIcon" />}
      variant="ghost"
      alignment="right"
    >
      <Menu>
        <ConnectionDetailsButton vcs_connection={vcs_connection} isMenuButton />
        <hr />
        <RemoveConnectionButton vcs_connection={vcs_connection} isMenuButton />
      </Menu>
    </Dropdown>
  )
}
