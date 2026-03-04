import { lazy, Suspense, type ComponentProps } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

const ComponentsGraphRenderer = lazy(() =>
  import('./ComponentsGraphRenderer').then((mod) => ({
    default: mod.ComponentsGraphRenderer,
  }))
)

const LoadingFallback = () => (
  <Button disabled variant="ghost" isMenuButton>
    <span className="flex items-center gap-2">
      <Icon variant="Loading" />
      Loading graph...
    </span>
  </Button>
)

type ComponentsGraphProps = ComponentProps<
  typeof import('./ComponentsGraphRenderer').ComponentsGraphRenderer
>

const ComponentsGraph = (props: ComponentsGraphProps) => (
  <Suspense fallback={<LoadingFallback />}>
    <ComponentsGraphRenderer {...props} />
  </Suspense>
)

export { ComponentsGraph }
