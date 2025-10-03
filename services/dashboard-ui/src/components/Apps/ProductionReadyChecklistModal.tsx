'use client'

import React, { type FC, useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Check } from '@phosphor-icons/react'
import { Modal } from '@/components/Modal'
import { Button } from '@/components/Button'
import { Text } from '@/components/Typography'
import { useAccount } from '@/hooks/use-account'
import type { TAccount, TUserJourney, TUserJourneyStep } from '@/types'
import { ChecklistItem } from './ChecklistItem'
import { CreateAccountStepContent } from './CreateAccountStepContent'
import { CLIInstallStepContent } from './CLIInstallStepContent'
import { CreateAppStepContent } from './CreateAppStepContent'
import { AppSyncStepContent } from './AppSyncStepContent'
import { InstallCreationStepContent } from './InstallCreationStepContent'
import { OrgCreationStepContent } from './OrgCreationStepContent'
import { completeUserJourney } from '@/actions/accounts/complete-user-journey'

interface ProductionReadyChecklistModalProps {
  isOpen: boolean
  onClose: () => void
  account: TAccount | null
  orgId: string
  onForceClose?: () => void
}

// Journey helper functions
const getCurrentStep = (steps: TUserJourneyStep[]): TUserJourneyStep | null => {
  return steps.find((step) => !step.complete) || null
}

const getNextStep = (
  steps: TUserJourneyStep[],
  completedStepName: string
): TUserJourneyStep | null => {
  const completedIndex = steps.findIndex(
    (step) => step.name === completedStepName
  )
  const nextStep = steps[completedIndex + 1]
  return nextStep && !nextStep.complete ? nextStep : null
}

const detectNewlyCompletedStep = (
  oldJourney: TUserJourney,
  newJourney: TUserJourney
): TUserJourneyStep | null => {
  for (let i = 0; i < newJourney.steps.length; i++) {
    const oldStep = oldJourney.steps[i]
    const newStep = newJourney.steps[i]

    // Found a step that was incomplete but is now complete
    if (!oldStep?.complete && newStep?.complete) {
      return newStep
    }
  }
  return null
}

