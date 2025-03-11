'use client'

import classNames from 'classnames'
import { usePathname, useRouter } from 'next/navigation'
import React, { type FC } from 'react'
import { ArrowLeft, ArrowRight } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { useOrg } from '@/components/Orgs'
import type { TPagination } from '@/lib'

interface IPagination {
  param?: string
  pageData?: TPagination
  position?: 'center' | 'left' | 'right'
}

export const Pagination: FC<IPagination> = ({
  param,
  pageData = {
    hasNext: 'false',
    offset: '0',
  },
  position = 'center',
}) => {
  const { org } = useOrg()
  const pathname = usePathname()
  const router = useRouter()
  const offset = parseInt(pageData.offset)
  const hasNext = Boolean(pageData.hasNext === 'true')

  return org?.features?.['api-pagination'] ? (
    <div
      className={classNames('flex items-center gap-3', {
        'self-center': position === 'center',
        'self-end': position === 'right',
        'self-start': position === 'left',
      })}
    >
      <Button
        disabled={offset === 0}
        onClick={() => {
          const path = `${pathname}?${param}=${offset - 1}`
          router.push(path)
        }}
        className="text-sm flex items-center gap-1 !p-2"
        title="previous"
      >
        <ArrowLeft />
      </Button>

      <Button
        disabled={!hasNext}
        onClick={() => {
          const path = `${pathname}?${param}=${offset + 1}`
          router.push(path)
        }}
        className="text-sm flex items-center gap-1 !p-2"
        title="next"
      >
        <ArrowRight />
      </Button>
    </div>
  ) : null
}
