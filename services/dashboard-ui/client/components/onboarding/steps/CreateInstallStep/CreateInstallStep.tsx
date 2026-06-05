import { useRef } from 'react'
import { Icon } from '@/components/common/Icon'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { CreateInstallForm } from '@/components/installs/forms/CreateInstallForm'
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

interface ICreateInstallStepContent {
  app: TApp
  isLoading: boolean
  appError?: TAPIError | null
  isPending: boolean
  onSubmit: (formData: FormData) => Promise<any>
  onFormSubmitClick: () => void
  formRef: React.RefObject<HTMLFormElement | null>
}

export const CreateInstallStepContent = ({
  app,
  isLoading,
  appError,
  isPending,
  onSubmit,
  onFormSubmitClick,
  formRef,
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
        {appError?.error ||
          'Unable to load app configuration. Try again.'}
      </Banner>
    )
  }

  return (
    <div className="flex flex-col">
      <CreateInstallForm
        ref={formRef}
        appId={app.id}
        platform={
          (app.runner_config?.app_runner_type as 'aws' | 'azure') ?? 'aws'
        }
        inputConfig={app.input_config}
        onSubmit={(formData: FormData) => onSubmit(formData)}
        onCancel={() => {}}
      />

      <div className="flex justify-end">
        <Button
          variant="primary"
          disabled={isPending}
          onClick={onFormSubmitClick}
        >
          {isPending ? 'Creating install...' : 'Create install'}
        </Button>
      </div>
    </div>
  )
}
