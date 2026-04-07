import { useRef, useEffect, forwardRef } from 'react'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useFormPersistence } from '@/hooks/use-form-persistence'
import { ResumeDraftModal } from '../shared/ResumeDraftModal'
import { CreateInstallForm } from './CreateInstallForm'
import type { ICreateInstallForm } from '../shared/types'

export const CreateInstallFormContainer = forwardRef<HTMLFormElement, ICreateInstallForm>(
  ({ appId, inputConfig, onRegisterClearDraft, ...props }, ref) => {
    const formRef = useRef<HTMLFormElement>(null)
    const draftShownRef = useRef(false)
    const { addModal, removeModal } = useSurfaces()

    const {
      hasDraft,
      draftTimestamp,
      draftValues,
      clearDraft,
      restoreDraft,
      formKey,
    } = useFormPersistence({
      storageKey: `install-draft:${appId}`,
      formRef,
      enabled: true,
      configId: inputConfig?.id,
    })

    useEffect(() => {
      if (onRegisterClearDraft) {
        onRegisterClearDraft(clearDraft)
      }
    }, [onRegisterClearDraft, clearDraft])

    useEffect(() => {
      if (hasDraft && !draftShownRef.current && draftTimestamp) {
        draftShownRef.current = true

        let modalId: string
        const modal = (
          <ResumeDraftModal
            draftTimestamp={draftTimestamp}
            onResume={() => {
              restoreDraft()
              removeModal(modalId)
            }}
            onStartFresh={() => {
              clearDraft()
              draftShownRef.current = false
              removeModal(modalId)
            }}
            onClose={() => {
              removeModal(modalId)
            }}
          />
        )
        modalId = addModal(modal)
      }
    }, [hasDraft, draftTimestamp, restoreDraft, clearDraft, addModal, removeModal])

    return (
      <CreateInstallForm
        ref={(node) => {
          formRef.current = node
          if (typeof ref === 'function') {
            ref(node)
          } else if (ref) {
            ref.current = node
          }
        }}
        appId={appId}
        inputConfig={inputConfig}
        draftValues={draftValues}
        formKey={formKey}
        clearDraft={clearDraft}
        {...props}
      />
    )
  }
)

CreateInstallFormContainer.displayName = 'CreateInstallFormContainer'
