'use client'

import { type FormEvent, useRef, forwardRef, useEffect } from 'react'
import { usePathname } from 'next/navigation'
import { updateInstall } from '@/actions/installs/update-install'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useFormPersistence } from '@/hooks/use-form-persistence'
import { useSurfaces } from '@/hooks/use-surfaces'
import { InputConfigFields } from './shared/InputConfigFields'
import { PlatformFields } from './shared/PlatformFields'
import { ResumeDraftModal } from './shared/ResumeDraftModal'
import type { IUpdateInstallForm } from './shared/types'

const UpdateInstallOptions = () => {
  return (
    <fieldset className="flex flex-col gap-4 border-t pt-6">
      <legend className="flex flex-col gap-0 mb-4 pr-6">
        <span className="text-lg font-semibold">Update install resources</span>
        <span className="text-sm font-normal">
          Reprovision sandbox and redeploy components after updating install
          settings
        </span>
      </legend>

      <div className="flex gap-6 justify-start">
        <RadioInput
          name="form-control:update"
          value="skip"
          defaultChecked
          labelProps={{
            labelText: 'Skip updating resources',
            className:
              'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-0 !w-fit',
          }}
        />
        <RadioInput
          name="form-control:update"
          value="update"
          labelProps={{
            labelText: 'Update all resources',
            className:
              'hover:!bg-transparent focus:!bg-transparent active:!bg-transparent !px-0 !py-0 !w-fit',
          }}
        />
      </div>
    </fieldset>
  )
}

export const UpdateInstallForm = forwardRef<
  HTMLFormElement,
  IUpdateInstallForm
>(
  (
    {
      install,
      platform,
      inputConfig,
      onSubmit,
      onSuccess,
      onCancel,
      onFormSubmit,
      onRegisterClearDraft,
    },
    ref
  ) => {
    const path = usePathname()
    const { org } = useOrg()
    const formRef = useRef<HTMLFormElement>(null)
    const draftShownRef = useRef(false)
    const { addModal, removeModal } = useSurfaces()

    const { data, error, headers, isLoading, execute } = useServerAction({
      action: updateInstall,
    })

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
    }, [
      hasDraft,
      draftTimestamp,
      restoreDraft,
      clearDraft,
      addModal,
      removeModal,
    ])

    useServerActionToast({
      data,
      error,
      errorContent: <Text>Unable to update install {install.name}.</Text>,
      errorHeading: 'Install update failed',
      onSuccess: onSuccess
        ? () => {
            clearDraft()
            const result = { data, headers }
            onSuccess(result)
          }
        : undefined,
      successContent: <Text>Install {install.name} updated successfully!</Text>,
      successHeading: 'Install updated',
    })

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
      e.preventDefault()

      if (onFormSubmit) {
        onFormSubmit()
        return
      }

      const formData = new FormData(e.currentTarget)

      if (onSubmit) {
        try {
          const result = await onSubmit(formData)
          onSuccess?.(result)
        } catch (err) {
          console.error('Form submission error:', err)
        }
      } else {
        // Convert FormData to the format expected by updateInstall
        const formDataObj = Object.fromEntries(formData)
        const inputs = Object.keys(formDataObj).reduce(
          (acc, key) => {
            if (key.includes('inputs:')) {
              let value: any = formDataObj[key]
              if (value === 'on' || value === 'off') {
                value = Boolean(value === 'on').toString()
              }
              acc[key.replace('inputs:', '')] = value
            }
            return acc
          },
          {} as Record<string, any>
        )

        let body: any = {
          inputs,
          metadata: install.metadata || {},
        }

        execute({
          installId: install.id,
          orgId: org.id,
          path,
          body,
        })
      }
      clearDraft()
    }

    // Expose submit method to parent
    const submitForm = () => {
      if (formRef.current) {
        formRef.current.requestSubmit()
      }
    }

    return (
      <form
        key={formKey}
        ref={(node) => {
          formRef.current = node
          if (typeof ref === 'function') {
            ref(node)
          } else if (ref) {
            ref.current = node
          }
        }}
        onSubmit={handleSubmit}
        className="flex flex-col gap-8 justify-between focus:outline-none relative"
      >
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to update install, please try again.'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-8 max-w-4xl pb-12">
          <div className="flex flex-col gap-2">
            <Text variant="h3" weight="strong">
              Update {install.name}
            </Text>
            <Text variant="body" theme="neutral">
              Modify the configuration for this install.
            </Text>
          </div>

          {inputConfig && (
            <InputConfigFields
              key={formKey}
              inputConfig={inputConfig}
              install={install}
              draftValues={
                draftValues && Object.keys(draftValues).length > 0
                  ? draftValues
                  : undefined
              }
            />
          )}

          <UpdateInstallOptions />
        </div>
      </form>
    )
  }
)

UpdateInstallForm.displayName = 'UpdateInstallForm'
