import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { InstallProvider } from '@/providers/install-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TInstall } from '@/types'
import { EditInputsButton } from '../EditInputs'
import { ViewCurrentInputsButton } from '../ViewCurrentInputs'
import { EnableAutoApproveButton } from '../EnableAutoApprove'
import { ReprovisionButton } from '../Reprovision'
import { ForgetButton } from '../Forget'
import { SyncSecretsButton } from '../SyncSecrets'

interface IQuickManagementMenu {
  orgId: string
  installId: string
}

const QuickManagementMenu = ({ orgId, installId }: IQuickManagementMenu) => {
  return (
    <Menu>
      <Button href={`/${orgId}/installs/${installId}`}>
        View details
        <Icon variant="CaretRightIcon" />
      </Button>
      <hr />
      <Text variant="label" theme="neutral">
        Settings
      </Text>
      <EditInputsButton isMenuButton />
      <ViewCurrentInputsButton isMenuButton />
      <Button href={`/${orgId}/installs/${installId}/state`} isMenuButton>
        View state
        <Icon variant="CodeBlockIcon" />
      </Button>
      <EnableAutoApproveButton isMenuButton />
      <hr />
      <Text variant="label" theme="neutral">
        Controls
      </Text>
      <ReprovisionButton isMenuButton />
      <SyncSecretsButton isMenuButton />
      <hr />
      <Text variant="label" theme="neutral">
        Danger
      </Text>
      <span>
        <ForgetButton isMenuButton />
      </span>
    </Menu>
  )
}

export const QuickManagementDropdown = ({ install }: { install: TInstall }) => {
  return (
    <InstallProvider installId={install?.id} shouldPoll={false} loadingElement={<Skeleton height="24px" width="24px" />} errorElement={null}>
      <SurfacesProvider>
        <Dropdown
          alignment="right"
          buttonText=""
          buttonClassName="!p-1"
          icon={<Icon variant="DotsThreeVerticalIcon" />}
          id={install.id}
          variant="ghost"
        >
          <QuickManagementMenu orgId={install.org_id} installId={install.id} />
        </Dropdown>
      </SurfacesProvider>
    </InstallProvider>
  )
}
