import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { ReprovisionButton } from './Reprovision'

export const InstallManagementDropdown = () => {
  return (
    <Dropdown
      buttonText="Manage"
      id="install-mgmt"
      variant="primary"
      alignment="right"
    >
      <Menu className="min-w-56">
        <Text variant="label" theme="neutral">
          Settings
        </Text>
        <Button>
          Edit inputs <Icon variant="PencilSimpleLine" />
        </Button>
        <Button>
          Auto approve changes <Icon variant="ListChecks" />
        </Button>
        <Button>
          View state <Icon variant="CodeBlock" />
        </Button>
        <hr />
        <Text variant="label" theme="neutral">
          Controlls
        </Text>
        <ReprovisionButton isMenuButton />
        <Button>
          Deprovision install <Icon variant="ArrowURightDown" />
        </Button>
        <Button>
          Deprovision stack <Icon variant="StackMinus" />
        </Button>
        <hr />
        <Text variant="label" theme="neutral">
          Danger
        </Text>
        <Button>
          Break glass permissions <Icon variant="LockLaminated" />
        </Button>
        <span>
          <Button
            className="!text-red-800 dark:!text-red-500 !p-2 w-full justify-between"
            variant="ghost"
          >
            Forget install
            <Icon variant="Trash" />
          </Button>
        </span>
      </Menu>
    </Dropdown>
  )
}
