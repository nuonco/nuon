import { useState, type FormEvent } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { Status } from '@/components/common/Status'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Input } from '@/components/common/form/Input'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { createOrg, getOrg, adminAddSupportUsersToOrg } from '@/lib'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TOrg } from '@/types'

export const CreateOrgStep = ({
  onAdvance,
  nextStepTitle,
  setSharedData,
  sharedData,
}: IWizardStepComponentProps) => {
  const [createdOrg, setCreatedOrg] = useState<TOrg | null>(null)
  const [orgName, setOrgName] = useState('')
  const { user } = useAuth()
  const { isByoc, sfTrialEndpoint, adminApiUrl } = useConfig()
  const { isStepComplete, getStepMetadata } = useOnboardingJourney()

  const orgCreated = isStepComplete('org_created')
  const journeyOrgId = getStepMetadata('org_created', 'org_id') as
    | string
    | undefined

  const { mutate, isPending, error } = useMutation({
    mutationFn: (name: string) =>
      createOrg({ body: { name, use_sandbox_mode: false, tags: ['Trial'] } }),
    onSuccess: (org) => {
      setCreatedOrg(org)
      setSharedData('orgId', org.id)

      if (!isByoc && adminApiUrl) {
        adminAddSupportUsersToOrg({
          orgId: org.id,
          adminApiUrl,
          adminEmail: user?.email ?? '',
        }).catch(() => {})
      }

      if (!isByoc && sfTrialEndpoint) {
        const nameParts = (user?.name ?? '').split(' ')
        const firstName = nameParts[0] ?? ''
        const lastName = nameParts.slice(1).join(' ') || 'ULN'
        fetch(sfTrialEndpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            firstName,
            lastName,
            email: user?.email,
            companyName: `${sharedData.companyName ?? ''} | ${org.name}`,
            jobTitle: sharedData.jobTitle,
            notes: sharedData.tellUsMore,
            subject: 'trial-signup',
          }),
        }).catch(() => {})
      }
    },
  })

  const { mutate: generateName } = useMutation({
    mutationFn: async () => {
      const res = await fetch('/api/random-name')
      const data = await res.json()
      return data.name as string
    },
    onSuccess: (name) => setOrgName(name),
  })

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    mutate(orgName)
  }

  if (orgCreated && journeyOrgId && !createdOrg) {
    return (
      <CompletedOrgCard
        orgId={journeyOrgId}
        onAdvance={onAdvance}
        nextStepTitle={nextStepTitle}
      />
    )
  }

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
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          {error && (
            <Banner theme="error">
              {error?.error ||
                'Failed to create organization. Please try again.'}
            </Banner>
          )}
          <div className="flex flex-col gap-1">
            <Input
              id="org-name"
              name="orgName"
              placeholder="e.g. swift-harbor-ridge"
              required
              value={orgName}
              onChange={(e) => setOrgName(e.target.value)}
              labelProps={{ labelText: 'Organization name' }}
            />
            <Button
              className="!px-1"
              type="button"
              variant="ghost"
              onClick={() => generateName()}
            >
              <Icon variant="SparkleIcon" />
              Generate random name
            </Button>
          </div>
          <div className="flex justify-end">
            <Button type="submit" variant="primary" disabled={!orgName.trim()}>
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
                  Your organization has been created successfully!
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
              <Icon variant="CaretRight" weight="bold" />
            </Button>
          </div>
        </>
      )}
    </div>
  )
}

function CompletedOrgCard({
  orgId,
  onAdvance,
  nextStepTitle,
}: {
  orgId: string
  onAdvance: IWizardStepComponentProps['onAdvance']
  nextStepTitle: IWizardStepComponentProps['nextStepTitle']
}) {
  const { data: org, isLoading } = useQuery({
    queryKey: ['onboarding-org', orgId],
    queryFn: () => getOrg({ orgId }),
    refetchInterval: 10000,
  })

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
                Your organization has been created successfully!
              </Text>
              {org?.status !== 'active' ? (
                <Text variant="subtext" className="!block max-w-md">
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
          <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>
    </div>
  )
}
