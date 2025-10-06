import { Menu } from './Menu'
import { Button } from './Button'
import { Icon } from './Icon'
import { Link } from './Link'
import { Text } from './Text'

export const Default = () => (
  <Menu className="w-56">
    <Button>Button</Button>
    <Link href="#">Link</Link>
    <hr />
    <Text>Text</Text>
  </Menu>
)

export const ComplexMenu = () => (
  <Menu className="w-56">
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
    <Button>
      Reprovision install <Icon variant="ArrowURightUp" />
    </Button>
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
    <Button className="!text-red-800 dark:!text-red-500">
      Forget install
      <Icon variant="Trash" />
    </Button>
  </Menu>
)
