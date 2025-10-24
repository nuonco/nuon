'use client'

import React, { type FC, useState, useEffect } from 'react'
import { ArrowLeft, ArrowRight, Check } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Text } from '@/components/old/Typography'
import { useAccount } from '@/hooks/use-account'
import type { TAccount, TUserJourney, TUserJourneyStep } from '@/types'
import { OnboardingStepHeader } from './OnboardingStepHeader'
import { CreateAccountStepContent } from './CreateAccountStepContent'
import { CLIInstallStepContent } from './CLIInstallStepContent'
import { CreateAppStepContent } from './CreateAppStepContent'
import { AppSyncStepContent } from './AppSyncStepContent'
import { InstallCreationStepContent } from './InstallCreationStepContent'
import { OrgCreationStepContent } from './OrgCreationStepContent'
import { completeUserJourney } from '@/actions/accounts/complete-user-journey'
import { getUserJourneyStepMetadata } from '@/utils/user-journey-utils'

interface FullScreenOnboardingProps {
  isOpen: boolean
  onClose: () => void
  account: TAccount | null
}

// Removed getNextStep and detectNewlyCompletedStep functions
// These were used for automatic step advancement which is no longer needed

// Navigation component
const OnboardingNavigation: FC<{
  currentStepIndex: number
  steps: TUserJourneyStep[]
  onPreviousStep: () => void
  onNextStep: () => void
  onSkip: () => void
  canNavigateBack: boolean
  canNavigateForward: boolean
  canSkip: boolean
}> = ({
  currentStepIndex,
  steps,
  onPreviousStep,
  onNextStep,
  onSkip,
  canNavigateBack,
  canNavigateForward,
  canSkip,
}) => {
  return (
    <div className="border-b flex items-center justify-between p-4 md:p-6 bg-white/95 dark:bg-dark-grey-900/95 backdrop-blur-sm z-10 transition-all duration-200">
      {/* Left: Navigation arrows */}
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          className="!p-2 transition-all duration-200 hover:scale-110 disabled:opacity-50 disabled:cursor-not-allowed"
          onClick={onPreviousStep}
          disabled={!canNavigateBack}
        >
          <ArrowLeft size={20} />
        </Button>
        <Button
          variant="ghost"
          className="!p-2 transition-all duration-200 hover:scale-110 disabled:opacity-50 disabled:cursor-not-allowed"
          onClick={onNextStep}
          disabled={!canNavigateForward}
        >
          <ArrowRight size={20} />
        </Button>
      </div>

      {/* Center: Progress indicator */}
      <div className="flex items-center justify-center">
        <div className="flex items-center space-x-2">
          {steps.map((step, index) => (
            <React.Fragment key={step.name}>
              {/* Step circle */}
              <div
                className={`w-4 h-4 md:w-6 md:h-6 rounded-full flex items-center justify-center text-xs font-medium transition-all duration-300 ease-out transform flex-shrink-0 ${
                  step.complete
                    ? 'bg-green-500 text-white scale-100'
                    : index === currentStepIndex
                      ? 'bg-blue-500 text-white scale-110'
                      : 'bg-gray-200 dark:bg-gray-700 text-gray-500 scale-90'
                }`}
              >
                {step.complete ? (
                  <Check
                    size={8}
                    weight="bold"
                    className="md:w-3 md:h-3 transition-transform duration-200"
                  />
                ) : null}
              </div>

              {/* Connector line */}
              {index < steps.length - 1 && (
                <div
                  className={`h-0.5 w-4 md:w-6 rounded-full transition-all duration-300 ease-out flex-shrink-0 ${
                    step.complete
                      ? 'bg-green-500'
                      : 'bg-gray-200 dark:bg-gray-700'
                  }`}
                />
              )}
            </React.Fragment>
          ))}
        </div>
      </div>

      <div className="flex items-center gap-2 md:gap-3">
        {canSkip && (
          <Button
            variant="secondary"
            onClick={onSkip}
            className="transition-all duration-200 hover:scale-105"
          >
            <span className="hidden sm:inline">Skip</span>
            <span className="sm:hidden">Skip</span>
          </Button>
        )}
      </div>
    </div>
  )
}

