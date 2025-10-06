import { Button } from './Button'
import { Dropdown } from './Dropdown'
import { Icon } from './Icon'
import { Link } from './Link'
import { Menu } from './Menu'
import { Text } from './Text'

export const Default = () => (
  <Dropdown id="default-dropdown" buttonText="Default">
    <Menu className="min-w-48">
      <Text>Label</Text>
      <Button>Button</Button>
      <Link href="#">Link</Link>
      <hr />
      <Text>Label</Text>
      <Button>Button</Button>
    </Menu>
  </Dropdown>
)

export const Positions = () => (
  <div className="flex gap-4">
    <Dropdown id="below-dropdown" buttonText="Below" position="below">
      <Menu className="min-w-48">
        <Text>Label</Text>
        <Button>Button</Button>
        <Link href="#">Link</Link>
        <hr />
        <Text>Label</Text>
        <Button>Button</Button>
      </Menu>
    </Dropdown>
    <Dropdown id="above-dropdown" buttonText="Above" position="above">
      <Menu className="min-w-48">
        <Text>Label</Text>
        <Button>Button</Button>
        <Link href="#">Link</Link>
        <hr />
        <Text>Label</Text>
        <Button>Button</Button>
      </Menu>
    </Dropdown>
    <Dropdown id="beside-dropdown" buttonText="Beside" position="beside">
      <Menu className="min-w-48">
        <Text>Label</Text>
        <Button>Button</Button>
        <Link href="#">Link</Link>
        <hr />
        <Text>Label</Text>
        <Button>Button</Button>
      </Menu>
    </Dropdown>
  </div>
)

export const Alignments = () => (
  <div className="flex gap-4">
    <Dropdown id="left-dropdown" buttonText="Left" alignment="left">
      <Menu className="min-w-48">
        <Text>Label</Text>
        <Button>Button</Button>
        <Link href="#">Link</Link>
        <hr />
        <Text>Label</Text>
        <Button>Button</Button>
      </Menu>
    </Dropdown>
    <Dropdown id="right-dropdown" buttonText="Right" alignment="right">
      <Menu className="min-w-48">
        <Text>Label</Text>
        <Button>Button</Button>
        <Link href="#">Link</Link>
        <hr />
        <Text>Label</Text>
        <Button>Button</Button>
      </Menu>
    </Dropdown>
  </div>
)

export const NestedDropdowns = () => (
  <Dropdown id="nested-dropdown" buttonText="Default">
    <Menu className="min-w-48">
      <Text>Label</Text>
      <Button>Button</Button>
      <Link href="#">Link</Link>
      <hr />
      <Text>Label</Text>
      <Dropdown
        id="nested-dropdown-1"
        buttonText="Nested dropdown"
        position="beside"
        alignment="right"
        icon={<Icon variant="CaretRight" />}
      >
        <Menu className="min-w-48">
          <Text>Label</Text>
          <Button>Button</Button>
          <Link href="#">Link</Link>
          <hr />
          <Text>Label</Text>
          <Dropdown
            id="nested-dropdown-2"
            buttonText="Nested dropdown"
            position="beside"
            alignment="right"
            icon={<Icon variant="CaretRight" />}
          >
            <Menu className="min-w-48">
              <Text>Label</Text>
              <Button>Button</Button>
              <Link href="#">Link</Link>
              <hr />
              <Text>Label</Text>
              <Button>Button</Button>
            </Menu>
          </Dropdown>
        </Menu>
      </Dropdown>
    </Menu>
  </Dropdown>
)
