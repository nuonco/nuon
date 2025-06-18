'use client'

import classNames from 'classnames'
import React, { type FC, useCallback, useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { X, Rocket } from '@phosphor-icons/react'
import {
  Button,
  HeadingGroup,
  Text,
  TransitionDiv,
  type IButton,
} from '@/stratus/components/common'
import './Modal.css'

export interface IModal
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'tabIndex'> {
  actions?: React.ReactNode
  heading?: React.ReactNode
  isOpen?: boolean
  primaryAction?: IButton
  trigger?: Omit<IButton, 'onClick'>
}

export const Modal: FC<IModal> = ({
  actions,
  children,
  className,
  heading,
  isOpen = false,
  primaryAction,
  trigger = {
    children: 'Open',
  },
  ...props
}) => {
  const [isModalOpen, setIsModalOpen] = useState(isOpen)
  const modalRef = useRef(null)

  const handleClose = () => {
    setIsModalOpen(false)
  }

  const onEscape = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      handleClose()
    }
  }, [])

  useEffect(() => {
    document.addEventListener('keydown', onEscape, false)
    return () => {
      document.removeEventListener('keydown', onEscape, false)
    }
  }, [])

  useEffect(() => {
    if (isModalOpen) {
      setTimeout(() => {
        modalRef?.current?.focus()
      }, 155)
    }
  }, [isModalOpen])

  return (
    <>
      <Button
        onClick={() => {
          setIsModalOpen(true)
        }}
        {...trigger}
      >
        {trigger?.children}
      </Button>
      {typeof window === 'undefined'
        ? null
        : createPortal(
            <TransitionDiv
              className={classNames('modal-wrapper', {})}
              isVisible={isModalOpen}
            >
              <div className="modal-overlay" onClick={handleClose} />
              <div
                className={classNames('modal', {
                  [`${className}`]: Boolean(className),
                })}
                tabIndex={-1}
                ref={modalRef}
                {...props}
              >
                <div className="px-6 py-4 border-b flex items-center justify-between">
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
                    <Button className="!p-2" onClick={handleClose}>
                      <X />
                    </Button>
                  </div>
                </div>
                {children}
                <div className="px-6 py-4 border-t flex items-center gap-4 justify-between">
                  <Button type="button" onClick={handleClose}>
                    Close
                  </Button>
                  {primaryAction ? <Button {...primaryAction} /> : null}
                </div>
              </div>
            </TransitionDiv>,
            document.getElementById('surface-root')
          )}
    </>
  )
}

export const ExampleModal: FC = () => {
  return (
    <Modal
      className="max-w-2xl"
      heading={
        <>
          <Text
            className="flex items-center gap-1.5"
            variant="h3"
            weight="strong"
          >
            <Rocket /> Heading title
          </Text>
          <Text variant="base" weight="strong" theme="muted">
            Heading subtext
          </Text>
        </>
      }
      primaryAction={{
        children: 'Primary action',
        onClick: () => {},
        variant: 'primary',
      }}
      trigger={{ children: 'Custom text' }}
    >
      <div className="flex flex-col gap-4 px-6 py-4">
        <Text>Hey modal content</Text>
      </div>
    </Modal>
  )
}
