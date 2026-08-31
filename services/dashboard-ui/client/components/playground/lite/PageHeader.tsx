import { Fragment } from 'react'
import { Block } from './Block'
import { LinkBlock } from './LinkBlock'
import { Utilities } from './Utilities'
import { useShell } from './shell-context'
import type { ICrumb } from './types'
import { labelWidth } from './utils'

export interface IPageHeader {
  crumbs?: ICrumb[]
}

export const PageHeader = ({ crumbs = [] }: IPageHeader) => {
  const { toggleSidebar } = useShell()

  return (
    <header className="flex flex-none justify-between items-center gap-4 w-full">
      <div className="flex min-w-0 items-center gap-4">
        <Block
          className="h-[20px] cursor-pointer"
          title="Toggle sidebar"
          icon="SidebarSimpleIcon"
          iconSize={20}
          onClick={toggleSidebar}
        />
        <nav className="flex gap-2 items-center">
          <LinkBlock
            path="/"
            label="acme"
            className="h-[24px]"
            style={{ width: 100 }}
          />
          {crumbs.map((crumb, i) => (
            <Fragment key={crumb.label}>
              <span className="opacity-40">/</span>
              {crumb.path && i < crumbs.length - 1 ? (
                <LinkBlock
                  path={crumb.path}
                  label={crumb.label}
                  className="h-[24px]"
                  style={{ width: labelWidth(crumb.label) }}
                />
              ) : (
                <Block
                  className="h-[24px]"
                  style={{ width: labelWidth(crumb.label) }}
                  title={crumb.label}
                  text={crumb.label}
                />
              )}
            </Fragment>
          ))}
        </nav>
      </div>

      <Utilities />
    </header>
  )
}
