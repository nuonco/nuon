'use client'

import { useSearchParams } from 'next/navigation'
import React, { useEffect, useRef, useState } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { useAutoFocusOnVisible } from '@/hooks/use-auto-focus-on-visible'
import { useEscapeKey } from '@/hooks/use-escape-key'
import { useSurfaces } from '@/hooks/use-surfaces'
import { cn } from '@/utils/classnames'
import './Panel.css'

type TPanelSize = 'default' | 'half' | 'full'

export interface IPanel extends React.HTMLAttributes<HTMLDivElement> {
  childrenClassName?: string
  heading?: React.ReactNode
  isVisible?: boolean
  onClose?: () => void
  panelId?: string
  panelKey?: string
  size?: TPanelSize
  triggerButton?: Omit<IButtonAsButton, 'onClick'>
}

const PanelBase = ({
  className,
  children,
  childrenClassName,
  heading,
  isVisible = false,
  onClose,
  panelId,
  panelKey,
  size: initSize = 'default',
  ...props
}: Omit<IPanel, 'triggerButton'>) => {
  const [size, setSize] = useState(initSize)
  const { removePanel } = useSurfaces()
  const handleClose = () => {
    if (onClose) onClose?.()
    removePanel(panelId, panelKey)
  }
  const panelRef = useRef<HTMLDivElement>(null)
  // auto focus panel when in view
  useAutoFocusOnVisible(panelRef, isVisible)
  // handle close on esc key
  useEscapeKey(handleClose)

  return (
    <>
      <TransitionDiv
        className="panel-wrapper absolute top-0 left-0 w-screen h-screen flex"
        isVisible={isVisible}
      >
        <div
          className="panel-overlay bg-black/2 dark:bg-black/10 absolute top-0 left-0 w-screen h-screen flex"
          onClick={handleClose}
        />
        <section
          className={cn(
            'panel fixed h-screen top-0 right-0 border flex flex-col drop-shadow-2xl overflow-y-auto overflow-x-hidden',
            'bg-white dark:bg-dark-grey-900',
            {
              'w-screen md:w-104': size === 'default',
              'w-screen md:w-1/2': size === 'half',
              'w-screen': size === 'full',
            },
            className
          )}
          role="complementary"
          ref={panelRef}
          tabIndex={-1}
          {...props}
        >
          <header className="flex items-center justify-between shrink-0 h-18 px-4">
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
                <Button
                  className="!p-2 ml-auto"
                  variant="ghost"
                  onClick={() => {
                    setSize((prev: TPanelSize) => {
                      if (prev === initSize) {
                        return 'full'
                      } else {
                        return initSize
                      }
                    })
                  }}
                  title={
                    size === 'full'
                      ? `Resize to ${initSize} size`
                      : 'Expand to full screen'
                  }
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
              ) : null}
              <Button
                className="!p-2 ml-auto"
                variant="ghost"
                onClick={handleClose}
                title="Close panel"
                aria-label="Close panel"
              >
                <Icon variant="ArrowLineRight" />
              </Button>
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
  const searchParams = useSearchParams()
  const panel = <PanelBase {...props} />

  const handleAddPanel = () => {
    addPanel(panel, props.panelKey)
  }

  useEffect(() => {
    if (
      props.panelKey &&
      props.panelKey === searchParams?.get('panel') &&
      !props.isVisible
    ) {
      handleAddPanel()
    }
  }, [])

  return (
    <>
      {triggerButton ? (
        <Button onClick={handleAddPanel} {...triggerButton}>
          {triggerButton.children}
        </Button>
      ) : (
        panel
      )}
    </>
  )
}