// Progress indicator component
const StepProgressIndicator: FC<{
  steps: TUserJourneyStep[]
  currentStepIndex: number
}> = ({ steps, currentStepIndex }) => {
  return (
    <div className="flex items-center justify-center space-x-2 mb-6 px-4">
      {steps.map((step, index) => (
        <React.Fragment key={step.name}>
          {/* Step circle */}
          <div
            className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-medium transition-all duration-300 ease-out transform ${
              step.complete
                ? 'bg-green-500 text-white scale-100'
                : index === currentStepIndex
                  ? 'bg-blue-500 text-white scale-110 animate-pulse'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-500 scale-90'
            }`}
          >
            {step.complete ? (
              <Check
                size={12}
                weight="bold"
                className="transition-transform duration-200"
              />
            ) : (
              <span></span>
            )}
          </div>

          {/* Connector line */}
          {index < steps.length - 1 && (
            <div
              className={`h-0.5 w-8 transition-all duration-300 ease-out ${
                step.complete ? 'bg-green-500' : 'bg-gray-200 dark:bg-gray-700'
              }`}
            />
          )}
        </React.Fragment>
      ))}
    </div>
  )
}

export const ProductionReadyChecklistModal: FC<
  ProductionReadyChecklistModalProps
> = ({ isOpen, onClose, account, orgId, onForceClose }) => {
  const { refreshAccount } = useAccount()
  const router = useRouter()
  const [expandedItems, setExpandedItems] = useState<string[]>([])
  const [previousJourneyState, setPreviousJourneyState] =
    useState<TUserJourney | null>(null)

  // Get evaluation journey
  const accountWithJourneys = account as any
  const evaluationJourney = accountWithJourneys?.user_journeys?.find(
    (journey: TUserJourney) => journey.name === 'evaluation'
  )

  // Show all journey steps including automatic ones for complete visibility
  const allJourneySteps = evaluationJourney?.steps || []
  const currentStepIndex = allJourneySteps.findIndex((step) => !step.complete)

  // Auto-expansion and step progression logic
  useEffect(() => {
    if (!evaluationJourney || allJourneySteps.length === 0) return

    // Initial load: auto-expand current step
    if (!previousJourneyState) {
      const currentStep = getCurrentStep(allJourneySteps)
      if (currentStep) {
        setExpandedItems([currentStep.name])
      }
      setPreviousJourneyState(evaluationJourney)
      return
    }

    // Step progression: detect completion and advance
    const completedStep = detectNewlyCompletedStep(
      previousJourneyState,
      evaluationJourney
    )

    if (
      completedStep &&
      allJourneySteps.some((step) => step.name === completedStep.name)
    ) {
      const nextStep = getNextStep(allJourneySteps, completedStep.name)

      // Close completed step, open next step (or close all if journey complete)
      setExpandedItems(nextStep ? [nextStep.name] : [])

      // Auto-close modal when all journey steps are complete
      if (!nextStep) {
        setTimeout(() => {
          onClose()
        }, 3000) // Extended delay to show completion celebration
      }
    }

    setPreviousJourneyState(evaluationJourney)
  }, [account, evaluationJourney, onClose, allJourneySteps])

  if (!evaluationJourney || allJourneySteps.length === 0) {
    return null
  }

  // Server action to complete evaluation journey
  const completeEvaluationJourneyAction = async (): Promise<boolean> => {
    try {
      await completeUserJourney({ journeyName: 'evaluation' })
      return true
    } catch (err) {
      // eslint-disable-next-line no-console
      console.error('Exception completing evaluation journey:', err)
      return false
    }
  }

  const handleSkipAll = async () => {
    const success = await completeEvaluationJourneyAction()
    if (success) {
      // Refresh account data so the provider detects all steps are complete
      // This will trigger the modal to close automatically via the journey logic
      await refreshAccount()
    } else {
      // Show user-visible error feedback if skip fails
      // eslint-disable-next-line no-console
      console.warn(
        'Skip failed - user may need to refresh the page or try again'
      )
    }
  }

  const toggleExpand = (stepName: string) => {
    setExpandedItems((prev) =>
      prev.includes(stepName)
        ? prev.filter((name) => name !== stepName)
        : [...prev, stepName]
    )
  }

  // Coordinated navigation: close modal → navigate → auto-open install modal
  const handleNavigateToInstall = (appId: string, orgId: string) => {
    // Step 1: Force close the onboarding modal immediately
    if (onForceClose) {
      // Use force close to bypass journey completion logic
      onForceClose()
    } else {
      // Fallback to regular close (may not work if journey incomplete)
      onClose()
    }

    // Step 2: Navigate after a brief delay to allow modal close animation
    setTimeout(() => {
      router.push(`/${orgId}/apps/${appId}?createInstall=true`)
    }, 300) // 300ms allows modal close transition to complete
  }

  const appId = allJourneySteps.find((s) => s.name === 'app_created')?.metadata
    ?.app_id

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => {}} // Disable overlay click to close
      heading="Get started"
      actions={
        <Button variant="secondary" onClick={handleSkipAll}>
          Skip
        </Button>
      }
      className="max-w-2xl"
      showCloseButton={false} // Remove X button
    >
      <div className="space-y-4">
        {/* Step Progress Indicator */}
        <StepProgressIndicator
          steps={allJourneySteps}
          currentStepIndex={
            currentStepIndex === -1 ? allJourneySteps.length : currentStepIndex
          }
        />

        <div className="space-y-3">
          {allJourneySteps.map((step, index) => (
            <ChecklistItem
              key={step.name}
              step={step}
              isExpanded={expandedItems.includes(step.name)}
              onToggleExpand={() => toggleExpand(step.name)}
            >
              {step.name === 'account_created' ? (
                <CreateAccountStepContent stepComplete={step.complete} />
              ) : step.name === 'org_created' ? (
                <OrgCreationStepContent stepComplete={step.complete} />
              ) : step.name === 'cli_installed' ? (
                <CLIInstallStepContent stepComplete={step.complete} />
              ) : step.name === 'app_created' ? (
                <CreateAppStepContent
                  stepComplete={step.complete}
                  appId={step.metadata?.app_id}
                />
              ) : step.name === 'app_synced' ? (
                <AppSyncStepContent
                  stepComplete={step.complete}
                  selectedAppPath={step.metadata?.app_path || 'eks-simple'}
                />
              ) : step.name === 'install_created' ? (
                <InstallCreationStepContent
                  stepComplete={step.complete}
                  onClose={handleSkipAll}
                  installId={step.metadata?.install_id}
                  appId={appId}
                  orgId={orgId}
                  onNavigateToInstall={handleNavigateToInstall}
                />
              ) : null}
            </ChecklistItem>
          ))}
        </div>
      </div>
    </Modal>
  )
}
