'use client'

import { useSearchParams } from 'next/navigation'
import React, { useEffect, useRef } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { useAutoFocusOnVisible } from '@/hooks/use-auto-focus-on-visible'
import { useEscapeKey } from '@/hooks/use-escape-key'
import { useSurfaces } from '@/hooks/use-surfaces'
import { cn } from '@/utils/classnames'
import './Modal.css'

export interface IModal
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'tabIndex'> {
  actions?: React.ReactNode
  childrenClassName?: string
  heading?: React.ReactNode
  isVisible?: boolean
  modalId?: string
  modalKey?: string
  onClose?: () => void
  primaryActionTrigger?: IButtonAsButton
  size?: 'default' | 'half' | 'full'
  triggerButton?: Omit<IButtonAsButton, 'onClick'>
}

export const ModalBase = ({
  actions,
  children,
  childrenClassName,
  className,
  heading,
  isVisible = false,
  modalId,
  modalKey,
  onClose,
  primaryActionTrigger,
  size = 'default',
  ...props
}: Omit<IModal, 'triggerButton'>) => {
  const { removeModal } = useSurfaces()
  const handleClose = () => {
    if (onClose) onClose?.()
    removeModal(modalId, modalKey)
  }
  const modalRef = useRef<HTMLDivElement>(null)
  // auto focus modal when in view
  useAutoFocusOnVisible(modalRef, isVisible)
  // handle close on esc key
  useEscapeKey(handleClose)

  return (
    <>
      <TransitionDiv
        className={cn(
          'modal-wrapper absolute top-0 left-0 w-screen h-screen flex z-10',
          {}
        )}
        isVisible={isVisible}
      >
        <div
          className="modal-overlay bg-black/50 dark:bg-black/75 absolute top-0 left-0 w-screen h-screen flex"
          onClick={handleClose}
        />
        <div
          className={cn(
            'modal bg-white dark:bg-dark-grey-900 border flex flex-col m-auto rounded-md shadow-lg',
            {
              'max-w-xl': size === 'default',
              'max-w-1/2': size === 'half',
            },
            className
          )}
          role="dialog"
          aria-modal="true"
          tabIndex={-1}
          ref={modalRef}
          {...props}
        >
          <div className="py-6 px-4 border-b flex items-center justify-between">
            {heading ? (
              typeof heading === 'string' ? (
                <HeadingGroup>
                  <Text variant="h3" weight="strong">
                    {heading}
                  </Text>
                </HeadingGroup>
              ) : (
                <HeadingGroup>{heading}</HeadingGroup>
              )
            ) : null}
            <div className="flex items-center gap-4">
              {actions}
              <Button
                className="!p-2"
                onClick={handleClose}
                title="Close modal"
                aria-label="Close modal"
                variant="ghost"
              >
                <Icon variant="X" />
              </Button>
            </div>
          </div>
          <div
            className={cn(
              'p-6 flex flex-col gap-4 md:gap-6',
              childrenClassName
            )}
          >
            {children}
          </div>
          <div className="py-6 px-4 border-t flex items-center gap-4 justify-end">
            <Button type="button" onClick={handleClose}>
              {primaryActionTrigger ? 'Cancel' : 'Close'}
            </Button>
            {primaryActionTrigger ? <Button {...primaryActionTrigger} /> : null}
          </div>
        </div>
      </TransitionDiv>
    </>
  )
}

export const Modal = ({ triggerButton, ...props }: IModal) => {
  const { addModal } = useSurfaces()
  const searchParams = useSearchParams()
  const modal = <ModalBase {...props} />

  const handleAddModal = () => {
    addModal(modal, props.modalKey)
  }

  useEffect(() => {
    if (
      props.modalKey &&
      props.modalKey === searchParams?.get('modal') &&
      !props.isVisible
    ) {
      handleAddModal()
    }
  }, [])

  return (
    <>
      {triggerButton ? (
        <Button onClick={handleAddModal} {...triggerButton}>
          {triggerButton.children}
        </Button>
      ) : (
        modal
      )}
    </>
  )
}
