import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfig, createAppInstall, type TCreateAppInstallBody } from '@/lib'
import type { TApp } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { CreateInstallFromApp } from './CreateInstallFromApp'
import { BranchConnectionStep } from './BranchConnectionStep'

interface CreateInstallFromAppContainerProps {
  app: TApp
  configId: string
  onSelectApp: (app: TApp | undefined) => void
  onClose: () => void
  formRef?: React.RefObject<HTMLFormElement>
  modalId?: string
  onLoadingChange?: (loading: boolean) => void
  onRegisterClearDraft?: (clearFn: () => void) => void
  onInstallCreated?: () => void
}

export const CreateInstallFromAppContainer = ({
  app,
  configId,
  onSelectApp,
  onClose,
  formRef,
  modalId,
  onLoadingChange,
  onRegisterClearDraft,
  onInstallCreated,
}: CreateInstallFromAppContainerProps) => {
  const { org } = useOrg()
  const navigate = useNavigate()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [createdInstall, setCreatedInstall] = useState<{ id: string; workflowId?: string; suffix: string } | null>(null)

  const {
    data: config,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['app-config', org?.id, app.id, configId],
    queryFn: () => getAppConfig({ orgId: org.id, appId: app.id, appConfigId: configId, recurse: true }),
    enabled: !!org?.id,
  })

  const { mutateAsync, isPending: isSubmitting, error: submitError } = useMutation({
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

      const installConfig: TCreateAppInstallBody['install_config'] = {
        approval_option: formDataObj['auto-approve'] === 'on' ? 'approve-all' : 'prompt',
      }
      const vpcUrl = (formDataObj.vpc_nested_template_url as string)?.trim()
      const runnerUrl = (formDataObj.runner_nested_template_url as string)?.trim()
      if (vpcUrl) {
        installConfig!.vpc_nested_template_url = vpcUrl
      }
      if (runnerUrl) {
        installConfig!.runner_nested_template_url = runnerUrl
      }

      const labels: Record<string, string> = {}
      let labelIdx = 0
      while (formDataObj[`label:${labelIdx}:key`] !== undefined) {
        const key = (formDataObj[`label:${labelIdx}:key`] as string).trim()
        const value = (formDataObj[`label:${labelIdx}:value`] as string).trim()
        if (key) {
          labels[key] = value
        }
        labelIdx++
      }

      const body: TCreateAppInstallBody = {
        name: formDataObj.name as string,
        inputs: Object.keys(inputs).length > 0 ? inputs : undefined,
        install_config: installConfig,
        labels: Object.keys(labels).length > 0 ? labels : undefined,
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
          <Text>Install created.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['workflow-approvals'] })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
      const workflowId = result.data.workflow_id
      const suffix = result.data?.install_number === 1 ? '?onboardingComplete=true' : ''
      setCreatedInstall({ id: result.data.id, workflowId, suffix })
      onInstallCreated?.()
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

  const navigateToInstall = () => {
    removeModal(modalId)
    if (!createdInstall) return
    const { id, workflowId, suffix } = createdInstall
    if (workflowId) {
      navigate(`/${org?.id}/installs/${id}/workflows/${workflowId}${suffix}`)
    } else {
      navigate(`/${org?.id}/installs/${id}/workflows${suffix}`)
    }
  }

  if (createdInstall) {
    return (
      <BranchConnectionStep
        appId={app.id}
        installId={createdInstall.id}
        onDone={navigateToInstall}
      />
    )
  }

  return (
    <CreateInstallFromApp
      app={app}
      config={config}
      isLoading={isLoading}
      error={error}
      submitError={submitError}
      isSubmitting={isSubmitting}
      onSelectApp={onSelectApp}
      onClose={onClose}
      onSubmit={(formData) => mutateAsync(formData)}
      formRef={formRef}
      onRegisterClearDraft={onRegisterClearDraft}
    />
  )
}
