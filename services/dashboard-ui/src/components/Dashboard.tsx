import classNames from 'classnames'
import React, { type FC } from 'react'
import {
  BreadcrumbNav,
  ClickToCopy,
  Heading,
  Logo,
  Link,
  SignOutButton,
  Text,
  type TLink,
} from '@/components'

export const DashboardHeader: FC = () => {
  return (
    <header className="flex flex-wrap items-center justify-between gap-6 pb-6 border-b">
      <div className="flex items-center gap-6">
        <Logo />
      </div>

      <div className="flex gap-4 items-center">
        <Link className="text-sm" href="https://docs.nuon.co" target="_blank">
          Documentation
        </Link>

        <SignOutButton />
      </div>
    </header>
  )
}

export const Dashboard: FC<{ children: React.ReactElement }> = ({
  children,
}) => {
  return (
    <div className="flex flex-col gap-6 p-6 xl:px-24 w-full h-dvh overflow-auto">
      <DashboardHeader />
      {children}
    </div>
  )
}

export const DashboardContent: FC<{
  breadcrumb: Array<TLink>
  children: React.ReactElement
  heading?: React.ReactElement | string
  headingUnderline?: React.ReactElement | string
  statues?: React.ReactElement | null
  meta?: React.ReactElement | null
}> = ({
  breadcrumb,
  children,
  heading,
  headingUnderline,
  statues = null,
  meta = null,
}) => {
  return (
    <>
      <header className="flex justify-between items-center border-b px-6 py-4 h-[75px] bg-white/50 dark:bg-white/[.02]">
        <BreadcrumbNav links={breadcrumb} />
        <div>
          <Link href="https://docs.nuon.co" target="_blank" variant="ghost">
            Docs
          </Link>
        </div>
      </header>
      <main
        className="overflow-x-auto flex flex-col"
        style={{ height: 'calc(100% - 75px)' }}
      >
        {heading && (
          <header
            className={classNames(
              'px-6 pt-8 flex flex-col pt-6 gap-6 border-b',
              {
                'pb-8': !Boolean(meta),
              }
            )}
          >
            <div className="flex items-start justify-between">
              <hgroup className="flex flex-col gap-2">
                <Heading variant="title">{heading}</Heading>
                <ClickToCopy>
                  <Text variant="id">{headingUnderline}</Text>
                </ClickToCopy>
              </hgroup>

              {statues}
            </div>
            {meta}
          </header>
        )}

        {children}
      </main>
    </>
  )
}
