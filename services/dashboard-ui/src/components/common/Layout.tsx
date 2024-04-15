import React, { type FC } from 'react'
import { Header, Heading, Nav, type TLink } from '@/components'

export const Dashboard: FC<{ children: React.ReactElement }> = ({
  children,
}) => {
  return (
    <div className="flex flex-col gap-6 p-6 xl:px-24 w-full h-dvh overflow-hidden">
      <Header />
      {children}
    </div>
  )
}

export const Page: FC<{
  children: React.ReactElement | any
  heading: React.ReactElement
  links?: Array<TLink>
}> = ({ children, heading, links }) => {
  return (
    <main className="flex flex-col flex-auto items-start justify-start gap-6 w-full h-fit overflow-hidden">
      <div className="flex flex-col gap-6 w-full">
        {links && <Nav links={links} />}
        {heading}
      </div>
      <div className="flex flex-auto flex-col gap-6 w-full h-fit overflow-auto">
        {children}
      </div>
    </main>
  )
}
