'use client'

import React, { type FC, useState, useEffect } from 'react'
import { Modal } from '@/components/Modal'
import { Button } from '@/components/Button'
import { Text } from '@/components/Typography'
import { useAccount } from '@/components/AccountProvider'
import { ChecklistItem } from './ChecklistItem'
import { CLIInstallStepContent } from './CLIInstallStepContent'
import { CreateAppStepContent } from './CreateAppStepContent'
import { InstallCreationStepContent } from './InstallCreationStepContent'
import { OrgCreationStepContent } from './OrgCreationStepContent'
import type { TAccount, TUserJourney, TUserJourneyStep } from '@/types'
import { completeUserJourney } from '@/components/org-actions'

interface ProductionReadyChecklistModalProps {
  isOpen: boolean
  onClose: () => void
  account: TAccount | null
  orgId: string
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

const getStepInstructions = (stepName: string) => {
  switch (stepName) {
    case 'account_created':
      return (
        <div className="space-y-3">
          <Text className="text-green-600 dark:text-green-400">
            ✅ Your account has been created successfully!
          </Text>
          <Text className="text-sm text-gray-600 dark:text-gray-400">
            You&rsquo;re now ready to set up your organization and start deploying applications.
          </Text>
        </div>
      )
    case 'org_created':
      return null // Will be handled by OrgCreationStepContent component
    case 'cli_installed':
      return null // Will be handled by CLIInstallStepContent component
    case 'app_created':
      return null // Will be handled by CreateAppStepContent component
    case 'app_synced':
      return (
        <div className="space-y-3">
          <Text>
            Navigate to the app config directory and sync your app configuration
            to make it available for deployment.
          </Text>
          <div className="bg-gray-100 dark:bg-gray-800 p-3 rounded font-mono text-sm">
            cd my-app
          </div>
          <div className="bg-gray-100 dark:bg-gray-800 p-3 rounded font-mono text-sm">
            nuon apps sync
          </div>
        </div>
      )
    case 'install_created':
      return null // Will be handled by InstallCreationStepContent component
    default:
      return null
  }
}

export const ProductionReadyChecklistModal: FC<
  ProductionReadyChecklistModalProps
> = ({ isOpen, onClose, account, orgId }) => {
  const { refreshAccount } = useAccount()
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
      console.warn('Skip All failed - user may need to refresh the page or try again')
    }
  }

  const toggleExpand = (stepName: string) => {
    setExpandedItems((prev) =>
      prev.includes(stepName)
        ? prev.filter((name) => name !== stepName)
        : [...prev, stepName]
    )
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => {}} // Disable overlay click to close
      heading="Get started"
      actions={
        <Button variant="secondary" onClick={handleSkipAll}>
          Skip All
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
              {step.name === 'org_created' ? (
                <OrgCreationStepContent stepComplete={step.complete} />
              ) : step.name === 'cli_installed' ? (
                <CLIInstallStepContent stepComplete={step.complete} />
              ) : step.name === 'app_created' ? (
                <CreateAppStepContent
                  stepComplete={step.complete}
                  appId={step.metadata?.app_id}
                />
              ) : step.name === 'install_created' ? (
                <InstallCreationStepContent
                  stepComplete={step.complete}
                  onClose={onClose}
                  installId={step.metadata?.install_id}
                />
              ) : (
                getStepInstructions(step.name)
              )}
            </ChecklistItem>
          ))}
        </div>
      </div>
    </Modal>
  )
}

