'use client'

import { type FormEvent, useRef, forwardRef } from 'react'
import { usePathname } from 'next/navigation'
import { createAppInstall } from '@/actions/apps/create-app-install'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { InputConfigFields } from './shared/InputConfigFields'
import { PlatformFields } from './shared/PlatformFields'
import type { ICreateInstallForm } from './shared/types'

export const CreateInstallForm = forwardRef<HTMLFormElement, ICreateInstallForm>(({
  appId,
  platform,
  inputConfig,
  onSubmit,
  onSuccess,
  onCancel,
}, ref) => {
  const path = usePathname()
  const { org } = useOrg()
  const formRef = useRef<HTMLFormElement>(null)

  const { data, error, headers, isLoading, execute } = useServerAction({
    action: createAppInstall,
  })

  useServerActionToast({
    data,
    error,
    errorContent: <Text>Unable to create install.</Text>,
    errorHeading: 'Install creation failed',
    onSuccess: onSuccess ? () => {
      const result = { data, headers }
      onSuccess(result)
    } : undefined,
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
  }

  return (
    <form
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
          {error?.error || 'Unable to create install, please try again.'}
        </Banner>
      ) : null}

      <div className="flex flex-col gap-8 max-w-3xl">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-0">
            <Text variant="body" weight="strong">Install name *</Text>
            <Text variant="subtext" className="max-w-72">
              A unique name for this install
            </Text>
          </span>
          <Input
            id="install-name"
            name="name"
            placeholder="Enter install name"
            required
          />
        </div>

        {platform && <PlatformFields platform={platform} />}
        
        {inputConfig && (
          <InputConfigFields 
            inputConfig={inputConfig} 
          />
        )}
      </div>

    </form>
  )
})

CreateInstallForm.displayName = 'CreateInstallForm'