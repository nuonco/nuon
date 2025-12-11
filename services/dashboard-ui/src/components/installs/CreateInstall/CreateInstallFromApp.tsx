'use client'

import { useRef, useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { createAppInstall } from '@/actions/apps/create-app-install'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CreateInstallForm } from '@/components/installs/forms/CreateInstallForm'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TApp, TAppConfig } from '@/types'
import { FormSkeleton } from './FormSkeleton'
import { toSentenceCase } from '@/utils/string-utils'

interface CreateInstallFromAppProps {
  app: TApp
  configId: string
  onSelectApp: (app: TApp | undefined) => void
  onClose: () => void
  formRef?: React.RefObject<HTMLFormElement>
  modalId?: string
  onLoadingChange?: (loading: boolean) => void
}

export const CreateInstallFromApp = ({
  app,
  configId,
  onSelectApp,
  onClose,
  formRef: externalFormRef,
  modalId,
  onLoadingChange,
}: CreateInstallFromAppProps) => {
  const { org } = useOrg()
  const router = useRouter()
  const path = usePathname()
  const { removeModal } = useSurfaces()
  const internalFormRef = useRef<HTMLFormElement>(null)
  const formRef = externalFormRef || internalFormRef

  const {
    data: config,
    isLoading,
    error,
  } = useQuery<TAppConfig>({
    path: `/api/orgs/${org?.id}/apps/${app.id}/configs/${configId}?recurse=true`,
  })

  const {
    data: install,
    error: actionError,
    headers,
    isLoading: isSubmitting,
    execute,
  } = useServerAction({
    action: createAppInstall,
  })

  useEffect(() => {
    onLoadingChange?.(isSubmitting)
  }, [isSubmitting, onLoadingChange])

  useServerActionToast({
    data: install,
    error: actionError,
    errorContent: (
      <Text>
        {toSentenceCase(
          actionError?.error ||
            actionError?.description ||
            'Unable to create install.'
        )}
      </Text>
    ),
    errorHeading: 'Install creation failed',
    onSuccess: () => {
      const workflowId = headers?.['x-nuon-install-workflow-id']
      if (install?.id && workflowId) {
        router.push(
          `/${org?.id}/installs/${install.id}/workflows/${workflowId}?onboardingComplete=true`
        )
      }
      removeModal(modalId)
    },
    successContent: <Text>Install created successfully!</Text>,
    successHeading: 'Install created',
  })

  const nestInputsUnderGroups = (
    groups: TAppConfig['input']['input_groups'],
    inputs: TAppConfig['input']['inputs']
  ) => {
    return groups
      ? groups.map((group) => ({
          ...group,
          app_inputs:
            inputs?.filter((input) => input.group_id === group.id) || [],
        }))
      : []
  }

  if (isLoading) {
    return (
      <div>
        <div className="pb-4">
          <Button
            className="!flex items-center gap-1.5 cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeft" weight="bold" />
            Back
          </Button>
        </div>
        <FormSkeleton />
      </div>
    )
  }

  if (error || !config) {
    return (
      <div>
        <div className="pb-4">
          <Button
            className="!flex items-center gap-1.5 cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
            onClick={() => onSelectApp(undefined)}
          >
            <Icon variant="CaretLeft" weight="bold" />
            Back
          </Button>
        </div>
        <Banner theme="error">
          {error?.error || 'Unable to load app configuration'}
        </Banner>
      </div>
    )
  }

  return (
    <div>
      <div className="pb-4">
        <Button
          className="!flex items-center gap-1.5 cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
          onClick={() => onSelectApp(undefined)}
        >
          <Icon variant="CaretLeft" weight="bold" />
          Back
        </Button>
      </div>

      <CreateInstallForm
        ref={formRef}
        appId={app.id}
        platform={app.runner_config.app_runner_type as 'aws' | 'azure'}
        inputConfig={{
          ...config.input,
          input_groups: nestInputsUnderGroups(
            config.input?.input_groups,
            config.input?.inputs
          ),
        }}
        onSubmit={(formData: FormData) => {
          return execute({
            appId: app.id,
            orgId: org?.id || '',
            path,
            formData,
          })
        }}
        onCancel={onClose}
      />
    </div>
  )
}
