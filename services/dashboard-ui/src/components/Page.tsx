import React, { type FC } from 'react'
import { Heading, Nav, type TLink, Text } from '@/components'

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

export const PageTitle: FC<{ overline?: string; title: string }> = ({
  overline,
  title,
}) => {
  return (
    <hgroup className="flex flex-col gap-2 max-w-xl">
      {overline ? <Text variant="overline">{overline}</Text> : null}
      <Heading level={1} variant="title">
        {title}
      </Heading>
    </hgroup>
  )
}

export const PageSummary: FC<{ children?: React.ReactNode | any }> = ({
  children,
}) => {
  return (
    <Text className="flex flex-wrap gap-4 items-center" variant="caption">
      {children}
    </Text>
  )
}
