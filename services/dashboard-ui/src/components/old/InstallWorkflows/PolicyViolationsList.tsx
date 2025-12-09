import React from 'react'
import { Notice } from '@/components/old/Notice'
import { Text } from '@/components/old/Typography'

export interface PolicyViolation {
  policy_id: string
  message: string
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

  return (
    <Notice variant="error" className="!p-4 w-full">
      <Text variant="med-14" className="mb-2">
        Policy Violations ({violations.length})
      </Text>
      <ul className="list-disc list-inside space-y-2">
        {violations.map((violation, index) => (
          <li key={`${violation.policy_id}-${index}`} className="text-sm">
            <Text as="span">{violation.message || 'Policy check failed'}</Text>
          </li>
        ))}
      </ul>
    </Notice>
  )
}
