import { useState, useEffect, useRef } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { OrgAvatar } from '@/components/orgs/OrgAvatar'
import { cn } from '@/utils/classnames'
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
          <Icon variant="CheckCircleIcon" size={20} weight="fill" theme="success" className="shrink-0" />
        )}
      </div>
    </Card>
  )
}

export { OrgCard, useProgressMessage, WAITING_MESSAGES }

interface IExistingOrgCardProps {
  org: TOrg
  selected?: boolean
  disabled?: boolean
  pending?: boolean
  onSelect: () => void
}

function ExistingOrgCard({ org, selected, disabled, pending, onSelect }: IExistingOrgCardProps) {
  return (
    <button
      type="button"
      onClick={onSelect}
      disabled={disabled}
      className={cn(
        'flex items-center gap-4 p-4 rounded-md border text-left w-full transition-all',
        'bg-white dark:bg-dark-grey-900',
        selected
          ? '!border-primary-500 ring-2 ring-primary-500'
          : 'border-cool-grey-500/24 dark:border-cool-grey-500/24',
        disabled
          ? 'opacity-60 cursor-not-allowed'
          : !selected && 'hover:!border-primary-500 hover:ring-2 hover:ring-primary-500 cursor-pointer'
      )}
    >
      <OrgAvatar name={org.name!} size="lg" />
      <div className="flex flex-col flex-1 min-w-0 gap-0">
        <Text variant="base" weight="strong" className="truncate">{org.name}</Text>
        <Text variant="subtext" theme="neutral" className="truncate font-mono">{org.id}</Text>
      </div>
      {pending ? (
        <Icon variant="Loading" size={20} />
      ) : selected ? (
        <Icon variant="CheckCircleIcon" weight="fill" theme="success" size={20} />
      ) : (
        <Icon variant="CaretRightIcon" weight="bold" size={20} />
      )}
    </button>
  )
}

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
  existingOrgs?: TOrg[]
  isExistingOrgsLoading?: boolean
  attachingOrgId?: string | null
  isAttachPending?: boolean
  onSelectExistingOrg?: (orgId: string) => void
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
  existingOrgs,
  isExistingOrgsLoading,
  attachingOrgId,
  isAttachPending,
  onSelectExistingOrg,
  onSubmit,
  onAdvance,
  onGenerateName,
}: IWelcomeNameOrgStep) => {
  const isWorking = isPending || waiting || !!isAttachPending
  const waitingMessage = useProgressMessage(WAITING_MESSAGES, waiting)

  const attachedOrgId = org?.id
  const mergedOrgs: TOrg[] = (() => {
    const list = [...(existingOrgs ?? [])]
    if (org && !list.some((o) => o.id === org.id)) {
      list.unshift(org)
    }
    return list
  })()
  const hasExistingOrgs = mergedOrgs.length > 0

  // The user can explicitly switch to the create-new-org form via the
  // "+ Create a new organization" button. We default to `false` so that while
  // existing orgs are loading we don't render the create form, then jump back
  // to the list as soon as the data arrives.
  const [userClickedCreate, setUserClickedCreate] = useState(false)

  // If we've finished loading and the user truly has no orgs (and none is
  // attached to the onboarding session), there's nothing to pick from -- jump
  // straight into the create form so the wizard doesn't dead-end.
  const noOrgsAvailable = !isExistingOrgsLoading && !hasExistingOrgs
  const showCreateView = userClickedCreate || noOrgsAvailable

  const description = showCreateView
    ? 'Set up a fresh organization for this app.'
    : 'Pick an organization to create the app in, or create a new one.'

  const header = (
    <div className="mb-12">
      <Text variant="h2" role="heading" level={2} className="mb-2">
        Create your organization
      </Text>
      <Text variant="body" theme="neutral" as="p" className="max-w-md !text-pretty">
        {description}
      </Text>
    </div>
  )

  const canGoBackToList = userClickedCreate && hasExistingOrgs

  if (waiting || isPending) {
    return (
      <>
        {header}
        <div className="flex flex-col gap-8">
          {error && (
            <Banner theme="error">
              {error.error ?? 'Failed to create organization'}
            </Banner>
          )}
          {stepError && <Banner theme="error">{stepError}</Banner>}
          <OrgCard
            name={displayName}
            orgId={displayId}
            status="waiting"
            waitingMessage={waitingMessage}
          />
        </div>
      </>
    )
  }

  return (
    <>
      {header}
      <form onSubmit={onSubmit} className="flex flex-col gap-6">
        {error && (
          <Banner theme="error">
            {error.error ?? 'Failed to create organization'}
          </Banner>
        )}
        {stepError && (
          <Banner theme="error">{stepError}</Banner>
        )}

        {!showCreateView && (
          <>
            {isExistingOrgsLoading && !hasExistingOrgs ? (
              <div className="flex flex-col gap-2" aria-busy="true">
                {[0, 1].map((i) => (
                  <div
                    key={i}
                    className="h-[72px] rounded-md border border-cool-grey-500/24 dark:border-cool-grey-500/24 animate-pulse bg-cool-grey-100/50 dark:bg-dark-grey-800/50"
                  />
                ))}
              </div>
            ) : (
              <div className="flex flex-col gap-2">
                {mergedOrgs.map((o) => (
                  <ExistingOrgCard
                    key={o.id}
                    org={o}
                    selected={attachedOrgId === o.id}
                    disabled={isAttachPending}
                    pending={isAttachPending && attachingOrgId === o.id}
                    onSelect={() => {
                      if (!o.id) return
                      if (attachedOrgId === o.id) {
                        onAdvance()
                      } else {
                        onSelectExistingOrg?.(o.id)
                      }
                    }}
                  />
                ))}
              </div>
            )}
            <button
              type="button"
              onClick={() => setUserClickedCreate(true)}
              disabled={isWorking}
              className={cn(
                'flex items-center justify-center gap-2 p-4 rounded-md border border-dashed text-left w-full transition-all',
                'border-cool-grey-500/40 dark:border-cool-grey-500/40',
                isWorking
                  ? 'opacity-60 cursor-not-allowed'
                  : 'hover:!border-primary-500 hover:bg-primary-500/5 cursor-pointer'
              )}
            >
              <Icon variant="PlusIcon" weight="bold" size={16} />
              <Text weight="strong">Create a new organization</Text>
            </button>
          </>
        )}

        {showCreateView && (
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

        {showCreateView && (
          <div className="flex items-center justify-between w-full">
            {canGoBackToList ? (
              <Button
                type="button"
                variant="secondary"
                onClick={() => setUserClickedCreate(false)}
                disabled={isWorking}
              >
                <Icon variant="CaretLeftIcon" weight="bold" /> Back
              </Button>
            ) : (
              <div />
            )}
            <Button type="submit" variant="primary" disabled={!orgName.trim() || isWorking}>
              {isPending ? 'Creating...' : 'Continue'}{' '}
              {!isWorking && <Icon variant="CaretRightIcon" weight="bold" />}
            </Button>
          </div>
        )}
      </form>
    </>
  )
}
