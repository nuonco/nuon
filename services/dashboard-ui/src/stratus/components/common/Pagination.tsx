'use client'

import { usePathname, useRouter } from 'next/navigation'
import React, { type FC } from 'react'
import { cn } from '@/stratus/components/helpers'
import type { TPaginationPageData } from '@/types'
import { Button } from './Button'
import { Icon } from './Icon'

export interface IPagination {
  limit?: number
  param?: string
  pageData?: TPaginationPageData
  position?: 'center' | 'left' | 'right'
}

export const Pagination: FC<IPagination> = ({
  limit = 10,
  param,
  pageData = {
    hasNext: 'false',
    offset: '0',
  },
  position = 'center',
}) => {
  const pathname = usePathname()
  const router = useRouter()
  const offset = parseInt(pageData.offset)
  const hasNext = Boolean(pageData.hasNext === 'true')

  return (
    <div
      className={cn('flex items-center gap-3', {
        'self-center': position === 'center',
        'self-end': position === 'right',
        'self-start': position === 'left',
      })}
    >
      <Button
        disabled={offset === 0}
        onClick={() => {
          const path = `${pathname}?${param}=${offset === limit + 1 ? 0 : offset - limit}`
          router.push(path)
        }}
        title="previous"
      >
        <Icon variant="ArrowLeft" />
      </Button>

      <Button
        disabled={!hasNext}
        onClick={() => {
          const path = `${pathname}?${param}=${offset === 0 ? limit + 1 : offset + limit}`
          router.push(path)
        }}
        title="next"
      >
        <Icon variant="ArrowRight" />
      </Button>
    </div>
  )
}