export const FullScreenOnboarding: FC<FullScreenOnboardingProps> = ({
  isOpen,
  onClose,
  account,
}) => {
  const { refreshAccount } = useAccount()
  // Removed previousJourneyState - no longer needed without step-by-step auto-advancement
  const [manualStepIndex, setManualStepIndex] = useState<number | null>(null) // For manual navigation
  const [stepTransition, setStepTransition] = useState<'enter' | 'exit' | null>(
    null
  )

  // Get evaluation journey
  const accountWithJourneys = account as any
  const evaluationJourney = accountWithJourneys?.user_journeys?.find(
    (journey: TUserJourney) => journey.name === 'evaluation'
  )

  // Show all journey steps including automatic ones for complete visibility
  const allJourneySteps = evaluationJourney?.steps || []

  // Find first incomplete step for navigation guidance (not auto-jumping)
  const firstIncompleteStepIndex = allJourneySteps.findIndex(
    (step: TUserJourneyStep) => !step.complete
  )

  // Determine which step to display: manual navigation takes precedence, otherwise start at step 0
  const displayStepIndex = manualStepIndex !== null ? manualStepIndex : 0 // Always start at the first step (account_created) for user control

  const displayStep = allJourneySteps[displayStepIndex] || null

  // Determine if we should show the advancement button
  const shouldShowAdvanceButton =
    displayStep?.complete &&
    firstIncompleteStepIndex !== -1 &&
    displayStepIndex !== firstIncompleteStepIndex

  // Get the next step name for button text
  const nextStepName =
    firstIncompleteStepIndex !== -1
      ? allJourneySteps[firstIncompleteStepIndex]?.name
      : null

  if (!isOpen || !evaluationJourney || allJourneySteps.length === 0) {
    return null
  }

  // Server action to complete evaluation journey
  const completeEvaluationJourneyAction = async (): Promise<boolean> => {
    try {
      await completeUserJourney({ journeyName: 'evaluation' })
      return true
    } catch (err) {
      console.error('Exception completing evaluation journey:', err)
      return false
    }
  }

  const handleSkipAll = async () => {
    const success = await completeEvaluationJourneyAction()
    if (success) {
      // Refresh account data so the provider detects all steps are complete
      // This will trigger the modal to close automatically via the journey logic
      onClose()
    } else {
      console.warn(
        'Skip failed - user may need to refresh the page or try again'
      )
    }
  }

  // Manual navigation functions with smooth transitions
  const handlePreviousStep = () => {
    if (displayStepIndex > 0) {
      setStepTransition('exit')
      setTimeout(() => {
        setManualStepIndex(displayStepIndex - 1)
        setStepTransition('enter')
      }, 150)
      setTimeout(() => setStepTransition(null), 450)
    }
  }

  const handleNextStep = () => {
    if (displayStepIndex < allJourneySteps.length - 1) {
      setStepTransition('exit')
      setTimeout(() => {
        setManualStepIndex(displayStepIndex + 1)
        setStepTransition('enter')
      }, 150)
      setTimeout(() => setStepTransition(null), 450)
    }
  }

  // Navigate to the first incomplete step with smooth transition
  const handleAdvanceToCurrentStep = () => {
    if (firstIncompleteStepIndex === -1) return

    setStepTransition('exit')
    setTimeout(() => {
      setManualStepIndex(firstIncompleteStepIndex)
      setStepTransition('enter')
    }, 150)
    setTimeout(() => setStepTransition(null), 450)
  }

  // Generate contextual button text based on step name
  const getAdvanceButtonText = (stepName: string | null): string => {
    switch (stepName) {
      case 'org_created':
        return 'Continue to Organization Setup'
      case 'cli_installed':
        return 'Continue to CLI Installation'
      case 'app_created':
        return 'Continue to App Creation'
      case 'app_synced':
        return 'Continue to App Sync'
      case 'install_created':
        return 'Continue to Install Creation'
      default:
        return 'Continue to Next Step'
    }
  }

  const orgId = getUserJourneyStepMetadata(
    account,
    'evaluation',
    'org_created',
    'org_id'
  )
  const appId = getUserJourneyStepMetadata(
    account,
    'evaluation',
    'app_created',
    'app_id'
  )

  const canNavigateBack = displayStepIndex > 0
  const canNavigateForward = displayStepIndex < allJourneySteps.length - 1

  return (
    <div className="fixed inset-0 z-100 bg-white dark:bg-dark-grey-900 overflow-hidden transition-all duration-300 ease-out flex flex-col">
      {/* Navigation Header */}
      <OnboardingNavigation
        currentStepIndex={displayStepIndex}
        steps={allJourneySteps}
        onPreviousStep={handlePreviousStep}
        onNextStep={handleNextStep}
        onSkip={handleSkipAll}
        canNavigateBack={canNavigateBack}
        canNavigateForward={canNavigateForward}
        canSkip={!!orgId}
      />

      {/* Main Content */}
      <div className="pb-8 p-4 md:p-8 overflow-y-auto flex-1">
        <div className="max-w-4xl mx-auto">
          {/* Single Step Display */}
          {displayStep && (
            <div
              className={`transition-all duration-300 ease-out ${
                stepTransition === 'exit'
                  ? 'opacity-0 transform translate-x-4'
                  : stepTransition === 'enter'
                    ? 'opacity-100 transform translate-x-0'
                    : 'opacity-100 transform translate-x-0'
              }`}
            >
              {/* Step Header */}
              <OnboardingStepHeader step={displayStep} />

              {/* Step Content */}
              <div className="max-w-3xl mx-auto p-6 md:p-8">
                {displayStep.name === 'account_created' ? (
                  <CreateAccountStepContent
                    stepComplete={displayStep.complete}
                    account={account}
                  />
                ) : displayStep.name === 'org_created' ? (
                  <OrgCreationStepContent
                    stepComplete={displayStep.complete}
                    orgId={orgId}
                  />
                ) : displayStep.name === 'cli_installed' ? (
                  <CLIInstallStepContent stepComplete={displayStep.complete} />
                ) : displayStep.name === 'app_created' ? (
                  <CreateAppStepContent
                    stepComplete={displayStep.complete}
                    appId={displayStep.metadata?.app_id}
                  />
                ) : displayStep.name === 'app_synced' ? (
                  <AppSyncStepContent
                    stepComplete={displayStep.complete}
                    selectedAppPath={
                      displayStep.metadata?.app_path || 'eks-simple'
                    }
                  />
                ) : displayStep.name === 'install_created' ? (
                  <InstallCreationStepContent
                    stepComplete={displayStep.complete}
                    onClose={handleSkipAll}
                    installId={displayStep.metadata?.install_id}
                    appId={appId}
                    orgId={orgId}
                  />
                ) : (
                  <div className="text-center py-8">
                    <Text variant="reg-14">Step content not available</Text>
                  </div>
                )}

                {/* Smart navigation button for completed steps */}
                {shouldShowAdvanceButton && (
                  <div className="mt-6 flex justify-end">
                    <Button
                      variant="primary"
                      onClick={handleAdvanceToCurrentStep}
                      className="px-3 py-1 text-sm"
                    >
                      <Text>{getAdvanceButtonText(nextStepName)}</Text>
                    </Button>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
