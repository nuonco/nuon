import { cn } from '@/utils/classnames'
import { Block } from './Block'
import { NavBlock } from './NavBlock'
import { useShell } from './shell-context'
import type { INavItem } from './types'

export interface ISidebar {
  primaryNav: INavItem[]
  secondaryNav?: INavItem[]
}

export const Sidebar = ({ primaryNav, secondaryNav = [] }: ISidebar) => {
  const { isSidebarOpen } = useShell()

  const iconSize = isSidebarOpen ? 16 : 24
  const itemClass = cn('w-full h-[32px]', !isSidebarOpen && 'justify-center')

  return (
    <aside
      className={cn(
        'flex flex-none flex-col gap-24 w-full overflow-y-auto p-4 transition-all',
        'my-4 ml-4 rounded-lg shadow-lg backdrop-blur-md',
        'bg-cool-grey-100/70 dark:bg-dark-grey-800/70',
        {
          'max-w-56': isSidebarOpen,
          'max-w-14': !isSidebarOpen,
        }
      )}
    >
      <div className="flex flex-col gap-6">
        <Block
          className={cn('h-[20px]', !isSidebarOpen && 'justify-center')}
          title="Nuon"
          icon="CubeIcon"
          iconSize={isSidebarOpen ? 20 : 24}
          text="Nuon"
          collapsed={!isSidebarOpen}
        />

        <Block
          className={cn('w-full transition-all', {
            'h-[4rem]': isSidebarOpen,
            'h-[2.5rem] justify-center': !isSidebarOpen,
          })}
          title="org-switcher"
          icon="BuildingsIcon"
          iconSize={isSidebarOpen ? 20 : 24}
          text="acme"
          collapsed={!isSidebarOpen}
        />
      </div>

      <nav className="flex flex-col gap-4">
        {primaryNav.map((item) => (
          <NavBlock
            key={item.path}
            {...item}
            iconSize={iconSize}
            collapsed={!isSidebarOpen}
            className={itemClass}
          />
        ))}
      </nav>

      {secondaryNav.length > 0 && (
        <nav className="flex flex-col gap-4 mt-auto">
          {secondaryNav.map((item) => (
            <NavBlock
              key={item.path}
              {...item}
              iconSize={iconSize}
              collapsed={!isSidebarOpen}
              className={itemClass}
            />
          ))}
        </nav>
      )}
    </aside>
  )
}
