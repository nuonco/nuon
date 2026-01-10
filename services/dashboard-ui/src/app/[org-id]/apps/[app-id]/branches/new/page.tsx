'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/common/Button'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { StepVCSConfig } from './step-vcs-config'
import { StepDeploymentOrder } from './step-deployment-order'
import { StepReview } from './step-review'

// Mock data
export const mockVCSConnections = [
  { id: 'vcs1', name: 'github-org', type: 'github' },
  { id: 'vcs2', name: 'github-personal', type: 'github' },
]

export const mockRepos = [
  { id: 'repo1', name: 'nuonco/app-backend', private: true },
  { id: 'repo2', name: 'nuonco/app-frontend', private: false },
  { id: 'repo3', name: 'nuonco/infrastructure', private: true },
]

export const mockBranches = ['main', 'develop', 'staging', 'production']

export const mockInstalls = [
  {
    id: 'ins1',
    name: 'production-us-east',
    region: 'us-east-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins2',
    name: 'production-us-west',
    region: 'us-west-2',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins3',
    name: 'staging-us-east',
    region: 'us-east-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins4',
    name: 'staging-us-west',
    region: 'us-west-2',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins5',
    name: 'dev-environment',
    region: 'us-east-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins6',
    name: 'qa-environment',
    region: 'eu-west-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins7',
    name: 'demo-environment',
    region: 'us-west-2',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins8',
    name: 'sandbox-1',
    region: 'us-east-1',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins9',
    name: 'sandbox-2',
    region: 'us-west-2',
    status: 'active' as const,
    platform: 'aws' as const,
  },
  {
    id: 'ins10',
    name: 'test-env',
    region: 'us-east-1',
    status: 'inactive' as const,
    platform: 'aws' as const,
  },
]

export interface IFormData {
  branchName: string
  isManualOnly: boolean
  vcsConnection: string
  repo: string
  gitBranch: string
  directory: string
  pathFilter: string
  deploymentGroups: string[][]
  ungroupedInstalls: string[]
}

interface INewBranchPageProps {
  params: Promise<{
    'org-id': string
    'app-id': string
  }>
}

export default function NewBranchPage({ params }: INewBranchPageProps) {
  const router = useRouter()
  const [currentStep, setCurrentStep] = useState(1)
  const [formData, setFormData] = useState<IFormData>({
    branchName: '',
    isManualOnly: false,
    vcsConnection: '',
    repo: '',
    gitBranch: 'main',
    directory: '.',
    pathFilter: '',
    deploymentGroups: [],
    ungroupedInstalls: mockInstalls.map((i) => i.id),
  })

  // We need to unwrap the params promise
  const [orgId, setOrgId] = useState<string>('')
  const [appId, setAppId] = useState<string>('')

  // Unwrap params on mount
  useState(() => {
    params.then((p) => {
      setOrgId(p['org-id'])
      setAppId(p['app-id'])
    })
  })

  const updateFormData = (updates: Partial<IFormData>) => {
    setFormData((prev) => ({ ...prev, ...updates }))
  }

  const canProceedFromStep1 = () => {
    if (!formData.branchName) return false
    if (!formData.isManualOnly) {
      return Boolean(
        formData.vcsConnection && formData.repo && formData.gitBranch
      )
    }
    return true
  }

  const canProceedFromStep2 = () => {
    // At least one install must be in a group
    return formData.deploymentGroups.some((group) => group.length > 0)
  }

  const handleNext = () => {
    if (currentStep === 1 && canProceedFromStep1()) {
      setCurrentStep(2)
    } else if (currentStep === 2 && canProceedFromStep2()) {
      setCurrentStep(3)
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

  const handleCreate = () => {
    console.log('Creating branch with data:', formData)
    // Mock action - just log and navigate back
    if (orgId && appId) {
      router.push(`/${orgId}/apps/${appId}/branches`)
    }
  }

  const steps = [
    { number: 1, title: 'VCS Configuration' },
    { number: 2, title: 'Deployment Order' },
    { number: 3, title: 'Review & Create' },
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

      <div className="flex items-center gap-4 justify-between mb-8">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Create New Branch Configuration
          </Text>
        </HeadingGroup>
      </div>

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
          <StepVCSConfig formData={formData} updateFormData={updateFormData} />
        )}
        {currentStep === 2 && (
          <StepDeploymentOrder
            formData={formData}
            updateFormData={updateFormData}
          />
        )}
        {currentStep === 3 && <StepReview formData={formData} />}
      </div>

      {/* Navigation Buttons */}
      <div className="flex items-center gap-3 justify-between border-t pt-6">
        <Button variant="ghost" onClick={handleCancel}>
          Cancel
        </Button>

        <div className="flex items-center gap-3">
          {currentStep > 1 && (
            <Button variant="secondary" onClick={handleBack}>
              Back
            </Button>
          )}

          {currentStep < 3 && (
            <Button
              variant="primary"
              onClick={handleNext}
              disabled={
                (currentStep === 1 && !canProceedFromStep1()) ||
                (currentStep === 2 && !canProceedFromStep2())
              }
            >
              Next
            </Button>
          )}

          {currentStep === 3 && (
            <Button variant="primary" onClick={handleCreate}>
              Create Branch
            </Button>
          )}
        </div>
      </div>
    </PageSection>
  )
}
