import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { SandboxRunMenu } from './SandboxRunMenu'

interface ISandboxRunSwitcher
  extends Omit<IDropdown, 'children' | 'id' | 'buttonText'> {
  sandboxRunId: string
}

export const SandboxRunSwitcher = ({
  alignment = 'right',
  sandboxRunId,
  ...props
}: ISandboxRunSwitcher) => {
  return (
    <Dropdown
      id="runs-switcher"
      alignment={alignment}
      buttonText="Latest sandbox runs"
      {...props}
    >
      <SandboxRunMenu activeSandboxRunId={sandboxRunId} />
    </Dropdown>
  )
}
