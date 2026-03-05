import { useRef, forwardRef } from 'react'
import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { CreateInstallForm } from '@/components/installs/forms/CreateInstallForm'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@tanstack/react-query'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfigs, getAppConfig, createAppInstall, type TCreateAppInstallBody } from '@/lib'
import type { TAppConfig } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

interface ICreateInstall {}

const FormSkeleton = () => {
  return (
    <div className="flex flex-col gap-8 max-w-4xl">
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
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const formRef = useRef<HTMLFormElement>(null)

  const {
    data: configs,
    isLoading: configsLoading,
    error: configsError,
  } = useQuery({
    queryKey: ['app-configs', org?.id, app?.id],
    queryFn: () => getAppConfigs({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const {
    data: config,
    isLoading: configLoading,
    error: configError,
  } = useQuery({
    queryKey: ['app-config', org?.id, app?.id, configs?.[0]?.id],
    queryFn: () => getAppConfig({ orgId: org.id, appId: app.id, appConfigId: configs[0].id, recurse: true }),
    enabled: !!configs?.[0]?.id,
  })

  const { mutate, isPending: isSubmitting, error: actionError } = useMutation({
    mutationFn: (formData: FormData) => {
      const formDataObj = Object.fromEntries(formData)
      const inputs = Object.keys(formDataObj).reduce(
        (acc, key) => {
          if (key.includes('inputs:')) {
            let value = formDataObj[key] as string
            if (value === 'on' || value === 'off') {
              value = (value === 'on').toString()
            }
            acc[key.replace('inputs:', '')] = value
          }
          return acc
        },
        {} as Record<string, string>
      )

      const body: TCreateAppInstallBody = {
        name: formDataObj.name as string,
        inputs: Object.keys(inputs).length > 0 ? inputs : undefined,
        install_config: {
          approval_option: formDataObj['auto-approve'] === 'on' ? 'approve-all' : 'prompt',
        },
        metadata: { managed_by: 'nuon/dashboard' },
      }

      const platform = app?.runner_config?.app_runner_type
      if (platform === 'aws' && formDataObj.region) {
        body.aws_account = { iam_role_arn: '', region: formDataObj.region as string }
      } else if (platform === 'azure' && formDataObj.location) {
        body.azure_account = {
          location: formDataObj.location as string,
          service_principal_app_id: '',
          service_principal_password: '',
          subscription_id: '',
          subscription_tenant_id: '',
        }
      } else if (platform === 'gcp' && formDataObj.gcp_region) {
        body.gcp_account = {
          region: formDataObj.gcp_region as string,
        }
      }

      return createAppInstall({ appId: app?.id || '', body, orgId: org?.id || '' })
    },
    onSuccess: (result) => {
      addToast(
        <Toast heading="Install created" theme="success">
          <Text>Install created successfully!</Text>
        </Toast>
      )
      removeModal(props.modalId)
      const workflowId = result?.headers?.['x-nuon-install-workflow-id']
      const suffix = result.data?.install_number === 1 ? '?onboardingComplete=true' : ''
      
      if (workflowId) {
        navigate(`/${org?.id}/installs/${result.data.id}/workflows/${workflowId}${suffix}`)
      } else {
        navigate(`/${org?.id}/installs/${result.data.id}/workflows${suffix}`)
      }
    },
    onError: (error) => {
      addToast(
        <Toast heading="Install creation failed" theme="error">
          <Text>
            {toSentenceCase(
              error.error || error.description || 'Unable to create install.'
            )}
          </Text>
        </Toast>
      )
    },
  })

  const handleFormSubmit = () => {
    if (formRef.current) {
      formRef.current.requestSubmit()
    }
  }

  const isLoading = configsLoading || configLoading
  const hasError =
    configsError || configError || !configs || configs.length === 0
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
        canSubmit
          ? {
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
            }
          : undefined
      }
      {...props}
    >
      {isLoading ? (
        <FormSkeleton />
      ) : hasError ? (
        <Banner theme="error">
          {configsError?.error ||
            configError?.error ||
            'Unable to load app configuration'}
        </Banner>
      ) : (
        <CreateInstallFormContent
          ref={formRef}
          configId={configs[0]?.id}
          config={config}
          onSubmitAction={(formData: FormData) => mutate(formData)}
          {...props}
        />
      )}
    </Modal>
  )
}

const CreateInstallFormContent = forwardRef<
  HTMLFormElement,
  {
    configId: string
    config: TAppConfig
    onSubmitAction: (formData: FormData) => void
  } & ICreateInstall &
    IModal
>(({ configId, config, onSubmitAction, ...props }, ref) => {
  const { app } = useApp()
  const { removeModal } = useSurfaces()

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

  return (
    <CreateInstallForm
      ref={ref}
      appId={app.id}
      platform={app?.runner_config?.app_runner_type as 'aws' | 'azure' | 'gcp'}
      inputConfig={{
        ...config?.input,
        input_groups: nestInputsUnderGroups(
          config?.input?.input_groups,
          config?.input?.inputs
        ),
      }}
      onSubmit={async (formData) => {
        onSubmitAction(formData)
      }}
      onCancel={() => {
        removeModal(props.modalId)
      }}
    />
  )
})

CreateInstallFormContent.displayName = 'CreateInstallFormContent'

export const CreateInstallButton = ({
  ...props
}: ICreateInstall & IButtonAsButton) => {
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
