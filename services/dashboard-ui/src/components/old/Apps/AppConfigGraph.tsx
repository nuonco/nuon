'use client'

import dynamic from 'next/dynamic'
import type { ComponentProps } from 'react'
import { Button } from '@/components/old/Button'

const AppConfigGraph = dynamic<
  ComponentProps<
    typeof import('./AppConfigGraphRenderer').AppConfigGraphRenderer
  >
>(
  () =>
    import('./AppConfigGraphRenderer').then((mod) => ({
      default: mod.AppConfigGraphRenderer,
    })),
  {
    ssr: false,
    loading: () => (
      <Button className="text-sm" disabled>
        Loading dependency graph...
      </Button>
    ),
  }
)

export { AppConfigGraph }
