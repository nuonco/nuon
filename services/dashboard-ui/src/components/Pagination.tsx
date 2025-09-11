'use client'

import classNames from 'classnames'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import React, { type FC } from 'react'
import { ArrowLeft, ArrowRight } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { useOrg } from '@/hooks/use-org'
import type { TPagination } from '@/lib'

interface IPagination {
  limit?: number
  param?: string
  pageData?: TPagination
  position?: 'center' | 'left' | 'right'
}

export const Pagination: FC<IPagination> = ({
  limit = 10,
  param = 'offset',
  pageData = {
    hasNext: 'false',
    offset: '0',
  },
  position = 'center',
}) => {
  const { org } = useOrg()
  const pathname = usePathname()
  const router = useRouter()
  const searchParams = useSearchParams()
  const offset = parseInt(pageData.offset)
  const hasNext = Boolean(pageData.hasNext === 'true')

  // Helper to update the offset param, preserving others
  const buildPathWithOffset = (newOffset: number) => {
    const params = new URLSearchParams(searchParams.toString())
    if (newOffset === 0) {
      params.delete(param)
    } else {
      params.set(param, String(newOffset))
    }
    return `${pathname}?${params.toString()}`
  }

  return org?.features?.['api-pagination'] ? (
    <div
      className={classNames('flex items-center gap-3', {
        'self-center': position === 'center',
        'self-end': position === 'right',
        'self-start': position === 'left',
      })}
    >
      {offset === 0 && !hasNext ? null : (
        <>
          <Button
            disabled={offset === 0}
            onClick={() => {
              const newOffset = offset === limit ? 0 : offset - limit
              router.push(buildPathWithOffset(newOffset))
            }}
            className="text-sm flex items-center gap-1 !p-2"
            title="previous"
          >
            <ArrowLeft />
          </Button>

          <Button
            disabled={!hasNext}
            onClick={() => {
              const newOffset = offset === 0 ? limit : offset + limit
              router.push(buildPathWithOffset(newOffset))
            }}
            className="text-sm flex items-center gap-1 !p-2"
            title="next"
          >
            <ArrowRight />
          </Button>
        </>
      )}
    </div>
  ) : null
}
