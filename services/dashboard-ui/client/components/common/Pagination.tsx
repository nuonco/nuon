'use client'

import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { startTransition } from 'react'
import { usePagination } from '@/hooks/use-pagination'
import { cn } from '@/utils/classnames'
import { Button } from './Button'
import { Icon } from './Icon'

export interface IPagination {
  hasNext?: boolean
  limit?: number
  offset?: number
  param?: string
  position?: 'center' | 'left' | 'right'
}

export const Pagination = ({
  limit = 10,
  hasNext = false,
  offset = 0,
  param = 'offset',
  position = 'center',
}: IPagination) => {
  const pathname = usePathname()
  const router = useRouter()
  const searchParams = useSearchParams()
  const { isPaginating, setIsPaginating } = usePagination()

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

  // Handlers
  const handlePrev = () => {
    setIsPaginating(true)
    startTransition(() => {
      const newOffset = Math.max(offset - limit, 0)
      router.push(buildPathWithOffset(newOffset))
    })
  }

  const handleNext = () => {
    setIsPaginating(true)
    startTransition(() => {
      const newOffset = offset + limit
      router.push(buildPathWithOffset(newOffset))
    })
  }

  return (
    <div
      className={cn('flex items-center gap-3', {
        'self-center': position === 'center',
        'self-end': position === 'right',
        'self-start': position === 'left',
      })}
    >
      <Button
        disabled={offset === 0 || isPaginating}
        onClick={handlePrev}
        title="previous"
      >
        <Icon variant="ArrowLeft" />
      </Button>

      <Button
        disabled={!hasNext || isPaginating}
        onClick={handleNext}
        title="next"
      >
        <Icon variant="ArrowRight" />
      </Button>
    </div>
  )
}
