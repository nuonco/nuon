'use client'

import { type FormEvent, useRef, forwardRef, useEffect } from 'react'
import { usePathname } from 'next/navigation'
import { createAppInstall } from '@/actions/apps/create-app-install'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useFormPersistence } from '@/hooks/use-form-persistence'
import { useSurfaces } from '@/hooks/use-surfaces'
import { InputConfigFields } from './shared/InputConfigFields'
import { PlatformFields } from './shared/PlatformFields'
import { ResumeDraftModal } from './shared/ResumeDraftModal'
import type { ICreateInstallForm } from './shared/types'

export const CreateInstallForm = forwardRef<
  HTMLFormElement,
  ICreateInstallForm
>(
  (
    {
      appId,
      platform,
      inputConfig,
      onSubmit,
      onSuccess,
      onCancel,
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
      action: createAppInstall,
    })

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
      errorContent: <Text>Unable to create install.</Text>,
      errorHeading: 'Install creation failed',
      onSuccess: onSuccess
        ? () => {
            clearDraft()
            const result = { data, headers }
            onSuccess(result)
          }
        : undefined,
      successContent: <Text>Install created successfully!</Text>,
      successHeading: 'Install created',
    })

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
      e.preventDefault()

      const formData = new FormData(e.currentTarget)

      if (onSubmit) {
        try {
          const result = await onSubmit(formData)
          onSuccess?.(result)
        } catch (err) {
          console.error('Form submission error:', err)
        }
      } else {
        execute({
          appId,
          orgId: org.id,
          path,
          formData,
        })
      }

      clearDraft()
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
        className="flex flex-col min-h-[50vh] gap-8 justify-between focus:outline-none relative"
      >
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to create install, please try again.'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-8 max-w-4xl pb-12">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
            <span className="flex flex-col gap-0">
              <Text variant="body" weight="strong">
                Install name{' '}
                <Text className="ml-1" variant="subtext" theme="error">
                  {'*'}
                </Text>
              </Text>
              <Text variant="subtext" className="max-w-72">
                A unique name for this install
              </Text>
            </span>
            <Input
              id="install-name"
              name="name"
              placeholder="Enter install name"
              required
              defaultValue={draftValues?.name || ''}
            />
          </div>

          {platform && (
            <PlatformFields platform={platform} draftValues={draftValues} />
          )}

          {inputConfig && (
            <InputConfigFields
              inputConfig={inputConfig}
              draftValues={draftValues}
            />
          )}
        </div>
      </form>
    )
  }
)

CreateInstallForm.displayName = 'CreateInstallForm'
