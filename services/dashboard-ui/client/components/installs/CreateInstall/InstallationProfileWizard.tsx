import { useState, type FormEvent } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Input } from '@/components/common/form/Input'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Select } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import {
  createCustomerManagedInstall,
  getAppReleases,
  type TCustomerManagedInstall,
} from '@/lib'
import type { TApp } from '@/types'

type UpdateOwner = 'nuon' | 'customer'

export const InstallationProfileWizard = ({
  app,
  onBack,
  onUseNuon,
}: {
  app: TApp
  onBack?: () => void
  onUseNuon: () => void
}) => {
  const { org } = useOrg()
  const [owner, setOwner] = useState<UpdateOwner>('nuon')
  const [step, setStep] = useState<'profile' | 'connected'>('profile')
  const [result, setResult] = useState<TCustomerManagedInstall | null>(null)
  const [releaseId, setReleaseId] = useState('')
  const { data: releasesResult } = useQuery({
    queryKey: ['app-releases', org?.id, app.id, 'install-setup'],
    queryFn: () =>
      getAppReleases({ orgId: org!.id, appId: app.id, limit: 100 }),
    enabled: step === 'connected' && !!org?.id,
  })
  const releases = (releasesResult?.data ?? []).filter(
    (release) => release.status === 'ready'
  )
  const { mutateAsync, isPending, error } = useMutation({
    mutationFn: createCustomerManagedInstall,
    onSuccess: setResult,
  })

  const continueFromProfile = () => {
    if (owner === 'nuon') return onUseNuon()
    setStep('connected')
  }

  const submitInstall = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    void mutateAsync({
      orgId: org?.id || '',
      body: {
        app_id: app.id,
        intended_name: String(form.get('intended_name')),
        release_id: String(form.get('release_id')),
        telemetry: 'operational',
        aws_region: String(form.get('aws_region')),
        aws_account_id: String(form.get('aws_account_id')),
        inputs: {},
      },
    })
  }

  if (step === 'connected') {
    if (result) {
      return (
        <div className="flex flex-col gap-6">
          <Banner theme="success">
            <div className="flex flex-col gap-1">
              <Text weight="strong">Customer-managed install created</Text>
              <Text>
                Install {result.install.name} was created with a dedicated,
                install-scoped portal service account.
              </Text>
            </div>
          </Banner>
          <Text>
            Mint a standard token for service account{' '}
            <strong>{result.portal_service_account.id}</strong> and configure
            the portal with the API URL, org ID, install ID, and token.
          </Text>
        </div>
      )
    }

    return (
      <form className="flex flex-col gap-6" onSubmit={submitInstall}>
        <Button
          type="button"
          variant="ghost"
          className="w-fit"
          onClick={() => setStep('profile')}
        >
          Back
        </Button>
        <Banner theme="info">
          This creates the install, its normal provisioning workflow, and a
          dedicated install-scoped service account for the customer portal.
        </Banner>
        {error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Customer-managed install creation failed'}
          </Banner>
        ) : null}
        <Input
          id="intended-name"
          name="intended_name"
          required
          labelProps={{ labelText: 'Install name' }}
          placeholder="production"
        />
        <Input
          id="aws-region"
          name="aws_region"
          required
          labelProps={{ labelText: 'AWS region' }}
          placeholder="us-east-1"
        />
        <Input
          id="aws-account-id"
          name="aws_account_id"
          labelProps={{ labelText: 'AWS account ID' }}
          placeholder="123456789012"
        />
        <Select
          name="release_id"
          required
          value={releaseId}
          onChange={setReleaseId}
          labelProps={{ labelText: 'Release ID' }}
          options={releases.map((release) => ({
            value: release.id ?? '',
            label: `${release.id} · ${release.semantic_digest?.slice(0, 19)}…`,
          }))}
        />
        <Button
          type="submit"
          variant="primary"
          className="w-fit"
          disabled={isPending}
        >
          {isPending ? 'Creating install...' : 'Create install'}
        </Button>
      </form>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {onBack ? (
        <Button variant="ghost" className="w-fit" onClick={onBack}>
          Back
        </Button>
      ) : null}
      <fieldset className="flex flex-col gap-2">
        <Text as="legend" variant="h3" weight="strong">
          Who applies updates?
        </Text>
        <RadioInput
          name="update-owner"
          checked={owner === 'nuon'}
          onChange={() => setOwner('nuon')}
          labelProps={{ labelText: 'Nuon' }}
        />
        <RadioInput
          name="update-owner"
          checked={owner === 'customer'}
          onChange={() => setOwner('customer')}
          labelProps={{ labelText: 'Customer' }}
        />
      </fieldset>
      <Button variant="primary" className="w-fit" onClick={continueFromProfile}>
        Continue
      </Button>
    </div>
  )
}
