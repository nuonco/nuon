import React, { type FC } from 'react'
import { ArrowSquareOut } from '@phosphor-icons/react/dist/ssr'
import { Empty } from '@/components/old/Empty'
import { Link } from '@/components/old/Link'

export const NoActions: FC = () => {
  return (
    <div className="flex flex-auto flex-col items-center justify-center -translate-y-6">
      <Empty
        emptyTitle="No actions yet"
        emptyMessage="Save time by configuring your action workflows. Check out our resources."
        variant="actions"
      />
      <br />
      <Link
        className="flex gap-2 text-sm items-center"
        href="https://docs.nuon.co/concepts/actions"
        target="_blank"
      >
        Learn more <ArrowSquareOut size="14" />
      </Link>
    </div>
  )
}
