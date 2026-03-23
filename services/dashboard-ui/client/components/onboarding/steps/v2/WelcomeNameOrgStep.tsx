import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

const ADJECTIVES = [
  'swift',
  'bold',
  'bright',
  'calm',
  'clear',
  'crisp',
  'keen',
  'light',
  'quick',
  'sharp',
]
const NOUNS = [
  'harbor',
  'peak',
  'river',
  'cloud',
  'ridge',
  'grove',
  'meadow',
  'stone',
  'creek',
  'birch',
]

function generateOrgName() {
  const adj = ADJECTIVES[Math.floor(Math.random() * ADJECTIVES.length)]
  const noun = NOUNS[Math.floor(Math.random() * NOUNS.length)]
  return `${adj}-${noun}`
}

export const WelcomeNameOrgStep = ({
  onAdvance,
  setSharedData,
  nextStepTitle,
}: IWizardStepComponentProps) => {
  const [orgName, setOrgName] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!orgName.trim()) return
    setSharedData('orgName', orgName.trim())
    onAdvance()
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <div className="flex flex-col gap-1 w-full md:max-w-[400px]">
        <Input
          id="orgName"
          name="orgName"
          placeholder="e.g. swift-harbor"
          value={orgName}
          onChange={(e) => setOrgName(e.target.value)}
          labelProps={{ labelText: 'Organization name' }}
        />
        <Button
          className="!px-1"
          type="button"
          variant="ghost"
          onClick={() => setOrgName(generateOrgName())}
        >
          <Icon variant="SparkleIcon" />
          Generate random name
        </Button>
      </div>
      <div className="flex justify-end w-full">
        <Button type="submit" variant="primary" disabled={!orgName.trim()}>
          {nextStepTitle ?? 'Continue'}{' '}
          <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>
    </form>
  )
}
