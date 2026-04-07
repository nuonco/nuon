import { useState, useEffect, useRef } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { OrgAvatar } from '@/components/orgs/OrgAvatar'
import type { TAPIError, TOrg } from '@/types'

const WAITING_MESSAGES = [
  'Creating your organization...',
  'Setting up your workspace...',
  'Almost there...',
]

function useProgressMessage(messages: string[], enabled: boolean, intervalMs = 3000) {
  const [index, setIndex] = useState(0)
  const timerRef = useRef<ReturnType<typeof setInterval>>()

  useEffect(() => {
    if (!enabled) {
      setIndex(0)
      return
    }
    timerRef.current = setInterval(() => {
      setIndex((prev) => Math.min(prev + 1, messages.length - 1))
    }, intervalMs)
    return () => clearInterval(timerRef.current)
  }, [enabled, messages.length, intervalMs])

  return messages[index]
}

type TOrgCardStatus = 'idle' | 'waiting' | 'success'

interface IOrgCardProps {
  name: string
  orgId?: string
  status?: TOrgCardStatus
  waitingMessage?: string
}

const ORG_RING_SIZE = 28
const ORG_RING_STROKE = 3

function OrgSpinner() {
  const radius = (ORG_RING_SIZE - ORG_RING_STROKE) / 2
  const circumference = 2 * Math.PI * radius
  const arcLen = circumference * 0.25

  return (
    <div className="relative flex items-center justify-center shrink-0" style={{ width: ORG_RING_SIZE, height: ORG_RING_SIZE }}>
      <svg width={ORG_RING_SIZE} height={ORG_RING_SIZE} className="-rotate-90">
        <circle
          cx={ORG_RING_SIZE / 2}
          cy={ORG_RING_SIZE / 2}
          r={radius}
          fill="none"
          strokeWidth={ORG_RING_STROKE}
          style={{ stroke: 'var(--border-color)' }}
        />
      </svg>
      <svg
        width={ORG_RING_SIZE}
        height={ORG_RING_SIZE}
        className="absolute inset-0"
        style={{ transformOrigin: 'center', animation: 'spinner-rotate 1s linear infinite' }}
      >
        <circle
          cx={ORG_RING_SIZE / 2}
          cy={ORG_RING_SIZE / 2}
          r={radius}
          fill="none"
          strokeWidth={ORG_RING_STROKE}
          strokeLinecap="round"
          strokeDasharray={`${arcLen} ${circumference - arcLen}`}
          style={{ stroke: 'var(--color-green-600)' }}
        />
      </svg>
    </div>
  )
}

function OrgCard({ name, orgId, status = 'idle', waitingMessage }: IOrgCardProps) {
  return (
    <Card className="!gap-0 !p-4">
      <div className="flex items-center gap-4">
        <OrgAvatar name={name} size="lg" />
        <div className="flex flex-col flex-1 min-w-0 gap-0">
          <Text variant="base" weight="strong" className="truncate">{name}</Text>
          {status === 'waiting' && waitingMessage && (
            <Text variant="body" className="text-cool-grey-600 dark:text-cool-grey-400">{waitingMessage}</Text>
          )}
          {status === 'success' && (
            <Text variant="body" className="text-green-700 dark:text-green-500">Organization created</Text>
          )}
        </div>
        {status === 'waiting' && <OrgSpinner />}
        {status === 'success' && (
          <Icon variant="CheckCircle" size={20} weight="fill" theme="success" className="shrink-0" />
        )}
      </div>
    </Card>
  )
}

export { OrgCard, useProgressMessage, WAITING_MESSAGES }

interface IWelcomeNameOrgStep {
  org?: TOrg
  orgName: string
  setOrgName: (name: string) => void
  isPending: boolean
  waiting: boolean
  error?: TAPIError | null
  stepError?: string
  displayName: string
  displayId?: string
  onSubmit: (e: React.FormEvent) => void
  onAdvance: () => void
  onGenerateName: () => void
}

export const WelcomeNameOrgStep = ({
  org,
  orgName,
  setOrgName,
  isPending,
  waiting,
  error,
  stepError,
  displayName,
  displayId,
  onSubmit,
  onAdvance,
  onGenerateName,
}: IWelcomeNameOrgStep) => {
  const isWorking = isPending || waiting
  const waitingMessage = useProgressMessage(WAITING_MESSAGES, waiting)

  if (org) {
    return (
      <div className="flex flex-col gap-8">
        <OrgCard name={org.name!} orgId={org.id} status="success" />
        <div className="flex justify-end w-full">
          <Button type="button" variant="primary" onClick={onAdvance}>
            Continue{' '}
            <Icon variant="CaretRight" weight="bold" />
          </Button>
        </div>
      </div>
    )
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-8">
      {error && (
        <Banner theme="error">
          {error.error ?? 'Failed to create organization'}
        </Banner>
      )}
      {stepError && (
        <Banner theme="error">{stepError}</Banner>
      )}
      {!isWorking && (
        <div className="flex flex-col gap-1 w-full md:max-w-[400px]">
          <Input
            id="orgName"
            name="orgName"
            placeholder="e.g. swift-harbor-ridge"
            value={orgName}
            onChange={(e) => setOrgName(e.target.value)}
            labelProps={{ labelText: 'Organization name' }}
            disabled={isWorking}
          />
          <Button
            className="!px-1"
            type="button"
            variant="ghost"
            onClick={onGenerateName}
            disabled={isWorking}
          >
            <Icon variant="SparkleIcon" />
            Generate random name
          </Button>
        </div>
      )}
      {isWorking && (
        <OrgCard
          name={displayName}
          orgId={displayId}
          status="waiting"
          waitingMessage={waitingMessage}
        />
      )}
      <div className="flex justify-end w-full">
        <Button type="submit" variant="primary" disabled={!orgName.trim() || isWorking}>
          {waiting ? 'Setting up org...' : isPending ? 'Creating...' : 'Continue'}{' '}
          {!isWorking && <Icon variant="CaretRight" weight="bold" />}
        </Button>
      </div>
    </form>
  )
}
