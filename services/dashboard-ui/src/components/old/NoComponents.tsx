import React, { type FC } from 'react'
import { Link } from '@/components/old/Link'
import { Empty } from './Empty'
import { ArrowSquareOut } from '@phosphor-icons/react/dist/ssr'

export const NoComponents: FC = () => {
  return (
    <div className="flex flex-auto flex-col items-center justify-center">
      <Empty
        emptyTitle="No components yet"
        emptyMessage="Model your app by configuring components. Check out our resources."
        variant="table"
      />
      <br />
      <Link
        className="flex gap-2 text-sm items-center"
        href="https://docs.nuon.co/concepts/components"
        target="_blank"
      >
        Learn more <ArrowSquareOut size="14" />
      </Link>
    </div>
  )
}
