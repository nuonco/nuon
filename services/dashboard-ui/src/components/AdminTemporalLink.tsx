'use client'

import React, { type FC } from 'react'
import { ArrowSquareOut } from '@phosphor-icons/react'
import { Link } from '@/components/Link'

export const AdminTemporalLink: FC<{ namespace: string; id: string }> = ({
  id,
  namespace,
}) => {
  const env = window?.['env'] || 'prod'

  return (
    <Link
      href={
        env === 'local'
          ? `http://localhost:8233/namespaces/${namespace}/workflows/event-loop-${id}`
          : `http://temporal-ui.nuon.us-west-2.${env}.nuon.cloud:8080/namespaces/${namespace}/workflows/event-loop-${id}`
      }
      className="text-base gap-2"
      target="_blank"
    >
      View in Temporal <ArrowSquareOut />
    </Link>
  )
}
