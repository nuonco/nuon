import React, { type FC } from 'react'
import { Header, Heading, Nav, type TLink } from '@/components'

export const Dashboard: FC<{ children: React.ReactElement }> = ({
  children,
}) => {
  return (
    <div className="flex flex-col gap-6 p-6 xl:px-24 w-full h-dvh overflow-auto">
      <Header />
      {children}
    </div>
  )
}

export const Page: FC<{
  children: React.ReactNode | any
  header: React.ReactNode
  links?: Array<TLink>
}> = ({ children, header, links }) => {
  return (
    <main className="flex flex-col flex-auto items-start justify-start gap-6 w-full h-fit">
      <div className="flex flex-col gap-6 w-full">
        {links && <Nav links={links} />}
        {header}
      </div>
      <div className="flex flex-auto flex-col gap-6 w-full h-fit">
        {children}
      </div>
    </main>
  )
}

export const PageHeader: FC<{
  info?: React.ReactNode | null
  summary?: React.ReactNode | null
  title: React.ReactNode
}> = ({ info = null, summary = null, title }) => {
  return (
    <header className="flex flex-wrap gap-8 items-end border-b pb-6">
      <div className="flex flex-col flex-auto gap-4">
        {title}
        {summary}
      </div>

      <div className="flex-auto">{info}</div>
    </header>
  )
}
