import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'

export interface PolicyViolation {
  policy_id: string
  message: string
}

interface IPolicyViolations {
  step: TWorkflowStep
}

export const PolicyViolations = ({ step }: IPolicyViolations) => {
  const violations = step?.status?.metadata?.policy_violations as
    | PolicyViolation[]
    | undefined

  if (!violations || violations.length === 0) {
    return null
  }

  return (
    <Banner theme="error">
      <div className="flex flex-col gap-2 w-full">
        <Text weight="strong">Policy Violations ({violations.length})</Text>
        <ul className="list-disc list-inside space-y-1">
          {violations.map((violation, index) => (
            <li key={`${violation.policy_id}-${index}`}>
              <Text as="span" variant="subtext">
                {violation.message || 'Policy check failed'}
              </Text>
            </li>
          ))}
        </ul>
      </div>
    </Banner>
  )
}
