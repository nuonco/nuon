import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { ConnectGithubButton } from '@/components/vcs-connections/ConnectGithub'
import { VCSConnections } from '@/components/vcs-connections/VCSConnections'
import { useSidebar } from '@/hooks/use-sidebar'
import { useOrg } from '@/hooks/use-org'
import { cn } from '@/utils/classnames'
import './OrgAvatar.css'
import { OrgSummary } from './OrgSummary'
import { OrgsNavContainer as OrgsNav } from './OrgsNavContainer'

interface IOrgSwitcher
  extends Omit<IDropdown, 'buttonText' | 'children' | 'id'> {}

export const OrgSwitcher = ({}: IOrgSwitcher) => {
  const { isSidebarOpen } = useSidebar()
  const { org } = useOrg()
  if (!org) return null
  return (
    <Dropdown
      alignment="overlay"
      className="w-full md:w-[248px] duration-fastest transition-all"
      buttonClassName={cn(
        'w-full text-left duration-fastest transition-all !text-foreground !border-[var(--border-color)]',
        {
          '!px-4 !py-1.5 ': isSidebarOpen,
          '!p-[3px] !size-10 ': !isSidebarOpen,
        }
      )}
      buttonText={
        <OrgSummary isButtonSummary isSidebarOpen={isSidebarOpen} org={org} />
      }
      icon={isSidebarOpen ? <Icon variant="CaretUpDown" /> : null}
      closeOnBlur={false}
      id="org-switcher"
      position="overlay"
      variant="ghost"
    >
      <Menu
        className="w-[248px] h-fit max-h-[500px] overflow-y-scroll overflow-x-hidden focus:outline-primay-400 !p-0"
        tabIndex={-1}
        style={{ scrollbarGutter: 'stable' }}
      >
        <div className="p-3 border-b">
          <OrgSummary org={org} />
          <ID className="!flex mt-2">{org.id}</ID>
        </div>
        <div className="px-3 py-4 flex flex-col gap-4">
          <div className="flex justify-between items-center">
            <Text variant="subtext" weight="strong">
              GitHub connections
            </Text>
            <ConnectGithubButton />
          </div>
          <div className="flex flex-col gap-2">
            <VCSConnections vcsConnections={org?.vcs_connections} />
          </div>
        </div>
        <hr className="border-dashed mx-4" />
        <div className="px-1 py-4 flex flex-col gap-1.5">
          <div className="px-2">
            <Text variant="subtext" weight="strong">
              Organizations
            </Text>
          </div>
          <OrgsNav />
        </div>
      </Menu>
    </Dropdown>
  )
}
