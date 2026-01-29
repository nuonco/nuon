'use client'

import React, { type FC, useState } from 'react'
import { useRouter } from 'next/navigation'
import { ArrowLeft, ArrowRight, Check } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Text } from '@/components/old/Typography'
import { useAccount } from '@/hooks/use-account'
import type { TUserJourney, TUserJourneyStep } from '@/types'
import { OnboardingStepHeader } from '@/components/old/Apps/OnboardingStepHeader'
import { CreateAccountStepContent } from '@/components/old/Apps/CreateAccountStepContent'
import { CreateAppStepContent } from '@/components/old/Apps/CreateAppStepContent'
import { InstallCreationStepContent } from '@/components/old/Apps/InstallCreationStepContent'
import { OrgCreationStepContent } from '@/components/old/Apps/OrgCreationStepContent'
import { CLIInstallStepContent } from '@/components/old/Apps/CLIInstallStepContent'
import { AppSyncStepContent } from '@/components/old/Apps/AppSyncStepContent'
import { getUserJourneyStepMetadata } from '@/utils/user-journey-utils'

const OnboardingNavigation: FC<{
  currentStepIndex: number
  steps: TUserJourneyStep[]
  onPreviousStep: () => void
  onNextStep: () => void
  onStepClick: (index: number) => void
  onSkip: () => void
  canNavigateBack: boolean
  canNavigateForward: boolean
  canSkip: boolean
  firstIncompleteStepIndex: number
}> = ({
  currentStepIndex,
  steps,
  onPreviousStep,
  onNextStep,
  onStepClick,
  onSkip,
  canNavigateBack,
  canNavigateForward,
  canSkip,
  firstIncompleteStepIndex,
}) => {
  return (
    <div className="border-b flex items-center justify-between p-4 md:p-6 bg-white/95 dark:bg-dark-grey-900/95 backdrop-blur-sm z-10 transition-all duration-200">
      <div className="flex-1">
        {canSkip && (
          <Button
            variant="secondary"
            onClick={onSkip}
            className="transition-all duration-200 hover:scale-105"
          >
            Close
          </Button>
        )}
      </div>

      <div className="flex items-center justify-center gap-3">
        <Button
          variant="ghost"
          className="!p-2 transition-all duration-200 hover:scale-110 disabled:opacity-50 disabled:cursor-not-allowed"
          onClick={onPreviousStep}
          disabled={!canNavigateBack}
        >
          <ArrowLeft size={20} />
        </Button>

        <div className="flex items-center space-x-2">
          {steps.map((step, index) => {
            const canClick =
              step.complete ||
              index === firstIncompleteStepIndex ||
              (firstIncompleteStepIndex === -1 && index <= steps.length - 1)
            return (
              <React.Fragment key={step.name}>
                <button
                  type="button"
                  onClick={() => canClick && onStepClick(index)}
                  disabled={!canClick}
                  className={`w-4 h-4 md:w-6 md:h-6 rounded-full flex items-center justify-center text-xs font-medium transition-all duration-300 ease-out transform flex-shrink-0 ${
                    index === firstIncompleteStepIndex
                      ? 'bg-blue-500 text-white scale-110'
                      : step.complete
                        ? 'bg-green-500 text-white scale-100'
                        : 'bg-gray-200 dark:bg-gray-700 text-gray-500 scale-90'
                  } ${canClick ? 'cursor-pointer hover:scale-125' : 'cursor-not-allowed'}`}
                >
                  {step.complete ? (
                    <Check
                      size={8}
                      weight="bold"
                      className="md:w-3 md:h-3 transition-transform duration-200"
                    />
                  ) : null}
                </button>

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
            )
          })}
        </div>

        <Button
          variant="ghost"
          className="!p-2 transition-all duration-200 hover:scale-110 disabled:opacity-50 disabled:cursor-not-allowed"
          onClick={onNextStep}
          disabled={!canNavigateForward}
        >
          <ArrowRight size={20} />
        </Button>
      </div>

      <div className="flex-1" />
    </div>
  )
}

