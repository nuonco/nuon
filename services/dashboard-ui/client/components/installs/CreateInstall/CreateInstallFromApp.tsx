import { useRef, useEffect } from 'react'
import { useNavigate } from 'react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { CreateInstallForm } from '@/components/installs/forms/CreateInstallForm'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfig, createAppInstall, type TCreateAppInstallBody } from '@/lib'
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
  onRegisterClearDraft?: (clearFn: () => void) => void
}

export const CreateInstallFromApp = ({
  app,
  configId,
  onSelectApp,
  onClose,
  formRef: externalFormRef,
  modalId,
  onLoadingChange,
  onRegisterClearDraft,
}: CreateInstallFromAppProps) => {
  const { org } = useOrg()
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const internalFormRef = useRef<HTMLFormElement>(null)
  const formRef = externalFormRef || internalFormRef

  const {
    data: config,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['app-config', org?.id, app.id, configId],
    queryFn: () => getAppConfig({ orgId: org.id, appId: app.id, appConfigId: configId, recurse: true }),
    enabled: !!org?.id,
  })

  const { mutateAsync, isPending: isSubmitting } = useMutation({
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

      const platform = app.runner_config?.app_runner_type
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
      } else if (platform === 'gcp') {
        body.gcp_account = {}
      }

      return createAppInstall({ appId: app.id, body, orgId: org?.id || '' })
    },
    onSuccess: (result) => {
      addToast(
        <Toast heading="Install created" theme="success">
          <Text>Install created successfully!</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      removeModal(modalId)
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

  useEffect(() => {
    onLoadingChange?.(isSubmitting)
  }, [isSubmitting, onLoadingChange])

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
            className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
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
            className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
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
          className="cursor-pointer w-fit text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded !bg-transparent !border-none !p-0 !h-auto font-medium"
          onClick={() => onSelectApp(undefined)}
        >
          <Icon variant="CaretLeft" weight="bold" />
          Back
        </Button>
      </div>

      <CreateInstallForm
        ref={formRef}
        appId={app.id}
        platform={app.runner_config.app_runner_type as 'aws' | 'azure' | 'gcp'}
        inputConfig={{
          ...config.input,
          input_groups: nestInputsUnderGroups(
            config.input?.input_groups,
            config.input?.inputs
          ),
        }}
        onSubmit={(formData: FormData) => mutateAsync(formData)}
        onCancel={onClose}
        onRegisterClearDraft={onRegisterClearDraft}
      />
    </div>
  )
}
