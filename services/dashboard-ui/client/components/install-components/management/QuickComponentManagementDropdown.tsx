import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { DeployComponentButton } from '@/components/install-components/management/DeployComponent'
import { DriftScanComponentButton } from '@/components/install-components/management/DriftScanComponent'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TInstallComponent } from '@/types'

export const QuickComponentManagementDropdown = ({
  installComponent,
  orgId,
  installId,
}: {
  installComponent: TInstallComponent
  orgId: string
  installId: string
}) => {
  const component = installComponent.component
  if (!component) return null

  const href = `/${orgId}/installs/${installId}/components/${component.id}`

  return (
    <SurfacesProvider>
      <Dropdown
        alignment="right"
        buttonText=""
        buttonClassName="!p-1"
        icon={<Icon variant="DotsThreeVerticalIcon" />}
        id={`component-quick-${component.id}`}
        variant="ghost"
      >
        <Menu>
          <Button href={href}>
            View component
            <Icon variant="CaretRightIcon" />
          </Button>
          <hr />
          <Text variant="label" theme="neutral">
            Controls
          </Text>
          <DriftScanComponentButton
            component={component}
            isMenuButton
          />
          <DeployComponentButton
            component={component}
            isMenuButton
          />
        </Menu>
      </Dropdown>
    </SurfacesProvider>
  )
}
