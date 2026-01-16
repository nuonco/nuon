import React from 'react'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'

export interface PolicyViolation {
  policy_id: string
  message: string
  severity?: 'deny' | 'warn'
}

interface PolicyViolationsListProps {
  violations: PolicyViolation[]
}

export function PolicyViolationsList({
  violations,
}: PolicyViolationsListProps) {
  if (!violations || violations.length === 0) {
    return null
  }

  const denyViolations = violations.filter((v) => v.severity === 'deny')
  const warnViolations = violations.filter(
    (v) => v.severity === 'warn' || !v.severity
  )

  return (
    <div className="flex flex-col gap-2 w-full">
      {denyViolations.length > 0 && (
        <Banner theme="error">
          <div className="flex flex-col gap-2 w-full">
            <Text weight="strong">
              Policy Violations ({denyViolations.length})
            </Text>
            <ul className="list-disc list-inside space-y-1">
              {denyViolations.map((violation, index) => (
                <li key={`deny-${violation.policy_id}-${index}`}>
                  <Text variant="subtext">
                    {violation.message || 'Policy check failed'}
                  </Text>
                </li>
              ))}
            </ul>
          </div>
        </Banner>
      )}
      {warnViolations.length > 0 && (
        <Banner theme="warn">
          <div className="flex flex-col gap-2 w-full">
            <Text weight="strong">
              Policy Warnings ({warnViolations.length})
            </Text>
            <ul className="list-disc list-inside space-y-1">
              {warnViolations.map((violation, index) => (
                <li key={`warn-${violation.policy_id}-${index}`}>
                  <Text variant="subtext">
                    {violation.message || 'Policy warning'}
                  </Text>
                </li>
              ))}
            </ul>
          </div>
        </Banner>
      )}
    </div>
  )
}
