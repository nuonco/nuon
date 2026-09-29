import { useForm, useStore } from '@tanstack/react-form'
import { Status } from '@/components/common/Status'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TAPIError, TOrg } from '@/types'
import { createOrgSchema, type CreateOrgValues } from './schema'

interface ICreateOrgStep {
  onAdvance: IWizardStepComponentProps['onAdvance']
  nextStepTitle: IWizardStepComponentProps['nextStepTitle']
  createdOrg: TOrg | null
  isPending: boolean
  error: TAPIError | null
  onCreateOrg: (name: string) => void
  onGenerateName: () => Promise<string>
}

export const CreateOrgStep = ({
  onAdvance,
  nextStepTitle,
  createdOrg,
  isPending,
  error,
  onCreateOrg,
  onGenerateName,
}: ICreateOrgStep) => {
  const form = useForm({
    defaultValues: { orgName: '' } as CreateOrgValues,
    validators: { onMount: createOrgSchema, onChange: createOrgSchema },
    onSubmit: ({ value }) => onCreateOrg(value.orgName),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  return (
    <div className="flex flex-col gap-6">
      {isPending && (
        <Card>
          <div className="flex items-center justify-between">
            <Skeleton height="14px" width="320px" />
            <Skeleton height="22px" width="60px" />
          </div>
          <div className="flex flex-col gap-2">
            <Skeleton height="12px" width="120px" />
            <Skeleton height="12px" width="100px" />
          </div>
        </Card>
      )}

      {!createdOrg && !isPending && (
        <form
          autoComplete="off"
          noValidate
          onSubmit={(e) => e.preventDefault()}
          className="flex flex-col gap-4"
        >
          <FormErrorBanner
            error={error}
            fallback="Unable to create organization. Try again."
          />
          <div className="flex flex-col gap-1">
            <form.Field name="orgName">
              {(field) => (
                <FormInput
                  field={field}
                  id="org-name"
                  placeholder="swift-harbor-ridge"
                  labelProps={{ labelText: 'Organization name' }}
                />
              )}
            </form.Field>
            <Button
              className="!px-1"
              type="button"
              variant="ghost"
              onClick={async () => {
                const name = await onGenerateName()
                form.setFieldValue('orgName', name)
              }}
            >
              <Icon variant="SparkleIcon" />
              Generate random name
            </Button>
          </div>
          <div className="flex justify-end">
            <Button
              type="button"
              variant="primary"
              disabled={!canSubmit}
              onClick={() => form.handleSubmit()}
            >
              Create organization
            </Button>
          </div>
        </form>
      )}

      {createdOrg && (
        <>
          <Card>
            <div className="flex flex-col gap-4">
              <div className="flex items-center justify-between">
                <Text variant="body" weight="strong">
                  Your organization has been created.
                </Text>
                <Status
                  status={createdOrg.status ?? 'active'}
                  variant="badge"
                />
              </div>
              <div className="flex flex-col">
                <div className="flex items-center gap-2">
                  <Text variant="subtext" theme="neutral">
                    Name:
                  </Text>
                  <Text variant="body" weight="strong">
                    {createdOrg.name}
                  </Text>
                </div>
                <div className="flex items-center gap-2">
                  <Text variant="subtext" theme="neutral">
                    ID:
                  </Text>
                  <ID>{createdOrg.id}</ID>
                </div>
              </div>
            </div>
          </Card>

          <div className="flex justify-end">
            <Button variant="primary" onClick={onAdvance}>
              {nextStepTitle ?? 'Continue'}{' '}
              <Icon variant="CaretRightIcon" weight="bold" />
            </Button>
          </div>
        </>
      )}
    </div>
  )
}

interface ICompletedOrgCard {
  org: TOrg | undefined
  orgId: string
  isLoading: boolean
  onAdvance: IWizardStepComponentProps['onAdvance']
  nextStepTitle: IWizardStepComponentProps['nextStepTitle']
}

export const CompletedOrgCard = ({
  org,
  orgId,
  isLoading,
  onAdvance,
  nextStepTitle,
}: ICompletedOrgCard) => {
  if (isLoading) {
    return (
      <Card>
        <div className="flex items-center justify-between">
          <Skeleton height="14px" width="320px" />
          <Skeleton height="22px" width="60px" />
        </div>
        <div className="flex flex-col gap-2">
          <Skeleton height="12px" width="120px" />
          <Skeleton height="12px" width="100px" />
        </div>
      </Card>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <div className="flex flex-col">
              <Text variant="body" weight="strong">
                Your organization has been created.
              </Text>
              {org?.status !== 'active' ? (
                <Text variant="subtext" as="p" className="max-w-md">
                  It may take a few minutes to fully provision, but you
                  don&apos;t have to wait for it. You can continue to the next
                  step while it finishes.
                </Text>
              ) : null}
            </div>
            <Status status={org?.status ?? 'unknown'} variant="badge" />
          </div>
          <div className="flex flex-col">
            <div className="flex items-center gap-2">
              <Text variant="subtext" theme="neutral">
                Name:
              </Text>
              <Text variant="body" weight="strong">
                {org?.name}
              </Text>
            </div>
            <div className="flex items-center gap-2">
              <Text variant="subtext" theme="neutral">
                ID:
              </Text>
              <ID>{orgId}</ID>
            </div>
          </div>
        </div>
      </Card>

      <div className="flex justify-end">
        <Button variant="primary" onClick={onAdvance}>
          {nextStepTitle ?? 'Continue'}{' '}
          <Icon variant="CaretRightIcon" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
