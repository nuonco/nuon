import { useSidebar } from '@/hooks/use-sidebar'
import { Icon } from '@/components/common/Icon'
import { SidebarLogo } from '@/components/common/Logo/Logo'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { Text } from '@/components/common/Text'
import { MainNav } from '@/components/navigation/MainNav'
import { OrgSwitcher } from '@/components/orgs/OrgSwitcher'
import { UserDropdown } from '@/components/users/UserDropdown'
import type { TNuonVersion } from '@/types'
import { cn } from '@/utils/classnames'
import { MainSidebarButton } from './MainSidebarButton'

export const MainSidebar = ({
  versions,
  hideOrgContent = false,
}: {
  versions: TNuonVersion
  hideOrgContent?: boolean
}) => {
  const { isSidebarOpen } = useSidebar()
  return (
    <aside
      className={cn(
        'fixed inset-y-0 left-0 z-50 w-[280px] flex flex-col border-r',
        'transition-transform duration-fast ease-cubic',
        !isSidebarOpen ? 'translate-x-0' : '-translate-x-full',
        'md:static md:z-auto md:flex-none md:w-[4.5rem] md:translate-x-0',
        'md:transition-[width] md:duration-fast md:ease-cubic',
        { 'md:w-[17.5rem]': isSidebarOpen },
        'bg-gradient'
      )}
    >
      <header className="flex items-center justify-between h-16 px-4">
        <SidebarLogo />
        <div className="md:hidden">
          <MainSidebarButton variant="mobile-close" />
        </div>
      </header>
      <div className="p-4 flex flex-col gap-4 flex-auto">
        {!hideOrgContent && (
          <>
            <div className="flex h-14">
              <OrgSwitcher />
            </div>

            <MainNav />
          </>
        )}

        <div className="flex flex-auto items-end lg:hidden">
          <UserDropdown
            alignment="left"
            className="!w-full"
            buttonClassName="!w-full !h-[39px]"
            collapsible
            isCollapsed={!isSidebarOpen}
            hideOrgSettings={hideOrgContent}
            icon={isSidebarOpen ? <Icon variant="CaretUpIcon" /> : null}
            position="above"
          />
        </div>
      </div>
      {
        <div
          className={cn('grid transition-all ease-cubic duration-fast', {
            'grid-rows-[1fr] opacity-100': isSidebarOpen,
            'grid-rows-[0fr] opacity-0': !isSidebarOpen,
          })}
        >
          <div className="flex flex-col items-end p-4 overflow-hidden">
            <Text variant="label" theme="neutral">
              API: <b>{versions?.api?.version}</b>
            </Text>
            <Text variant="label" theme="neutral">
              UI: <b>{versions?.ui?.version}</b>
            </Text>
          </div>
        </div>
      }
    </aside>
  )
}
