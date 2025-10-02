'use client'

import { useRouter, usePathname, useSearchParams } from 'next/navigation'
import { useState, useEffect } from 'react'
import { createPortal } from 'react-dom'
import { CubeIcon } from '@phosphor-icons/react'
import { createAppInstall } from '@/actions/apps/create-app-install'
import { Button } from '@/components/Button'
import { InstallForm } from '@/components/InstallForm'
import { Loading } from '@/components/Loading'
import { Modal } from '@/components/Modal'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useAccount } from '@/hooks/use-account'
import type { TAppConfig, TUserJourney } from '@/types'

interface IAppCreateInstallButton {
  platform: string | 'aws' | 'azure'
}

export const AppCreateInstallButton = ({
  platform,
}: IAppCreateInstallButton) => {
  const [isOpen, setIsOpen] = useState(false)
  const router = useRouter()
  const searchParams = useSearchParams()

  // Check for createInstall URL parameter and auto-open modal
  useEffect(() => {
    const shouldAutoOpen = searchParams.get('createInstall') === 'true'
    if (shouldAutoOpen) {
      setIsOpen(true)
      // Clean up URL parameter to avoid issues with refresh/back button
      const url = new URL(window.location.href)
      url.searchParams.delete('createInstall')
      router.replace(url.pathname + url.search, { scroll: false })
    }
  }, [searchParams, router])

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="!max-w-5xl"
              isOpen={isOpen}
              heading={
                <span className="flex flex-col gap-2">
                  <Text variant="med-18">Create install</Text>
                  <Text variant="reg-14" className="!font-normal">
                    Enter the following information to setup your install.
                  </Text>
                </span>
              }
              onClose={() => {
                setIsOpen(false)
              }}
              contentClassName="px-0 py-0"
            >
              <LoadAppConfigs
                platform={platform}
                onClose={() => {
                  setIsOpen(false)
                }}
              />
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="flex items-center gap-2 text-sm font-medium"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <CubeIcon size={16} /> Create install
      </Button>
    </>
  )
}

interface ICreateInstallFromAppConfig {
  onClose: () => void
  platform?: 'aws' | 'azure' | string
}

const LoadAppConfigs = (props: ICreateInstallFromAppConfig) => {
  const { org } = useOrg()
  const { app } = useApp()
  const {
    data: configs,
    isLoading,
    error,
  } = useQuery<TAppConfig[]>({
    path: `/api/orgs/${org?.id}/apps/${app?.id}/configs`,
  })
  return isLoading ? (
    <div className="p-6">
      <Loading loadingText="Loading configs..." variant="stack" />
    </div>
  ) : error?.error ? (
    <div className="p-6">
      <Notice>{error?.error || 'Unable to load app configs'}</Notice>
    </div>
  ) : (
    <CreateInstallFromAppConfig configId={configs?.at(0)?.id} {...props} />
  )
}

const CreateInstallFromAppConfig = ({
  onClose,
  platform,
  configId,
}: ICreateInstallFromAppConfig & { configId: string }) => {
  const path = usePathname()
  const router = useRouter()
  const { org } = useOrg()
  const { app } = useApp()
  const { account } = useAccount()

  // Check if this install creation will complete the evaluation journey
  const isCompletingEvaluationJourney = () => {
    const accountWithJourneys = account as any
    const evaluationJourney = accountWithJourneys?.user_journeys?.find(
      (journey: TUserJourney) => journey.name === 'evaluation'
    )

    if (!evaluationJourney) return false

    // Check if install_created step is the only incomplete step
    const incompleteSteps = evaluationJourney.steps.filter((step: any) => !step.complete)
    return incompleteSteps.length === 1 && incompleteSteps[0].name === 'install_created'
  }
  const {
    data: config,
    isLoading,
    error,
  } = useQuery<TAppConfig>({
    path: `/api/orgs/${org?.id}/apps/${app?.id}/configs/${configId}?recurse=true`,
  })

  return (
    <>
      {isLoading ? (
        <div className="p-6">
          <Loading loadingText="Loading configs..." variant="stack" />
        </div>
      ) : error?.error ? (
        <div className="p-6">
          <Notice>{error?.error || 'Unable to load app config.'}</Notice>
        </div>
      ) : (
        <InstallForm
          onSubmit={(formData) => {
            return createAppInstall({
              appId: app.id,
              orgId: org.id,
              formData,
              path,
            })
          }}
          onSuccess={({ data: install, error, headers, status }) => {
            if (!error && status === 201) {
              const workflowUrl = `/${org.id}/installs/${install?.id}/workflows/${headers?.['x-nuon-install-workflow-id']}`

              // Add completion parameter if this completes the evaluation journey
              const isCompletingJourney = isCompletingEvaluationJourney()
              const finalUrl = isCompletingJourney
                ? `${workflowUrl}?onboardingComplete=true`
                : workflowUrl

              router.push(finalUrl)
            }
          }}
          onCancel={onClose}
          platform={platform}
          inputConfig={{
            ...config.input,
            input_groups: nestInputsUnderGroups(
              config.input?.input_groups,
              config.input?.inputs
            ),
          }}
        />
      )}
    </>
  )
}

function nestInputsUnderGroups(
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
) {
  return groups.map((group) => ({
    ...group,
    app_inputs: inputs.filter((input) => input.group_id === group.id),
  }))
}
