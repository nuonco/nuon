import { useNavigate } from 'react-router'
import { keepPreviousData, useMutation, useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import {
  buildCreateInstallBody,
  normalizeInstallPlatform,
  type InstallFormValues,
} from '@/components/installs/forms/InstallForm'
import {
  getApp,
  getInstall,
  createAppInstall,
  completeUserJourney,
} from '@/lib'
import { capturePostHogEvent } from '@/lib/posthog'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import { useToast } from '@/hooks/use-toast'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import { CompletedInstallCard, CreateInstallStepContent } from './CreateInstallStep'

export const CreateInstallStepContainer = ({
  onAdvance: _onAdvance,
}: IWizardStepComponentProps) => {
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
    placeholderData: keepPreviousData,
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
  const navigate = useNavigate()
  const { addToast } = useToast()

  const {
    data: app,
    isLoading,
    error: appError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['onboarding-app', orgId, appId],
    queryFn: () => getApp({ appId, orgId }),
  })

  const { mutateAsync, isPending, error: submitError } = useMutation({
    mutationFn: (values: InstallFormValues) =>
      createAppInstall({
        appId: app!.id,
        body: buildCreateInstallBody(
          values,
          normalizeInstallPlatform(app?.runner_config?.app_runner_type)
        ),
        orgId,
      }),
    onSuccess: async (result) => {
      capturePostHogEvent('install_created', {
        org_id: orgId,
        app_id: app?.id,
        install_id: result.data.id,
      })
      await completeUserJourney({ journeyName: 'evaluation' })
      addToast(
        <Toast heading="Install created" theme="success">
          <Text>Install created.</Text>
        </Toast>
      )
      const workflowId = result.data.workflow_id
      const suffix =
        result.data?.install_number === 1 ? '?onboardingComplete=true' : ''
      navigate(
        workflowId
          ? `/${orgId}/installs/${result.data.id}/workflows/${workflowId}${suffix}`
          : `/${orgId}/installs/${result.data.id}/workflows${suffix}`
      )
    },
  })

  return (
    <CreateInstallStepContent
      app={app}
      isLoading={isLoading}
      appError={appError}
      isPending={isPending}
      submitError={submitError as any}
      onSubmit={(values) => mutateAsync(values)}
    />
  )
}