export const OnboardingPageContent: FC = () => {
  const router = useRouter()
  const { account } = useAccount()
  const [manualStepIndex, setManualStepIndex] = useState<number | null>(null)
  const [stepTransition, setStepTransition] = useState<'enter' | 'exit' | null>(
    null
  )
  const [sfData, setSFData] = useState<Record<string, string>>()

  const accountWithJourneys = account as any
  const evaluationJourney = accountWithJourneys?.user_journeys?.find(
    (journey: TUserJourney) => journey.name === 'evaluation'
  )

  const allJourneySteps = evaluationJourney?.steps || []

  const firstIncompleteStepIndex = allJourneySteps.findIndex(
    (step: TUserJourneyStep) => !step.complete
  )

  const defaultStepIndex =
    firstIncompleteStepIndex !== -1
      ? firstIncompleteStepIndex
      : allJourneySteps.length - 1
  const displayStepIndex =
    manualStepIndex !== null ? manualStepIndex : defaultStepIndex
  const displayStep = allJourneySteps[displayStepIndex] || null

  const nextStepIndex = displayStepIndex + 1
  const nextStep = allJourneySteps[nextStepIndex]
  const shouldShowAdvanceButton = nextStep !== undefined

  const nextStepName = nextStep?.name || null

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

  const savedJobTitle = getUserJourneyStepMetadata(
    account,
    'evaluation',
    'account_created',
    'jobTitle'
  )
  const savedCompanyName = getUserJourneyStepMetadata(
    account,
    'evaluation',
    'account_created',
    'companyName'
  )
  const savedNotes = getUserJourneyStepMetadata(
    account,
    'evaluation',
    'account_created',
    'notes'
  )

  const initialFormValues = {
    jobTitle: savedJobTitle || '',
    companyName: savedCompanyName || '',
    notes: savedNotes || '',
  }

  const handleComplete = () => {
    if (orgId) {
      router.push(`/${orgId}/apps`)
    } else {
      router.push('/')
    }
  }

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

  const handleAdvanceToNextStep = () => {
    if (nextStepIndex >= allJourneySteps.length) return

    setStepTransition('exit')
    setTimeout(() => {
      setManualStepIndex(nextStepIndex)
      setStepTransition('enter')
    }, 150)
    setTimeout(() => setStepTransition(null), 450)
  }

  const handleStepClick = (index: number) => {
    if (index === displayStepIndex) return

    setStepTransition('exit')
    setTimeout(() => {
      setManualStepIndex(index)
      setStepTransition('enter')
    }, 150)
    setTimeout(() => setStepTransition(null), 450)
  }

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

  if (!evaluationJourney || allJourneySteps.length === 0) {
    return (
      <div className="min-h-screen bg-white dark:bg-dark-grey-900 flex items-center justify-center">
        <div className="text-center">
          <Text variant="semi-18" className="mb-4">
            Loading onboarding...
          </Text>
        </div>
      </div>
    )
  }

  const canNavigateBack = displayStepIndex > 0
  const canNavigateForward =
    displayStepIndex < allJourneySteps.length - 1 &&
    (firstIncompleteStepIndex === -1 || displayStepIndex < firstIncompleteStepIndex)

  return (
    <div className="h-screen bg-white dark:bg-dark-grey-900 overflow-hidden transition-all duration-300 ease-out flex flex-col">
      <OnboardingNavigation
        currentStepIndex={displayStepIndex}
        steps={allJourneySteps}
        onPreviousStep={handlePreviousStep}
        onNextStep={handleNextStep}
        onStepClick={handleStepClick}
        onSkip={handleComplete}
        canNavigateBack={canNavigateBack}
        canNavigateForward={canNavigateForward}
        canSkip={!!orgId}
        firstIncompleteStepIndex={firstIncompleteStepIndex}
      />

      <div className="pb-8 p-4 md:p-8 overflow-y-auto flex-1">
        <div className="max-w-4xl mx-auto">
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
              <OnboardingStepHeader step={displayStep} />

              <div className="max-w-3xl mx-auto p-6 md:p-8">
                {displayStep.name === 'account_created' ? (
                  <CreateAccountStepContent
                    stepComplete={displayStep.complete}
                    account={account}
                    setSFData={setSFData}
                    initialValues={initialFormValues}
                  />
                ) : displayStep.name === 'org_created' ? (
                  <OrgCreationStepContent
                    stepComplete={displayStep.complete}
                    orgId={orgId}
                    sfData={sfData}
                    skipNavigation
                  />
                ) : displayStep.name === 'cli_installed' ? (
                  <CLIInstallStepContent stepComplete={displayStep.complete} />
                ) : displayStep.name === 'app_created' ? (
                  <CreateAppStepContent
                    stepComplete={displayStep.complete}
                    appId={displayStep.metadata?.app_id}
                  />
                ) : displayStep.name === 'app_synced' ? (
                  <AppSyncStepContent stepComplete={displayStep.complete} />
                ) : displayStep.name === 'install_created' ? (
                  <InstallCreationStepContent
                    stepComplete={displayStep.complete}
                    onClose={handleComplete}
                    installId={displayStep.metadata?.install_id}
                    appId={appId}
                    orgId={orgId}
                  />
                ) : (
                  <div className="text-center py-8">
                    <Text variant="reg-14">Step content not available</Text>
                  </div>
                )}

                {shouldShowAdvanceButton && (
                  <div className="mt-6 flex justify-end">
                    <Button
                      form="sf-form"
                      variant="primary"
                      onClick={handleAdvanceToNextStep}
                      className="px-3 py-1 text-sm"
                      disabled={displayStep?.name !== 'account_created' && !displayStep?.complete}
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
