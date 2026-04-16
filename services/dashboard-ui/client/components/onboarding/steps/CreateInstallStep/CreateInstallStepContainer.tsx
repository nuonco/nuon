import { useRef } from 'react'
import { useNavigate } from 'react-router'
import { useQuery, useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import {
  getApp,
  getInstall,
  createAppInstall,
  completeUserJourney,
  type TCreateAppInstallBody,
} from '@/lib'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import { useToast } from '@/hooks/use-toast'
import { toSentenceCase } from '@/utils/string-utils'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import { CompletedInstallCard, CreateInstallStepContent } from './CreateInstallStep'

export const CreateInstallStepContainer = ({ onAdvance }: IWizardStepComponentProps) => {
  const { isStepComplete, getStepMetadata, orgId } = useOnboardingJourney()

  const installCreated = isStepComplete('install_created')
  const installId = getStepMetadata('install_created', 'install_id') as
    | string
    | undefined
  const appSynced = isStepComplete('app_synced')
  const appId = getStepMetadata('app_synced', 'app_id') as string | undefined

  if (installCreated && installId && orgId) {
    return <CompletedInstallCardContainer installId={installId} orgId={orgId} />
  }

  if (!appSynced || !appId || !orgId) {
    return (
      <div className="flex flex-col gap-4">
        <Text variant="body" theme="neutral">
          Waiting for app sync... Complete step 5 first.
        </Text>
      </div>
    )
  }

  return <CreateInstallStepContentContainer appId={appId} orgId={orgId} />
}

function CompletedInstallCardContainer({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) {
  const { data: install, isLoading } = useQuery({
    queryKey: ['onboarding-install', orgId, installId],
    queryFn: () => getInstall({ installId, orgId }),
    refetchInterval: 10000,
  })

  return (
    <CompletedInstallCard
      install={install}
      installId={installId}
      orgId={orgId}
      isLoading={isLoading}
    />
  )
}

function CreateInstallStepContentContainer({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) {
  const formRef = useRef<HTMLFormElement>(null)
  const navigate = useNavigate()
  const { addToast } = useToast()

  const {
    data: app,
    isLoading,
    error: appError,
  } = useQuery({
    queryKey: ['onboarding-app', orgId, appId],
    queryFn: () => getApp({ appId, orgId }),
  })

  const { mutateAsync, isPending } = useMutation({
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
          approval_option:
            formDataObj['auto-approve'] === 'on' ? 'approve-all' : 'prompt',
        },
        metadata: { managed_by: 'nuon/dashboard' },
      }

      const platform = app?.runner_config?.app_runner_type
      if (platform === 'aws' && formDataObj.region) {
        body.aws_account = {
          iam_role_arn: '',
          region: formDataObj.region as string,
        }
      } else if (platform === 'azure' && formDataObj.location) {
        body.azure_account = {
          location: formDataObj.location as string,
          service_principal_app_id: '',
          service_principal_password: '',
          subscription_id: '',
          subscription_tenant_id: '',
        }
      }

      return createAppInstall({ appId: app!.id, body, orgId })
    },
    onSuccess: async (result) => {
      await completeUserJourney({ journeyName: 'evaluation' })
      addToast(
        <Toast heading="Install created" theme="success">
          <Text>Install created successfully!</Text>
        </Toast>
      )
      const workflowId = result.data.workflow_id
      const suffix =
        result.data?.install_number === 1 ? '?onboardingComplete=true' : ''
      if (workflowId) {
        navigate(
          `/${orgId}/installs/${result.data.id}/workflows/${workflowId}${suffix}`
        )
      } else {
        navigate(`/${orgId}/installs/${result.data.id}/workflows${suffix}`)
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

  return (
    <CreateInstallStepContent
      app={app}
      isLoading={isLoading}
      appError={appError}
      isPending={isPending}
      onSubmit={(formData) => mutateAsync(formData)}
      onFormSubmitClick={() => formRef.current?.requestSubmit()}
      formRef={formRef}
    />
  )
}
