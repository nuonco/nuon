import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Banner } from '@/components/common/Banner'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { completeOrganizationStep, getOrg } from '@/lib'
import { useOnboardingPoll } from '@/hooks/use-onboarding-poll'
import type { TAPIError, TOnboarding } from '@/types'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

const fetchRandomName = async () => {
  const res = await fetch('/api/random-name')
  const data = await res.json()
  return data.name as string
}

export const WelcomeNameOrgStep = ({
  onAdvance,
  sharedData,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const [orgName, setOrgName] = useState('')
  const [waiting, setWaiting] = useState(false)

  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id

  const { data: org } = useQuery({
    queryKey: ['org', orgId],
    queryFn: () => getOrg({ orgId: orgId! }),
    enabled: !!orgId,
  })

  const { mutate: generateName } = useMutation({
    mutationFn: fetchRandomName,
    onSuccess: (name) => setOrgName(name),
  })

  const { mutate: submit, isPending, error } = useMutation({
    mutationFn: () => completeOrganizationStep({ body: { name: orgName.trim() } }),
    onSuccess: (ob) => {
      setSharedData('onboarding', ob)
      if (ob.status_v2?.status === 'processing') {
        setWaiting(true)
      } else {
        onAdvance()
      }
    },
  })

  useOnboardingPoll({
    enabled: waiting,
    onResolved: (ob) => {
      setWaiting(false)
      setSharedData('onboarding', ob)
      if (ob.status_v2?.status === 'error') return
      onAdvance()
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!orgName.trim() || isPending || waiting) return
    submit()
  }

  const isWorking = isPending || waiting

  if (org) {
    return (
      <div className="flex flex-col gap-4">
        <Card className="flex items-center gap-4 p-4 w-full md:max-w-[400px]">
          <Icon variant="Buildings" size={24} />
          <div className="flex flex-col flex-1 min-w-0">
            <Text weight="strong">{org.name}</Text>
            <Text variant="subtext" theme="neutral">
              {org.sandbox_mode ? 'Sandbox' : 'Organization'} &middot; {org.status}
            </Text>
          </div>
          <Badge theme={org.status === 'active' ? 'success' : 'neutral'} size="sm">
            {org.status}
          </Badge>
        </Card>
        <div className="flex justify-end w-full">
          <Button type="button" variant="primary" onClick={onAdvance}>
            {nextStepTitle ?? 'Continue'}{' '}
            <Icon variant="CaretRight" weight="bold" />
          </Button>
        </div>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      {error && (
        <Banner theme="error">
          {(error as TAPIError).error ?? 'Failed to create organization'}
        </Banner>
      )}
      {onboarding?.status_v2?.status === 'error' && onboarding?.step_error && (
        <Banner theme="error">{onboarding.step_error}</Banner>
      )}
      <div className="flex flex-col gap-1 w-full md:max-w-[400px]">
        <Input
          id="orgName"
          name="orgName"
          placeholder="e.g. swift-harbor-ridge"
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
      <div className="flex justify-end w-full">
        <Button type="submit" variant="primary" disabled={!orgName.trim() || isWorking}>
          {waiting ? 'Setting up org...' : isPending ? 'Creating...' : (nextStepTitle ?? 'Continue')}{' '}
          {!isWorking && <Icon variant="CaretRight" weight="bold" />}
        </Button>
      </div>
    </form>
  )
}
