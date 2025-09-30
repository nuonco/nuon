'use client'

import React, { type FC, useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Modal } from '@/components/Modal'
import { Button } from '@/components/Button'
import { Text } from '@/components/Typography'
import { useAccount } from '@/components/AccountProvider'
import { ChecklistItem } from './ChecklistItem'
import { CLIInstallStepContent } from './CLIInstallStepContent'
import { CreateAppStepContent } from './CreateAppStepContent'
import { AppSyncStepContent } from './AppSyncStepContent'
import { InstallCreationStepContent } from './InstallCreationStepContent'
import { OrgCreationStepContent } from './OrgCreationStepContent'
import type { TAccount, TUserJourney, TUserJourneyStep } from '@/types'
import { completeUserJourney } from '@/components/org-actions'

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
        }, 2000) // Brief delay to show completion state
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
      await completeUserJourney('evaluation')
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
      <div className="space-y-2">
        <div className="space-y-3">
          {allJourneySteps.map((step, index) => (
            <ChecklistItem
              key={step.name}
              step={step}
              isExpanded={expandedItems.includes(step.name)}
              onToggleExpand={() => toggleExpand(step.name)}
            >
              {step.name === 'account_created' ? (
                <div className="space-y-6">
                  {/* Success Message - Shown when step is complete */}
                  {step.complete && (
                    <div className="space-y-3 pb-4 border-b border-gray-200 dark:border-gray-700">
                      <div className="flex items-center gap-2">
                        <div className="w-2 h-2 bg-green-500 rounded-full" />
                        <Text
                          variant="semi-14"
                          className="text-green-800 dark:text-green-200"
                        >
                          Your account has been created successfully!
                        </Text>
                      </div>
                      <Text className="text-gray-600 dark:text-gray-400">
                        You&rsquo;re now ready to set up your organization and start
                        deploying applications.
                      </Text>
                    </div>
                  )}

                  {/* Original Step Instructions - Always shown */}
                  <div className={`space-y-3 ${step.complete ? 'opacity-75' : ''}`}>
                    <Text className="text-gray-600 dark:text-gray-400">
                      Welcome to Nuon! Your account creation is the first step in setting up your deployment platform.
                    </Text>
                    <Text className="text-sm text-gray-500 dark:text-gray-500">
                      With your account created, you can now proceed to create an organization and start managing your applications.
                    </Text>
                  </div>
                </div>
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
                (() => {
                  // Get app_id from current step or fallback to app_created step
                  const appId = step.metadata?.app_id ||
                    allJourneySteps.find(s => s.name === 'app_created')?.metadata?.app_id

                  return (
                    <InstallCreationStepContent
                      stepComplete={step.complete}
                      onClose={handleSkipAll}
                      installId={step.metadata?.install_id}
                      appId={appId}
                      orgId={orgId}
                      onNavigateToInstall={handleNavigateToInstall}
                    />
                  )
                })()
              ) : null}
            </ChecklistItem>
          ))}
        </div>
      </div>
    </Modal>
  )
}
