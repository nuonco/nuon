import React from 'react'
import { Notice } from '@/components/old/Notice'
import { Text } from '@/components/old/Typography'

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
        <Notice variant="error" className="!p-4 w-full">
          <Text variant="med-14" className="mb-2">
            Policy Violations ({denyViolations.length})
          </Text>
          <ul className="list-disc list-inside space-y-2">
            {denyViolations.map((violation, index) => (
              <li
                key={`deny-${violation.policy_id}-${index}`}
                className="text-sm"
              >
                <Text as="span">
                  {violation.message || 'Policy check failed'}
                </Text>
              </li>
            ))}
          </ul>
        </Notice>
      )}
      {warnViolations.length > 0 && (
        <Notice variant="warning" className="!p-4 w-full">
          <Text variant="med-14" className="mb-2">
            Policy Warnings ({warnViolations.length})
          </Text>
          <ul className="list-disc list-inside space-y-2">
            {warnViolations.map((violation, index) => (
              <li
                key={`warn-${violation.policy_id}-${index}`}
                className="text-sm"
              >
                <Text as="span">{violation.message || 'Policy warning'}</Text>
              </li>
            ))}
          </ul>
        </Notice>
      )}
    </div>
  )
}
