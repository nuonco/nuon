'use client'

import { useRef, forwardRef } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { createAppInstall } from '@/actions/apps/create-app-install'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { CreateInstallForm } from '@/components/installs/forms/CreateInstallForm'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TAppConfig } from '@/types'

interface ICreateInstall {}

const FormSkeleton = () => {
  return (
    <div className="flex flex-col gap-8 max-w-3xl">
      {/* Install name section */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
        <span className="flex flex-col gap-1">
          <Skeleton width="100px" height="16px" />
          <Skeleton width="160px" height="14px" />
        </span>
        <Skeleton width="100%" height="40px" />
      </div>

      {/* AWS Settings */}
      <fieldset className="flex flex-col gap-6 border-t pt-6">
        <Skeleton width="140px" height="24px" />
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="130px" height="16px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
      </fieldset>

      {/* First input group */}
      <fieldset className="flex flex-col gap-6 border-t pt-6">
        <div className="flex flex-col gap-1 mb-6">
          <Skeleton width="220px" height="24px" />
          <Skeleton width="280px" height="16px" />
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="80px" height="16px" />
            <Skeleton width="200px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
      </fieldset>

      {/* Second input group */}
      <fieldset className="flex flex-col gap-6 border-t pt-6">
        <div className="flex flex-col gap-1 mb-6">
          <Skeleton width="180px" height="24px" />
          <Skeleton width="140px" height="16px" />
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="90px" height="16px" />
            <Skeleton width="160px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="110px" height="16px" />
            <Skeleton width="240px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
          <span className="flex flex-col gap-1">
            <Skeleton width="85px" height="16px" />
            <Skeleton width="300px" height="14px" />
          </span>
          <Skeleton width="100%" height="40px" />
        </div>
      </fieldset>
    </div>
  )
}

const CreateInstallModal = ({ ...props }: ICreateInstall & IModal) => {
  const { org } = useOrg()
  const { app } = useApp()
  const router = useRouter()
  const path = usePathname()
  const { removeModal } = useSurfaces()
  const formRef = useRef<HTMLFormElement>(null)

  const {
    data: configs,
    isLoading: configsLoading,
    error: configsError,
  } = useQuery<TAppConfig[]>({
    path: `/api/orgs/${org?.id}/apps/${app?.id}/configs`,
  })

  const {
    data: config,
    isLoading: configLoading,
    error: configError,
  } = useQuery<TAppConfig>({
    path: `/api/orgs/${org?.id}/apps/${app?.id}/configs/${configs?.[0]?.id}?recurse=true`,
    enabled: !!configs?.[0]?.id,
  })

  const { data: install, error: actionError, headers, isLoading: isSubmitting, execute } = useServerAction({
    action: createAppInstall,
  })

  useServerActionToast({
    data: install,
    error: actionError,
    errorContent: <Text>Unable to create install.</Text>,
    errorHeading: 'Install creation failed',
    onSuccess: () => {
      const workflowId = headers?.['x-nuon-install-workflow-id']
      if (install?.id && workflowId) {
        router.push(
          `/${org?.id}/installs/${install.id}/workflows/${workflowId}?onboardingComplete=true`
        )
      }
      removeModal(props.modalId)
    },
    successContent: <Text>Install created successfully!</Text>,
    successHeading: 'Install created',
  })

  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }

  const isLoading = configsLoading || configLoading
  const hasError = configsError || configError || !configs || configs.length === 0
  const canSubmit = !isLoading && !hasError && config

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
          <Icon variant="Cube" size="24" />
          Create install
        </Text>
      }
      size="3/4"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      primaryActionTrigger={
        canSubmit ? {
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
          variant: 'primary',
        } : undefined
      }
      {...props}
    >
      {isLoading ? (
        <FormSkeleton />
      ) : hasError ? (
        <Banner theme="error">
          {configsError?.error || configError?.error || 'Unable to load app configuration'}
        </Banner>
      ) : (
        <CreateInstallFormContent 
          ref={formRef}
          configId={configs[0]?.id} 
          config={config}
          onSubmitAction={(formData: FormData) => {
            return execute({
              appId: app?.id || '',
              orgId: org?.id || '',
              path,
              formData,
            })
          }}
          {...props} 
        />
      )}
    </Modal>
  )
}

const CreateInstallFormContent = forwardRef<HTMLFormElement, { 
  configId: string
  config: TAppConfig
  onSubmitAction: (formData: FormData) => Promise<any>
} & ICreateInstall & IModal>(({
  configId,
  config,
  onSubmitAction,
  ...props 
}, ref) => {
  const { org } = useOrg()
  const { app } = useApp()
  const { removeModal } = useSurfaces()

  const nestInputsUnderGroups = (
    groups: TAppConfig['input']['input_groups'],
    inputs: TAppConfig['input']['inputs']
  ) => {
    return groups
      ? groups.map((group) => ({
          ...group,
          app_inputs: inputs?.filter((input) => input.group_id === group.id) || [],
        }))
      : []
  }

  const getPlatform = (): 'aws' | 'azure' => {
    return 'aws'
  }

  return (
    <CreateInstallForm
      ref={ref}
      appId={app.id}
      platform={getPlatform()}
      inputConfig={{
        ...config?.input,
        input_groups: nestInputsUnderGroups(
          config?.input?.input_groups,
          config?.input?.inputs
        ),
      }}
      onSubmit={onSubmitAction}
      onCancel={() => {
        removeModal(props.modalId)
      }}
    />
  )
})

CreateInstallFormContent.displayName = 'CreateInstallFormContent'

export const CreateInstallButton = ({ ...props }: ICreateInstall & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <CreateInstallModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      <Icon variant="Cube" />
      Create install
    </Button>
  )
}
