import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { DeployComponentButton } from '@/components/install-components/management/DeployComponent'
import { DriftScanComponentButton } from '@/components/install-components/management/DriftScanComponent'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TInstallComponent } from '@/types'

export const QuickComponentManagementDropdown = ({
  installComponent,
  orgId,
  installId,
  removed = false,
}: {
  installComponent: TInstallComponent
  orgId: string
  installId: string
  removed?: boolean
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
          {removed ? (
            <Tooltip
              className="block !w-full"
              position="left"
              tipContent={
                <Text variant="subtext">
                  This component is no longer in the install's app config
                  version.
                </Text>
              }
            >
              <Button isMenuButton disabled className="pointer-events-none w-full">
                Deploy component
                <Icon variant="CloudArrowUpIcon" />
              </Button>
            </Tooltip>
          ) : (
            <>
              <DriftScanComponentButton component={component} isMenuButton />
              <DeployComponentButton component={component} isMenuButton />
            </>
          )}
        </Menu>
      </Dropdown>
    </SurfacesProvider>
  )
}
