'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/common/Button'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { StepBasicInfo } from './step-basic-info'
import { StepVCSConfig } from './step-vcs-config'
import { StepInstallGroups } from './step-install-groups'
import { StepReview } from './step-review'
import { IFormData, IAppBranchConfig } from './types'
import { getMockInstalls, saveConfigToLocalStorage } from './mock-data'

interface INewBranchPageProps {
  params: Promise<{
    'org-id': string
    'app-id': string
  }>
}

export default function NewBranchPage({ params }: INewBranchPageProps) {
  const router = useRouter()
  const [currentStep, setCurrentStep] = useState(1)
  const mockInstalls = getMockInstalls()
  const [formData, setFormData] = useState<IFormData>({
    branchName: '',
    description: '',
    isManualOnly: false,
    vcsConnection: '',
    repo: '',
    gitBranch: 'main',
    directory: '.',
    pathFilter: '',
    installGroups: [],
    ungroupedInstalls: mockInstalls.map((i) => i.id),
  })

  // We need to unwrap the params promise
  const [orgId, setOrgId] = useState<string>('')
  const [appId, setAppId] = useState<string>('')

  // Unwrap params on mount
  useEffect(() => {
    params.then((p) => {
      setOrgId(p['org-id'])
      setAppId(p['app-id'])
    })
  }, [params])

  const updateFormData = (updates: Partial<IFormData>) => {
    setFormData((prev) => ({ ...prev, ...updates }))
  }

  const canProceedFromStep1 = () => {
    return Boolean(formData.branchName.trim())
  }

  const canProceedFromStep2 = () => {
    if (!formData.isManualOnly) {
      return Boolean(
        formData.vcsConnection && formData.repo && formData.gitBranch
      )
    }
    return true
  }

  const canProceedFromStep3 = () => {
    // At least one install must be in a group
    return formData.installGroups.some((group) => group.installIds.length > 0)
  }

  const handleNext = () => {
    if (currentStep === 1 && canProceedFromStep1()) {
      setCurrentStep(2)
    } else if (currentStep === 2 && canProceedFromStep2()) {
      setCurrentStep(3)
    } else if (currentStep === 3 && canProceedFromStep3()) {
      setCurrentStep(4)
    }
  }

  const handleBack = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1)
    }
  }

  const handleCancel = () => {
    if (orgId && appId) {
      router.push(`/${orgId}/apps/${appId}/branches`)
    }
  }

  const [isSubmitting, setIsSubmitting] = useState(false)
  const [createdBranchId, setCreatedBranchId] = useState<string | null>(null)

  const handleCreate = async () => {
    console.log('🚀 handleCreate called with formData:', formData)
    console.log('📍 Current step:', currentStep)
    
    setIsSubmitting(true)

    try {
      // Simulate API delay for better UX (reduced to 300ms)
      await new Promise((resolve) => setTimeout(resolve, 300))

      // Generate a mock branch ID
      const branchId = `branch-${Date.now()}`

      // Convert formData to IAppBranchConfig
      const config: IAppBranchConfig = {
        name: formData.branchName,
        description: formData.description,
        vcsEnabled: !formData.isManualOnly,
        vcsConnectionId: formData.vcsConnection || undefined,
        repository: formData.repo || undefined,
        gitBranch: formData.gitBranch || undefined,
        directory: formData.directory || undefined,
        pathFilter: formData.pathFilter || undefined,
        installGroups: formData.installGroups,
      }

      // Save to localStorage
      saveConfigToLocalStorage(branchId, config)

      console.log('✅ Created branch configuration (stub):', config)
      console.log('📦 Saved to localStorage key:', `app-branch-config-${branchId}`)
      console.log('🎉 Branch ID:', branchId)

      // Set success state
      setCreatedBranchId(branchId)
    } catch (error) {
      console.error('❌ Error creating branch (stub):', error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const steps = [
    { number: 1, title: 'Basic Info' },
    { number: 2, title: 'VCS Configuration' },
    { number: 3, title: 'Install Groups' },
    { number: 4, title: 'Review & Create' },
  ]

  return (
    <PageSection isScrollable>
      {orgId && appId && (
        <Breadcrumbs
          breadcrumbs={[
            {
              path: `/${orgId}`,
              text: 'Organization',
            },
            {
              path: `/${orgId}/apps`,
              text: 'Apps',
            },
            {
              path: `/${orgId}/apps/${appId}`,
              text: 'App',
            },
            {
              path: `/${orgId}/apps/${appId}/branches`,
              text: 'Branches',
            },
            {
              path: `/${orgId}/apps/${appId}/branches/new`,
              text: 'New',
            },
          ]}
        />
      )}

      <form onSubmit={(e) => e.preventDefault()}>
        <div className="flex items-center gap-4 justify-between mb-8">
          <HeadingGroup>
            <Text variant="base" weight="strong">
              Create New Branch Configuration
            </Text>
          </HeadingGroup>
        </div>

        {/* Debug Info in Development */}
        {process.env.NODE_ENV === 'development' && (
          <div className="mb-4 p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <Text variant="xs" className="font-mono">
              Step: {currentStep}/4 | Manual: {formData.isManualOnly ? 'Yes' : 'No'} | Name: {formData.branchName || '(empty)'} | VCS: {formData.vcsConnection || '(none)'} | Repo: {formData.repo || '(none)'}
            </Text>
          </div>
        )}

      {/* Step Indicator */}
      <div className="flex items-center gap-4 mb-8">
        {steps.map((step, index) => (
          <div key={step.number} className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-strong ${
                  currentStep === step.number
                    ? 'bg-primary-600 text-white'
                    : currentStep > step.number
                      ? 'bg-green-600 text-white'
                      : 'bg-cool-grey-200 dark:bg-dark-grey-700 text-cool-grey-600 dark:text-cool-grey-400'
                }`}
              >
                {currentStep > step.number ? '✓' : step.number}
              </div>
              <Text
                variant="sm"
                weight={currentStep === step.number ? 'strong' : 'normal'}
              >
                {step.title}
              </Text>
            </div>
            {index < steps.length - 1 && (
              <div className="w-12 h-0.5 bg-cool-grey-200 dark:bg-dark-grey-700" />
            )}
          </div>
        ))}
      </div>

      {/* Step Content */}
      <div className="mb-8">
        {currentStep === 1 && (
          <StepBasicInfo formData={formData} updateFormData={updateFormData} />
        )}
        {currentStep === 2 && (
          <StepVCSConfig formData={formData} updateFormData={updateFormData} />
        )}
        {currentStep === 3 && (
          <StepInstallGroups
            formData={formData}
            updateFormData={updateFormData}
          />
        )}
        {currentStep === 4 && (
          <StepReview
            formData={formData}
            isSuccess={!!createdBranchId}
            branchId={createdBranchId || undefined}
            orgId={orgId}
            appId={appId}
            onSubmit={handleCreate}
            isSubmitting={isSubmitting}
          />
        )}
      </div>

        {/* Navigation Buttons */}
        {!createdBranchId && (
          <div className="flex items-center gap-3 justify-between border-t pt-6">
            <Button variant="ghost" onClick={handleCancel} type="button">
              Cancel
            </Button>

            <div className="flex items-center gap-3">
              {currentStep > 1 && (
                <Button variant="secondary" onClick={handleBack} type="button">
                  Back
                </Button>
              )}

              {currentStep < 4 && (
                <Button
                  variant="primary"
                  onClick={handleNext}
                  type="button"
                  data-wizard-next
                  disabled={
                    (currentStep === 1 && !canProceedFromStep1()) ||
                    (currentStep === 2 && !canProceedFromStep2()) ||
                    (currentStep === 3 && !canProceedFromStep3())
                  }
                >
                  Next
                </Button>
              )}

              {currentStep === 4 && (
                <Button
                  variant="primary"
                  onClick={handleCreate}
                  type="button"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? 'Creating...' : 'Create Branch'}
                </Button>
              )}
            </div>
          </div>
        )}

        {/* Quick Navigation for Development */}
        {process.env.NODE_ENV === 'development' && currentStep < 3 && (
          <div className="fixed bottom-4 right-4 z-50">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setCurrentStep(3)}
              type="button"
              className="shadow-lg"
            >
              Skip to Install Groups
            </Button>
          </div>
        )}
      </form>
    </PageSection>
  )
}