import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import type { TInstallComponent } from '@/types'
import { useInstallLink } from '@/hooks/use-install-path'

export const QuickComponentManagementDropdown = ({
  installComponent,
  orgId,
  installId,
}: {
  installComponent: TInstallComponent
  orgId: string
  installId: string
}) => {
  const installLink = useInstallLink()
  const component = installComponent.component
  if (!component) return null

  const href = installLink({
    orgId: orgId,
    installId: installId,
    suffix: `/components/${component.id}`,
  })

  return (
    <Dropdown
      alignment="right"
      buttonText=""
      buttonClassName="!p-1"
      icon={<Icon variant="DotsThreeVerticalIcon" />}
      id={`component-quick-${component.id}`}
      variant="ghost"
    >
      <Menu>
        <Button href={href}>View component</Button>
      </Menu>
    </Dropdown>
  )
}
