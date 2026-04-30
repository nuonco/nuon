import { useNavigate } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type IButtonAsButton } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { type IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfigs, getAppConfig, createAppInstall, type TCreateAppInstallBody } from '@/lib'
import { toSentenceCase } from '@/utils/string-utils'
import { CreateInstallModal, CreateInstallButton as CreateInstallButtonComponent } from './CreateInstall'

interface ICreateInstall {}

const CreateInstallModalContainer = ({ ...props }: ICreateInstall & IModal) => {
  const { org } = useOrg()
  const { app } = useApp()
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

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
      } else if (platform === 'gcp') {
        body.gcp_account = {}
      }

      return createAppInstall({ appId: app?.id || '', body, orgId: org?.id || '' })
    },
    onSuccess: (result) => {
      addToast(
        <Toast heading="Install created" theme="success">
          <Text>Install created successfully!</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      removeModal(props.modalId)
      const workflowId = result.data.workflow_id
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

  const isLoading = configsLoading || configLoading
  const hasError =
    configsError || configError || !configs || configs.length === 0

  return (
    <CreateInstallModal
      isLoading={isLoading}
      hasError={!!hasError}
      configsError={configsError}
      configError={configError}
      config={config}
      configs={configs}
      isSubmitting={isSubmitting}
      appId={app.id}
      platform={app?.runner_config?.app_runner_type as 'aws' | 'azure' | 'gcp'}
      onSubmitAction={(formData) => mutateAsync(formData)}
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const CreateInstallButtonContainer = ({
  onClick: _onClick,
  ...props
}: ICreateInstall & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <CreateInstallModalContainer />

  return (
    <CreateInstallButtonComponent
      onClick={() => addModal(modal)}
      {...props}
    />
  )
}
