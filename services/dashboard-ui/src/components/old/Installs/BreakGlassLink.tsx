'use client'

import React, { type FC } from 'react'
import { LockLaminated } from '@phosphor-icons/react'
import { Link } from '@/components/old/Link'
import { useOrg } from '@/hooks/use-org'

export const BreakGlassLink: FC<{ installId: string }> = ({ installId }) => {
  const { org } = useOrg()

  return org?.features?.['install-break-glass'] ? (
    <Link
      className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
      href={`/${org.id}/installs/${installId}/break-glass`}
      variant="ghost"
    >
      <LockLaminated size="16" />
      Break glass permissions
    </Link>
  ) : null
}
