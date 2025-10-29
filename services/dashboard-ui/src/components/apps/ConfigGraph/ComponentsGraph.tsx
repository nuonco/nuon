'use client'

import dynamic from 'next/dynamic'
import type { ComponentProps } from 'react'
import { Button } from '@/components/common/Button'

const ComponentsGraph = dynamic<
  ComponentProps<
    typeof import('./ComponentsGraphRenderer').ComponentsGraphRenderer
  >
>(
  () =>
    import('./ComponentsGraphRenderer').then((mod) => ({
      default: mod.ComponentsGraphRenderer,
    })),
  {
    ssr: false,
    loading: () => (
      <Button disabled variant="ghost">
        Loading dependency graph...
      </Button>
    ),
  }
)

export { ComponentsGraph }
