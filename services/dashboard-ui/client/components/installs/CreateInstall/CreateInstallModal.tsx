import { useState, useRef, useEffect } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TApp } from '@/types'
import { AppSelectContainer as AppSelect } from './AppSelectContainer'
import { LoadAppConfigsContainer as LoadAppConfigs } from './LoadAppConfigsContainer'

interface ICreateInstall {}

export const CreateInstallModal = ({ ...props }: ICreateInstall & IModal) => {
  const [selectedApp, setSelectedApp] = useState<TApp | undefined>()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)
  const modalRef = useRef<HTMLDivElement>(null)
  const clearDraftRef = useRef<(() => void) | null>(null)

  const handleClose = () => {
    setSelectedApp(undefined)
    props.onClose?.()
  }

  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }


  const modalProps = selectedApp
    ? {
        primaryActionTrigger: {
          children: isSubmitting ? (
            <span className="flex items-center gap-2">
              <Icon variant="Loading" />
              Creating install
            </span>
          ) : (
            <span className="flex items-center gap-2">
              <Icon variant="CubeIcon" />
              Create install
            </span>
          ),
          disabled: isSubmitting,
          onClick: handleFormSubmit,
          variant: 'primary' as const,
        },
        secondaryActionTrigger: {
          children: 'Cancel',
          onClick: handleClose,
          variant: 'ghost' as const,
        },
      }
    : {}

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            flex
            className="gap-4"
            variant="h3"
            weight="strong"
          >
            <Icon variant="CubeIcon" size="24" />
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
      size={selectedApp ? 'xl' : 'default'}
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      onClose={handleClose}
      {...props}
      {...modalProps}
    >
      {selectedApp ? (
        <LoadAppConfigs
          app={selectedApp}
          onSelectApp={setSelectedApp}
          onClose={handleClose}
          formRef={formRef}
          modalId={props.modalId}
          onLoadingChange={setIsSubmitting}
          onRegisterClearDraft={(fn) => {
            clearDraftRef.current = fn
          }}
        />
      ) : (
        <AppSelect onSelectApp={setSelectedApp} onClose={handleClose} />
      )}
    </Modal>
  )
}
