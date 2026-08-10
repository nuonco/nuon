import { Icon } from '@/components/common/Icon'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import {
  InstallForm,
  useInstallForm,
  type InstallFormValues,
} from '@/components/installs/forms/InstallForm'
import { InstallStatuses } from '@/components/installs/InstallStatuses'
import type { TApp, TInstall, TAPIError } from '@/types'

interface ICompletedInstallCard {
  install?: TInstall
  installId: string
  orgId: string
  isLoading: boolean
}

export const CompletedInstallCard = ({
  install,
  installId,
  orgId,
  isLoading,
}: ICompletedInstallCard) => {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton height="40px" width="100%" />
        <Skeleton height="40px" width="100%" />
        <Skeleton height="40px" width="100%" />
      </div>
    )
  }

  return (
    <Card>
      <div className="flex items-center justify-between">
        <div className="flex flex-col">
          <Text variant="body" weight="strong">
            {install?.name}
          </Text>
          <ID>{installId}</ID>
        </div>
      </div>

      <InstallStatuses install={install} />

      <Text variant="subtext">
        <Link href={`/${orgId}/installs/${installId}`}>
          View install <Icon variant="CaretRightIcon" />
        </Link>
      </Text>
    </Card>
  )
}

const AUTO_APPROVE_DESCRIPTION =
  'Automatically approve and apply all future changes without manual confirmation. Defaulted to on for a faster trial flow.'

const OnboardingInstallForm = ({
  app,
  isPending,
  submitError,
  onSubmit,
}: {
  app: TApp
  isPending: boolean
  submitError?: TAPIError | null
  onSubmit: (values: InstallFormValues) => Promise<unknown> | void
}) => {
  const platform =
    (app.runner_config?.app_runner_type as 'aws' | 'azure') ?? 'aws'

  const { form, canSubmit } = useInstallForm({
    mode: 'create',
    platform,
    inputConfig: app.input_config,
    defaultAutoApprove: true,
    onSubmit,
  })

  return (
    <div className="flex flex-col gap-6">
      <FormErrorBanner error={submitError} fallback="Unable to create install" />
      <InstallForm
        form={form}
        mode="create"
        platform={platform}
        inputConfig={app.input_config}
        autoApproveDescription={AUTO_APPROVE_DESCRIPTION}
      />
      <div className="flex justify-end">
        <Button
          variant="primary"
          disabled={!canSubmit || isPending}
          onClick={() => form.handleSubmit()}
        >
          {isPending ? 'Creating install...' : 'Create install'}
        </Button>
      </div>
    </div>
  )
}

interface ICreateInstallStepContent {
  app?: TApp
  isLoading: boolean
  appError?: TAPIError | null
  isPending: boolean
  submitError?: TAPIError | null
  onSubmit: (values: InstallFormValues) => Promise<unknown> | void
}

export const CreateInstallStepContent = ({
  app,
  isLoading,
  appError,
  isPending,
  submitError,
  onSubmit,
}: ICreateInstallStepContent) => {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton height="40px" width="100%" />
        <Skeleton height="40px" width="100%" />
        <Skeleton height="40px" width="100%" />
      </div>
    )
  }

  if (appError || !app) {
    return (
      <Banner theme="error">
        {appError?.error || 'Unable to load app configuration. Try again.'}
      </Banner>
    )
  }

  return (
    <OnboardingInstallForm
      app={app}
      isPending={isPending}
      submitError={submitError}
      onSubmit={onSubmit}
    />
  )
}
