'use client'

import dynamic from 'next/dynamic'
import type { ComponentProps } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

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
      <Button disabled variant="ghost" isMenuButton>
        <span className="flex items-center gap-2">
          <Icon variant="Loading" />
          Loading graph...
        </span>
      </Button>
    ),
  }
)

export { ComponentsGraph }
