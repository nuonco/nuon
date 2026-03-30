import { useMutation } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TIconVariant } from '@/components/common/Icon'
import type { TOnboarding } from '@/types'
import { completeGetStartedStep } from '@/lib'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

type NextStepLink = {
  icon: TIconVariant
  title: string
  description: string
  href: string
  badge?: string
}

type NextStepSection = {
  icon: TIconVariant
  title: string
  step: number
  description: string
  links: NextStepLink[]
}

const NEXT_STEP_SECTIONS: NextStepSection[] = [
  {
    icon: 'Plugs',
    title: 'Connect your app',
    step: 1,
    description:
      'Wire your SaaS into Nuon — define install inputs, connect your CI/CD pipeline, and configure component sources.',
    links: [
      {
        icon: 'BookOpen',
        title: 'App inputs & config',
        description: 'Define what customers provide at install time',
        href: 'https://docs.nuon.co',
      },
      {
        icon: 'BookOpen',
        title: 'Connect CI/CD',
        description: 'Trigger deploys from GitHub Actions, etc.',
        href: 'https://docs.nuon.co',
      },
      {
        icon: 'BookOpen',
        title: 'Component sources',
        description: 'Helm, Terraform, Docker, raw manifests',
        href: 'https://docs.nuon.co',
      },
      {
        icon: 'BookOpen',
        title: 'nuon apps init',
        description: 'Pull your app config locally with the CLI',
        href: 'https://docs.nuon.co',
      },
    ],
  },
  {
    icon: 'ShieldCheck',
    title: 'Day 2 operations',
    step: 2,
    description:
      'Keep customer installs healthy, secure, and auditable after go-live.',
    links: [
      {
        icon: 'BookOpen',
        title: 'Drift detection',
        description: 'Detect & auto-reconcile config drift across installs',
        href: 'https://docs.nuon.co',
        badge: 'Reliability',
      },
      {
        icon: 'BookOpen',
        title: 'Break-glass access',
        description: 'Audited emergency access to customer infrastructure',
        href: 'https://docs.nuon.co',
        badge: 'Security',
      },
      {
        icon: 'BookOpen',
        title: 'Policies',
        description: 'OPA rules that gate installs and deployments',
        href: 'https://docs.nuon.co',
        badge: 'Governance',
      },
      {
        icon: 'BookOpen',
        title: 'Approvals',
        description: 'Require human sign-off before risky changes apply',
        href: 'https://docs.nuon.co',
        badge: 'Automation',
      },
    ],
  },
  {
    icon: 'Browser',
    title: 'Create an installer',
    step: 3,
    description:
      'Give customers a self-service portal to install, configure, and manage their own instance.',
    links: [
      {
        icon: 'BookOpen',
        title: 'Installer overview',
        description: 'What the customer-facing portal looks like',
        href: 'https://docs.nuon.co',
      },
      {
        icon: 'BookOpen',
        title: 'Portal branding',
        description: 'Custom domain, logo, colours, and copy',
        href: 'https://docs.nuon.co',
      },
      {
        icon: 'BookOpen',
        title: 'Customer inputs',
        description: 'What you ask customers to provide before install',
        href: 'https://docs.nuon.co',
      },
      {
        icon: 'BookOpen',
        title: 'Embed the installer',
        description: 'Drop the portal into your app or docs site',
        href: 'https://docs.nuon.co',
      },
    ],
  },
]

export const NextStepsStep = ({ onAdvance, sharedData }: IWizardStepComponentProps) => {
  const onboarding = sharedData.onboarding as TOnboarding | undefined
  const orgId = onboarding?.org_id

  const { mutate: completeStep } = useMutation({
    mutationFn: () => completeGetStartedStep({ orgId: orgId! }),
  })

  const installId = onboarding?.install_id

  const handleOpenControlPlane = () => {
    if (orgId) completeStep()
    if (orgId && installId) {
      window.location.href = `/${orgId}/installs/${installId}`
    } else {
      onAdvance()
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-2">
        <Text variant="h2">What's next</Text>
        <Text variant="body" theme="neutral">
          Your environment is provisioned. Here are some ways to get the most out
          of Nuon.
        </Text>
      </div>

      {NEXT_STEP_SECTIONS.map((section) => (
        <div key={section.title} className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <Icon variant={section.icon} size={20} />
              <Text variant="label">{section.title}</Text>
              <Badge size="sm" theme="neutral">
                Step {section.step}
              </Badge>
            </div>
            <Text variant="body" theme="neutral">
              {section.description}
            </Text>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {section.links.map((link) => (
              <a
                key={link.title}
                href={link.href}
                target="_blank"
                rel="noopener noreferrer"
                className="no-underline"
              >
                <Card className="gap-2 p-4 h-full hover:border-cool-grey-500 transition-colors relative">
                  {link.badge && (
                    <Badge
                      size="sm"
                      theme="default"
                      className="absolute top-3 right-3"
                    >
                      {link.badge}
                    </Badge>
                  )}
                  <div className="flex items-center gap-2">
                    <Icon variant={link.icon} size={16} />
                    <Text variant="label">{link.title}</Text>
                  </div>
                  <Text variant="body" theme="neutral">
                    {link.description}
                  </Text>
                </Card>
              </a>
            ))}
          </div>
        </div>
      ))}

      <Button
        type="button"
        variant="primary"
        onClick={handleOpenControlPlane}
        className="self-center"
      >
        <Icon variant="SquaresFour" /> Open control plane
      </Button>
    </div>
  )
}
