import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { InstallWizard } from '@/components/installs/InstallWizard'
import { useInstall } from '@/hooks/use-install'
import { useInstallLink } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { deprovisionInstall, forgetInstall } from '@/lib'
import type { TAPIError } from '@/types'
import { ConfirmStep } from './ConfirmStep'
import { ReviewStep } from './ReviewStep'
import { TeardownStep } from './TeardownStep'
import { TEARDOWN_WIZARD_STEPS, teardownWizardStatuses } from './wizard-path'

const toApiError = (error: unknown, fallback: string): TAPIError => {
  const apiError = (error ?? {}) as Partial<TAPIError>
  return {
    error: apiError.error || fallback,
    description: apiError.description || '',
    user_error: !!apiError.user_error,
    status: apiError.status,
  }
}

export const TeardownInstallWizardContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const installLink = useInstallLink()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [farthest, setFarthest] = useState(0)
  const [showing, setShowing] = useState(0)
  const [workflowId, setWorkflowId] = useState<string>()
  const installHref = installLink({
    installId: install.id,
    appId: install.app_id,
  })

  const {
    mutate: startTeardown,
    isPending: isStarting,
    error: startError,
  } = useMutation({
    mutationFn: async () => {
      const result = await deprovisionInstall({
        orgId: org.id,
        installId: install.id,
        body: {
          plan_only: false,
          error_behavior: 'abort',
        },
      })
      if (!result.data.workflow_id) {
        throw new Error('Unable to track the teardown workflow.')
      }
      return result.data.workflow_id
    },
    onSuccess: (nextWorkflowId) => {
      setWorkflowId(nextWorkflowId)
      setFarthest(2)
      setShowing(2)
      queryClient.invalidateQueries({
        queryKey: ['install', org.id, install.id],
      })
      queryClient.invalidateQueries({ queryKey: ['active-workflows'] })
    },
  })

  const {
    mutate: forget,
    isPending: isForgetting,
    error: forgetError,
  } = useMutation({
    mutationFn: () =>
      forgetInstall({
        orgId: org.id,
        installId: install.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['installs'] })
      queryClient.invalidateQueries({ queryKey: ['app-installs'] })
      navigate(
        install.app_branch_id
          ? `/${org.id}/apps/${install.app_id}/branches/${install.app_branch_id}/installs`
          : `/${org.id}/apps/${install.app_id}/installs`
      )
    },
  })

  const step = TEARDOWN_WIZARD_STEPS[showing]
  const locked = farthest === TEARDOWN_WIZARD_STEPS.length - 1
  const advance = () => {
    const next = Math.min(showing + 1, TEARDOWN_WIZARD_STEPS.length - 1)
    setFarthest((current) => Math.max(current, next))
    setShowing(next)
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-6 pt-8 pb-8">
      <InstallWizard
        steps={TEARDOWN_WIZARD_STEPS.map((item, index) => ({
          id: item.id,
          label: item.navLabel,
          status: teardownWizardStatuses(farthest)[index],
        }))}
        selectedId={step.id}
        title={step.title}
        description={step.description}
        onSelect={(id) => {
          if (locked) return
          const index = TEARDOWN_WIZARD_STEPS.findIndex(
            (item) => item.id === id
          )
          if (index !== -1 && index <= farthest) setShowing(index)
        }}
      >
        {step.id === 'review' ? (
          <ReviewStep
            installName={install.name}
            installHref={installHref}
            onAdvance={advance}
          />
        ) : null}
        {step.id === 'confirm' ? (
          <ConfirmStep
            installName={install.name}
            installHref={installHref}
            isPending={isStarting}
            error={
              startError
                ? toApiError(startError, 'Unable to start teardown')
                : null
            }
            onGoBack={() => setShowing(0)}
            onSubmit={startTeardown}
          />
        ) : null}
        {step.id === 'teardown' && workflowId ? (
          <TeardownStep
            workflowId={workflowId}
            installName={install.name}
            installHref={installHref}
            isForgetting={isForgetting}
            forgetError={
              forgetError
                ? toApiError(forgetError, 'Unable to forget install')
                : null
            }
            onForget={forget}
          />
        ) : null}
      </InstallWizard>
    </div>
  )
}
