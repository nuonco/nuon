import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'
import type { TIconVariant } from '@/components/common/Icon'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'

const NEXT_STEP_CARDS: { icon: TIconVariant; title: string; description: string }[] = [
  {
    icon: 'SquaresFour',
    title: 'Explore your app',
    description: 'View your app config, builds, and deployment history.',
  },
  {
    icon: 'Users',
    title: 'Invite your team',
    description: 'Add teammates and set up roles for your organization.',
  },
  {
    icon: 'BookOpen',
    title: 'Read the docs',
    description: 'Learn more about Nuon workflows and advanced configuration.',
  },
]

export const NextStepsStep = ({ onAdvance }: IWizardStepComponentProps) => (
  <div className="flex flex-col gap-6">
    <div className="flex flex-col items-center gap-3 py-4">
      <Icon variant="CheckCircle" size={48} weight="fill" />
      <Text variant="heading" className="text-center">
        You're all set!
      </Text>
      <Text variant="body" theme="neutral" className="text-center max-w-sm">
        Your environment is ready. Here's what you can do next.
      </Text>
    </div>

    <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
      {NEXT_STEP_CARDS.map((card) => (
        <Card key={card.title} className="gap-3 p-4">
          <Icon variant={card.icon} size={20} />
          <Text variant="label">{card.title}</Text>
          <Text variant="body" theme="neutral">
            {card.description}
          </Text>
        </Card>
      ))}
    </div>

    <div className="flex self-end">
      <Button type="button" variant="primary" onClick={onAdvance}>
        Go to dashboard <Icon variant="ArrowRight" weight="bold" />
      </Button>
    </div>
  </div>
)
