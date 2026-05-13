import { useSearchParams } from 'react-router'
import React, { useEffect, useRef, useState } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { useAutoFocusOnVisible } from '@/hooks/use-auto-focus-on-visible'
import { useEscapeKey } from '@/hooks/use-escape-key'
import { useSurfaces } from '@/hooks/use-surfaces'
import { Tooltip } from '@/components/common/Tooltip'
import { cn } from '@/utils/classnames'
import './Panel.css'

type TPanelSize = 'default' | 'half' | '3/4' | 'full'

export interface IPanel extends React.HTMLAttributes<HTMLDivElement> {
  childrenClassName?: string
  defaultExpanded?: boolean
  heading?: React.ReactNode
  headerClassName?: string
  isVisible?: boolean
  onClose?: () => void
  onSizeChange?: (size: TPanelSize) => void
  panelId?: string
  panelKey?: string
  size?: TPanelSize
  triggerButton?: Omit<IButtonAsButton, 'onClick'>
}

export const PanelBase = ({
  className,
  children,
  childrenClassName,
  defaultExpanded,
  heading,
  headerClassName,
  isVisible = false,
  onClose,
  onSizeChange,
  panelId,
  panelKey,
  size: initSize = 'default',
  ...props
}: Omit<IPanel, 'triggerButton'>) => {
  const [size, setSize] = useState(defaultExpanded ? 'full' : initSize)
  const { removePanel, panels } = useSurfaces()
  const handleClose = () => {
    if (onClose) onClose?.()
    removePanel(panels?.at(-1)?.id, panelKey)
  }
  const panelRef = useRef<HTMLDivElement>(null)
  // auto focus panel when in view
  useAutoFocusOnVisible(panelRef, isVisible)
  // handle close on esc key
  useEscapeKey(handleClose)

  return (
    <>
      <TransitionDiv
        className="panel-wrapper absolute top-0 left-0 w-screen h-screen flex z-10"
        isVisible={isVisible}
      >
        <div
          className="panel-overlay backdrop-blur-xs bg-black/2 dark:bg-black/10 absolute top-0 left-0 w-screen h-screen flex"
          onClick={handleClose}
        />
        <section
          className={cn(
            'panel fixed h-screen top-0 right-0 border flex flex-col drop-shadow-2xl overflow-y-auto overflow-x-hidden',
            'bg-white dark:bg-dark-grey-900',
            {
              'w-screen md:w-104': size === 'default',
              'w-screen md:w-1/2': size === 'half',
              'w-screen md:w-3/4': size === '3/4',
              'w-screen': size === 'full',
            },
            className
          )}
          role="complementary"
          ref={panelRef}
          tabIndex={-1}
          {...props}
        >
          <header
            className={cn(
              'flex items-center justify-between shrink-0 h-18 px-4 md:px-6',
              headerClassName
            )}
          >
            {heading ? (
              typeof heading === 'string' ? (
                <HeadingGroup>
                  <Text variant="base" weight="strong">
                    {heading}
                  </Text>
                </HeadingGroup>
              ) : (
                <HeadingGroup>{heading}</HeadingGroup>
              )
            ) : null}
            <div className="flex items-center gap-4">
              {initSize !== 'full' ? (
                <Tooltip
                  position="bottom"
                  tipContent={
                    <Text variant="subtext">
                      {size === 'full'
                        ? `Resize to ${initSize} size`
                        : 'Expand to full screen'}
                    </Text>
                  }
                >
                  <Button
                    className="!p-2 ml-auto"
                    variant="ghost"
                    onClick={() => {
                      setSize((prev: TPanelSize) => {
                        const next = prev === initSize ? 'full' : initSize
                        onSizeChange?.(next)
                        return next
                      })
                    }}
                    aria-label={
                      size === 'full'
                        ? `Resize to ${initSize} size`
                        : 'Expand to full screen'
                    }
                  >
                    <Icon
                      variant={size === 'full' ? 'CornersIn' : 'CornersOut'}
                    />
                  </Button>
                </Tooltip>
              ) : null}
              <Tooltip
                position="bottom"
                tipContent={
                  <Text variant="subtext">Close panel</Text>
                }
              >
                <Button
                  className="!p-2 ml-auto"
                  variant="ghost"
                  onClick={handleClose}
                  aria-label="Close panel"
                >
                  <Icon variant="ArrowLineRight" />
                </Button>
              </Tooltip>
            </div>
          </header>
          <div
            className={cn(
              'px-4 md:px-6 pb-4 md:pb-6 flex flex-col flex-auto gap-4 md:gap-6',
              childrenClassName
            )}
          >
            {children}
          </div>
        </section>
      </TransitionDiv>
    </>
  )
}

export const Panel = ({ triggerButton, ...props }: IPanel) => {
  const { addPanel } = useSurfaces()
  const [searchParams] = useSearchParams()
  const openedByTrigger = useRef(false)
  const panel = <PanelBase {...props} />

  const handleAddPanel = () => {
    addPanel(panel, props.panelKey)
  }

  const panelParam = searchParams?.get('panel')
  useEffect(() => {
    if (openedByTrigger.current) {
      openedByTrigger.current = false
      return
    }
    if (
      props.panelKey &&
      props.panelKey === panelParam &&
      !props.isVisible
    ) {
      // Defer so this runs after the pathname-change panel clear in SurfacesProvider
      const timer = setTimeout(() => handleAddPanel(), 0)
      return () => clearTimeout(timer)
    }
  }, [panelParam])

  const handleTriggerClick = () => {
    openedByTrigger.current = true
    handleAddPanel()
  }

  return (
    <>
      {triggerButton ? (
        <Button onClick={handleTriggerClick} {...triggerButton}>
          {triggerButton.children}
        </Button>
      ) : (
        panel
      )}
    </>
  )
}
