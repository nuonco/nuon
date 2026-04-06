import { useEffect, useRef } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import type { TIconVariant } from '@/components/common/Icon'
import type { TOnboarding } from '@/types'
import { completeGetStartedStep } from '@/lib'
import { useConfetti } from '@/hooks/use-confetti'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

// --- Data model ---

interface INextStep {
  icon: TIconVariant
  title: string
  description: string
  href: string
}

interface ISection {
  tabKey: string
  description: string
  graphic: string
  steps: INextStep[]
}

function buildSections(): ISection[] {
  return [
    {
      tabKey: 'connect your app',
      description: 'Wire your SaaS into Nuon. Define install inputs, connect your CI/CD pipeline, and configure component sources.',
      graphic: '/onboarding-graphics/connect-app-dark.png',
      steps: [
        {
          icon: 'SlidersHorizontal',
          title: 'App inputs & config',
          description: 'Define what customers provide at install time',
          href: 'https://docs.nuon.co/guides/configuring-inputs-and-secrets',
        },
        {
          icon: 'GitBranch',
          title: 'Connect CI/CD',
          description: 'Trigger deploys from GitHub Actions, etc.',
          href: 'https://docs.nuon.co/guides/github-actions',
        },
        {
          icon: 'Package',
          title: 'Component sources',
          description: 'Helm, Terraform, Docker, raw manifests',
          href: 'https://docs.nuon.co/concepts/components',
        },
        {
          icon: 'TerminalWindow',
          title: 'nuon apps init',
          description: 'Pull your app config locally with the CLI',
          href: 'https://docs.nuon.co/guides/app-init',
        },
      ],
    },
    {
      tabKey: 'day 2 operations',
      description: 'Keep customer installs healthy, secure, and auditable after go-live.',
      graphic: '/onboarding-graphics/day2-ops-dark.png',
      steps: [
        {
          icon: 'ArrowsClockwise',
          title: 'Drift detection',
          description: 'Automatic detection of infrastructure changes with scheduled scans and remediation',
          href: 'https://docs.nuon.co/updates/021-drift-detection',
        },
        {
          icon: 'ShieldCheck',
          title: 'Break-glass access',
          description: 'Secure emergency access to customer environments with full audit trails',
          href: 'https://docs.nuon.co/config-ref/break-glass',
        },
        {
          icon: 'Scales',
          title: 'Policies',
          description: 'Enforce compliance and security standards across deployments',
          href: 'https://docs.nuon.co/guides/configuring-policies',
        },
        {
          icon: 'UserCheck',
          title: 'Approvals',
          description: 'Let customers review and approve updates on their terms',
          href: 'https://docs.nuon.co/updates/010-approvals',
        },
      ],
    },
    {
      tabKey: 'create an installer',
      description: 'Give customers a self-service portal to install, configure, and manage their own instance.',
      graphic: '/onboarding-graphics/installer-dark.png',
      steps: [
        {
          icon: 'Layout',
          title: 'Installer overview',
          description: 'What the customer-facing portal looks like',
          href: 'https://docs.nuon.co/concepts/customer-portal',
        },
        {
          icon: 'Palette',
          title: 'Portal branding',
          description: 'Custom domain, logo, colours, and copy',
          href: 'https://docs.nuon.co/guides/custom-domains',
        },
        {
          icon: 'TextAa',
          title: 'Customer inputs',
          description: 'What you ask customers to provide before install',
          href: 'https://docs.nuon.co/concepts/app-inputs',
        },
        {
          icon: 'CodeBlock',
          title: 'Embed the installer',
          description: 'Drop the portal into your app or docs site',
          href: 'https://docs.nuon.co/guides/control-plane-integration',
        },
      ],
    },
  ]
}

// --- Step row ---

function StepRow({ step }: { step: INextStep }) {
  return (
    <Card className="!p-4 !gap-0 flex-row items-center">
      <div className="flex items-center justify-center w-8 h-8 rounded-full border shrink-0">
        <Icon variant={step.icon} size={16} />
      </div>
      <div className="flex flex-col flex-1 min-w-0 ml-4">
        <Text variant="base" weight="strong">{step.title}</Text>
        <Text variant="body" theme="neutral">{step.description}</Text>
      </div>
      <Text variant="subtext" className="shrink-0 ml-4">
        <Link href={step.href} isExternal>
          Learn more <Icon variant="ArrowSquareOutIcon" size={12} />
        </Link>
      </Text>
    </Card>
  )
}

// --- Tab content for a section ---

function SectionContent({ section }: { section: ISection }) {
  return (
    <div className="flex flex-col gap-5 pt-6">
      <div className="overflow-hidden rounded-lg border bg-cool-grey-50 dark:bg-dark-grey-900">
        <img
          src={section.graphic}
          alt={section.tabKey}
          className="w-full h-auto object-cover"
        />
      </div>
      <Text variant="body" theme="neutral">{section.description}</Text>
      <div className="flex flex-col gap-3">
        {section.steps.map((step) => (
          <StepRow key={step.title} step={step} />
        ))}
      </div>
    </div>
  )
}

// --- Main step ---

export const NextStepsStep = ({ onAdvance, sharedData }: IWizardStepComponentProps) => {
  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id
  const installId = onboarding?.install_id

  const sections = buildSections()

  const { fireCelebrationConfetti } = useConfetti()
  const confettiFired = useRef(false)

  useEffect(() => {
    if (!confettiFired.current) {
      confettiFired.current = true
      const timer = setTimeout(() => fireCelebrationConfetti(), 300)
      return () => clearTimeout(timer)
    }
  }, [fireCelebrationConfetti])

  const { mutate: completeStep } = useMutation({
    mutationFn: () => completeGetStartedStep({ orgId: orgId! }),
  })

  const handleContinue = () => {
    if (orgId) completeStep()
    if (orgId && installId) {
      window.location.href = `/${orgId}/installs/${installId}`
    } else {
      onAdvance()
    }
  }

  const tabs: Record<string, React.ReactNode> = {}
  for (const section of sections) {
    tabs[section.tabKey] = <SectionContent section={section} />
  }

  return (
    <div className="flex flex-col gap-8">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col gap-1">
          <div className="flex items-center gap-3">
            <Text variant="h2" role="heading" level={2}>
              Your install is live
            </Text>
            <Badge theme="success">Active</Badge>
          </div>
          <Text variant="body" theme="neutral">
            Everything is provisioned and ready to go.
          </Text>
        </div>
        <Button
          type="button"
          variant="primary"
          onClick={handleContinue}
        >
          View install <Icon variant="CaretRight" weight="bold" />
        </Button>
      </div>

      {/* Tabbed sections */}
      {sections.length > 0 && <Tabs tabs={tabs} />}
    </div>
  )
}
