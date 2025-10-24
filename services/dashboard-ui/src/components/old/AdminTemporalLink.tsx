'use client'

import React, { type FC } from 'react'
import { ArrowSquareOut } from '@phosphor-icons/react'
import { Link } from '@/components/old/Link'

export const AdminTemporalLink: FC<{ namespace: string; id: string }> = ({
  id,
  namespace,
}) => {
  return (
    <Link
      href={`/admin/temporal/namespaces/${namespace}/workflows/event-loop-${id}`}
      className="text-base gap-2"
      target="_blank"
    >
      View in Temporal <ArrowSquareOut />
    </Link>
  )
}
