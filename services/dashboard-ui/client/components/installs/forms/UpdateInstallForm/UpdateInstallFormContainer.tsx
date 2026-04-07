import { useRef, useEffect, forwardRef } from 'react'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useFormPersistence } from '@/hooks/use-form-persistence'
import { ResumeDraftModal } from '../shared/ResumeDraftModal'
import { UpdateInstallForm } from './UpdateInstallForm'
import type { IUpdateInstallForm } from '../shared/types'

export const UpdateInstallFormContainer = forwardRef<
  HTMLFormElement,
  IUpdateInstallForm
>(
  (props, ref) => {
    const { install, inputConfig, onRegisterClearDraft } = props
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
      storageKey: `install-update-draft:${install.id}`,
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
      <UpdateInstallForm
        {...props}
        ref={ref}
        draftValues={draftValues}
        formKey={formKey}
        clearDraft={clearDraft}
      />
    )
  }
)

UpdateInstallFormContainer.displayName = 'UpdateInstallFormContainer'
