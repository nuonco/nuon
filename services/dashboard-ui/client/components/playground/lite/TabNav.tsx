import { NavBlock } from './NavBlock'
import { labelWidth } from './utils'
import type { INavItem } from './types'

export interface ITabNav {
  tabs: INavItem[]
}

export const TabNav = ({ tabs }: ITabNav) => (
  <nav className="flex gap-6 items-center">
    {tabs.map((tab) => (
      <NavBlock
        key={tab.path}
        {...tab}
        exact
        className="h-[20px]"
        style={{ width: labelWidth(tab.label) }}
      />
    ))}
  </nav>
)
