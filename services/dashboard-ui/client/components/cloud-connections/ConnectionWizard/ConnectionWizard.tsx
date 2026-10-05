import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { DetailPage } from '@/components/layout/DetailPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { cn } from '@/utils/classnames'

const STEPS = ['Cloud and account', 'Access', 'Run in your cloud', 'Verify']
export const ConnectionWizard = ({
  step,
  created,
  onStep,
  children,
}: {
  step: number
  created: boolean
  onStep: (step: number) => void
  children: ReactNode
}) => (
  <DetailPage
    header={
      <SectionHeader
        title={created ? 'Set up cloud connection' : 'Create cloud connection'}
        description="Let Nuon manage resources in your AWS account through a role you control."
        status={<Badge theme="info">AWS · OIDC</Badge>}
      />
    }
  >
    <div className="grid gap-8 lg:grid-cols-[15rem_minmax(0,1fr)]">
      <nav
        aria-label="Cloud connection steps"
        className="flex gap-2 overflow-x-auto lg:flex-col"
      >
        {STEPS.map((label, index) => {
          const number = index + 1
          const disabled = created ? number < 3 : number > step
          return (
            <Button
              key={label}
              variant="ghost"
              className={cn(
                'min-w-fit !justify-start !p-3 !shadow-none',
                step === number && '!bg-primary-50 dark:!bg-primary-950'
              )}
              aria-current={step === number ? 'step' : undefined}
              disabled={disabled}
              tooltipProps={
                disabled
                  ? {
                      tipContent: created
                        ? 'Connection details are saved and cannot be changed'
                        : 'Complete the current step first',
                    }
                  : undefined
              }
              onClick={() => onStep(number)}
            >
              <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border">
                {number < step ? (
                  <Icon variant="CheckIcon" size={14} theme="success" />
                ) : (
                  number
                )}
              </span>
              <Text
                as="span"
                variant="subtext"
                weight={step === number ? 'strong' : undefined}
              >
                {label}
              </Text>
            </Button>
          )
        })}
      </nav>
      <div className="min-w-0">{children}</div>
    </div>
  </DetailPage>
)
