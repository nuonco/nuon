'use client'

import { useState, useRef, useEffect } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TApp } from '@/types'
import { AppSelect } from './AppSelect'
import { LoadAppConfigs } from './LoadAppConfigs'

interface ICreateInstall {}

export const CreateInstallModal = ({ ...props }: ICreateInstall & IModal) => {
  const [selectedApp, setSelectedApp] = useState<TApp | undefined>()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)
  const modalRef = useRef<HTMLDivElement>(null)

  const handleClose = () => {
    setSelectedApp(undefined)
    props.onClose?.()
  }

  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }


  // For the app selection phase, we don't need any action buttons
  // The modal's close button (X) will handle cancellation
  const modalProps = selectedApp
    ? {
        // When showing install form, use primaryActionTrigger for submit button
        primaryActionTrigger: {
          children: isSubmitting ? (
            <span className="flex items-center gap-2">
              <Icon variant="Loading" />
              Creating install
            </span>
          ) : (
            <span className="flex items-center gap-2">
              <Icon variant="Cube" />
              Create install
            </span>
          ),
          disabled: isSubmitting,
          onClick: handleFormSubmit,
          variant: 'primary' as const,
        }
      }
    : {
        // When showing app selection, no primary action needed - selection happens via radio buttons
      }

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            className="inline-flex gap-4 items-center"
            variant="h3"
            weight="strong"
          >
            <Icon variant="Cube" size="24" />
            Create install
          </Text>
          {!selectedApp && (
            <Text
              variant="body"
              className="text-cool-grey-600 dark:text-cool-grey-400"
            >
              Select an app to create an install
            </Text>
          )}
        </div>
      }
      size={selectedApp ? "3/4" : "default"}
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      onClose={handleClose}
      {...modalProps}
      {...props}
    >
      {selectedApp ? (
        <LoadAppConfigs
          app={selectedApp}
          onSelectApp={setSelectedApp}
          onClose={handleClose}
          formRef={formRef}
          modalId={props.modalId}
          onLoadingChange={setIsSubmitting}
        />
      ) : (
        <AppSelect onSelectApp={setSelectedApp} onClose={handleClose} />
      )}
    </Modal>
  )
}
