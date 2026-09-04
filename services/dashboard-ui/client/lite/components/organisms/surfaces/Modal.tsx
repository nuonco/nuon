import {
  useId,
  useLayoutEffect,
  useRef,
  type HTMLAttributes,
  type ReactNode,
  type RefObject,
} from 'react'
import { cn } from '@/utils/classnames'
import { useFocusContainment } from '../../../hooks/use-focus-containment'
import { Button, type IButton } from '../../atoms/Button'
import { Card } from '../../atoms/Card'
import { Icon } from '../../atoms/Icon'
import { Text } from '../../atoms/Text'
import { SurfaceOverlay } from './SurfaceOverlay'
import { useSurfaceInstance } from './SurfaceHost'
import { SurfaceTransition } from './SurfaceTransition'

export type TModalSize = 'sm' | 'default' | 'lg' | 'xl' | 'full'

export interface IModal
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  bodyClassName?: string
  children: ReactNode
  description?: ReactNode
  dismissible?: boolean
  footerContent?: ReactNode
  heading: ReactNode
  headerActions?: ReactNode
  initialFocusRef?: RefObject<HTMLElement | null>
  onClose?: () => void
  primaryAction?: IButton
  secondaryAction?: IButton
  showFooter?: boolean
  size?: TModalSize
}

const SIZE_CLASSES: Record<TModalSize, string> = {
  sm: 'max-w-sm',
  default: 'max-w-xl',
  lg: 'max-w-3xl',
  xl: 'max-w-5xl',
  full: 'max-w-[calc(100vw-2rem)]',
}

export const Modal = ({
  bodyClassName,
  children,
  className,
  description,
  dismissible = true,
  footerContent,
  heading,
  headerActions,
  initialFocusRef,
  onClose,
  primaryAction,
  secondaryAction,
  showFooter = true,
  size = 'default',
  ...props
}: IModal) => {
  const instance = useSurfaceInstance()
  const modalRef = useRef<HTMLDivElement>(null)
  const headingId = useId()

  const close = () => {
    onClose?.()
    instance.close()
  }

  useFocusContainment({
    active: instance.topmost && instance.visible,
    containerRef: modalRef,
    initialFocusRef,
    onEscape: dismissible ? close : undefined,
  })

  useLayoutEffect(() => {
    if (modalRef.current) modalRef.current.inert = !instance.topmost
  }, [instance.topmost])

  return (
    <div
      className="pointer-events-none fixed inset-0 flex items-center justify-center p-4"
      style={{ zIndex: instance.zIndex }}
    >
      <SurfaceOverlay
        visible={instance.visible}
        topmost={instance.topmost}
        onClose={dismissible ? close : undefined}
      />
      <SurfaceTransition
        variant="modal"
        visible={instance.visible}
        coveredBy={instance.coveredBy}
        className={cn(
          'flex max-h-[calc(100vh-2rem)] w-full',
          SIZE_CLASSES[size]
        )}
      >
        <Card
          ref={modalRef}
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
            'flex max-h-full min-h-0 w-full flex-col overflow-hidden',
            className
          )}
          {...props}
        >
          <header className="flex shrink-0 items-start justify-between gap-4 px-4 pt-4 pb-2 sm:px-6 sm:pt-6">
            <div className="flex min-w-0 flex-col gap-1">
              <div id={headingId}>
                {typeof heading === 'string' ? (
                  <Text as="h2" variant="heading">
                    {heading}
                  </Text>
                ) : (
                  heading
                )}
              </div>
              {description ? (
                typeof description === 'string' ? (
                  <Text color="secondary">{description}</Text>
                ) : (
                  description
                )
              ) : null}
            </div>
            <div className="flex shrink-0 items-center gap-2">
              {headerActions}
              {dismissible ? (
                <Button
                  variant="ghost"
                  size="sm"
                  iconOnly
                  aria-label="Close modal"
                  onClick={close}
                >
                  <Icon variant="XIcon" size={18} />
                </Button>
              ) : null}
            </div>
          </header>
          <div
            className={cn(
              'flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-4 py-3 sm:px-6 sm:py-4',
              bodyClassName
            )}
          >
            {children}
          </div>
          {showFooter ? (
            <footer className="flex shrink-0 items-center justify-between gap-4 px-4 pt-2 pb-4 sm:px-6 sm:pb-6">
              <div className="flex min-w-0 items-center gap-2">
                {footerContent}
              </div>
              <div className="flex shrink-0 items-center gap-2">
                {secondaryAction ? (
                  <Button {...secondaryAction} />
                ) : (
                  <Button variant="secondary" onClick={close}>
                    {primaryAction ? 'Cancel' : 'Close'}
                  </Button>
                )}
                {primaryAction ? <Button {...primaryAction} /> : null}
              </div>
            </footer>
          ) : null}
        </Card>
      </SurfaceTransition>
    </div>
  )
}
