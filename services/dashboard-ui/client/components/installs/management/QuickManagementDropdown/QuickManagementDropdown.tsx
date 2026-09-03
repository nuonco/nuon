import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { InstallProvider } from '@/providers/install-provider'
import { InstallAppConfigProvider } from '@/providers/install-app-config-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TInstall } from '@/types'
import { EditLabelsButton } from '../EditLabels'
import { EditInputsButton } from '../EditInputs'
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
      <Button href={`/${orgId}/installs/${installId}`}>View details</Button>
      <hr />
      <Text variant="label" theme="neutral">
        Settings
      </Text>
      <EditLabelsButton isMenuButton />
      <EditInputsButton isMenuButton />
      <Button href={`/${orgId}/installs/${installId}/inputs`} isMenuButton>
        Current inputs
        <Icon variant="ListChecksIcon" />
      </Button>
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
  const [hasOpened, setHasOpened] = useState(false)

  return (
    <InstallProvider
      installId={install?.id}
      shouldPoll={false}
      loadingElement={<Skeleton height="24px" width="24px" />}
      errorElement={null}
    >
      <InstallAppConfigProvider enabled={hasOpened}>
        <SurfacesProvider>
          <Dropdown
            alignment="right"
            buttonText=""
            buttonClassName="!p-1"
            icon={<Icon variant="DotsThreeVerticalIcon" />}
            id={install.id}
            variant="ghost"
            onOpenChange={(isOpen) => {
              if (isOpen) setHasOpened(true)
            }}
          >
            <QuickManagementMenu
              orgId={install.org_id}
              installId={install.id}
            />
          </Dropdown>
        </SurfacesProvider>
      </InstallAppConfigProvider>
    </InstallProvider>
  )
}
