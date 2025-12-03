'use client'

import { type FormEvent, useRef } from 'react'
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

export const CreateInstallForm = ({
  appId,
  platform,
  inputConfig,
  onSubmit,
  onSuccess,
  onCancel,
}: ICreateInstallForm) => {
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
    onSuccess: () => {
      const result = { data, headers }
      onSuccess(result)
    },
    successContent: <Text>Install created successfully!</Text>,
    successHeading: 'Install created',
  })

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    
    const formData = new FormData(e.currentTarget)
    
    if (onSubmit) {
      try {
        const result = await onSubmit(formData)
        onSuccess(result)
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
      ref={formRef}
      onSubmit={handleSubmit}
      className="flex flex-col gap-8 justify-between focus:outline-none relative"
    >
      {error ? (
        <Banner theme="error">
          {error?.error || 'Unable to create install, please try again.'}
        </Banner>
      ) : null}

      <div className="flex flex-col gap-8 max-w-3xl">
        <div className="flex flex-col gap-6">
          <Input
            id="install-name"
            name="name"
            placeholder="Enter install name"
            labelProps={{ labelText: 'Install name *' }}
            helperText="A unique name for this install"
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
}