import {
  useId,
  useLayoutEffect,
  useRef,
  useState,
  type HTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { useFocusContainment } from '../../../hooks/use-focus-containment'
import { Button } from '../../atoms/Button'
import { Card } from '../../atoms/Card'
import { Icon } from '../../atoms/Icon'
import { Text } from '../../atoms/Text'
import { SurfaceOverlay } from './SurfaceOverlay'
import { useSurfaceInstance } from './SurfaceHost'
import { SurfaceTransition } from './SurfaceTransition'

export type TPanelSize = 'default' | 'half' | 'wide' | 'full'

export interface IPanel
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  bodyClassName?: string
  children: ReactNode
  defaultSize?: TPanelSize
  expandable?: boolean
  heading: ReactNode
  headerActions?: ReactNode
  onClose?: () => void
  onSizeChange?: (size: TPanelSize) => void
  size?: TPanelSize
}

const SIZE_CLASSES: Record<TPanelSize, string> = {
  default: 'md:w-[28rem]',
  half: 'md:w-1/2',
  wide: 'md:w-3/4',
  full: 'md:w-[calc(100vw-1.5rem)]',
}

export const Panel = ({
  bodyClassName,
  children,
  className,
  defaultSize = 'default',
  expandable = true,
  heading,
  headerActions,
  onClose,
  onSizeChange,
  size: controlledSize,
  ...props
}: IPanel) => {
  const instance = useSurfaceInstance()
  const [uncontrolledSize, setUncontrolledSize] =
    useState<TPanelSize>(defaultSize)
  const size = controlledSize ?? uncontrolledSize
  const panelRef = useRef<HTMLDivElement>(null)
  const headingId = useId()

  const close = () => {
    onClose?.()
    instance.close()
  }

  const toggleSize = () => {
    const next = size === 'full' ? defaultSize : 'full'
    if (controlledSize === undefined) setUncontrolledSize(next)
    onSizeChange?.(next)
  }

  useFocusContainment({
    active: instance.topmost && instance.visible,
    containerRef: panelRef,
    initialFocus: 'container',
    onEscape: close,
  })

  useLayoutEffect(() => {
    if (panelRef.current) panelRef.current.inert = !instance.topmost
  }, [instance.topmost])

  return (
    <div
      className="pointer-events-none fixed inset-0 flex justify-end p-3"
      style={{ zIndex: instance.zIndex }}
    >
      <SurfaceOverlay
        visible={instance.visible}
        topmost={instance.topmost}
        onClose={close}
      />
      <SurfaceTransition
        variant="panel"
        visible={instance.visible}
        coveredBy={instance.coveredBy}
        className={cn('h-full w-[calc(100vw-1.5rem)]', SIZE_CLASSES[size])}
      >
        <Card
          ref={panelRef}
          role="dialog"
          aria-modal={instance.topmost ? 'true' : undefined}
          aria-labelledby={headingId}
          aria-hidden={!instance.topmost || undefined}
          tabIndex={-1}
          padding="none"
          blur="lg"
          opacity="strong"
          shadow="floating"
          className={cn(
            'flex h-full w-full flex-col overflow-hidden outline-none',
            className
          )}
          {...props}
        >
          <header className="flex min-h-14 shrink-0 items-center justify-between gap-4 px-4 py-3 sm:px-6">
            <div id={headingId} className="min-w-0">
              {typeof heading === 'string' ? (
                <Text as="h2" variant="heading">
                  {heading}
                </Text>
              ) : (
                heading
              )}
            </div>
            <div className="flex shrink-0 items-center gap-2">
              {headerActions}
              {expandable && defaultSize !== 'full' ? (
                <Button
                  variant="ghost"
                  size="sm"
                  iconOnly
                  aria-label={
                    size === 'full'
                      ? `Resize panel to ${defaultSize}`
                      : 'Expand panel to full screen'
                  }
                  tooltip={
                    size === 'full'
                      ? `Resize to ${defaultSize}`
                      : 'Expand to full screen'
                  }
                  tooltipSide="bottom"
                  onClick={toggleSize}
                >
                  <Icon
                    variant={
                      size === 'full' ? 'CornersInIcon' : 'CornersOutIcon'
                    }
                    size={18}
                  />
                </Button>
              ) : null}
              <Button
                variant="ghost"
                size="sm"
                iconOnly
                aria-label="Close panel"
                tooltip="Close panel"
                tooltipSide="bottom"
                onClick={close}
              >
                <Icon variant="ArrowLineRightIcon" size={18} />
              </Button>
            </div>
          </header>
          <div
            className={cn(
              'flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-4 sm:p-6',
              bodyClassName
            )}
          >
            {children}
          </div>
        </Card>
      </SurfaceTransition>
    </div>
  )
}
