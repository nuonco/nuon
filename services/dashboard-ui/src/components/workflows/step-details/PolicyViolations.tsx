'use client'

import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import type { TPolicyReportExportFormat } from '@/lib'
import type { TWorkflowStep } from '@/types'

export interface PolicyViolation {
  policy_id: string
  message: string
  severity: 'deny' | 'warn'
}

interface IPolicyViolations {
  step: TWorkflowStep
}

const downloadReport = (
  orgId: string,
  reportId: string,
  format: TPolicyReportExportFormat
) => {
  const url = `/api/orgs/${orgId}/policy-reports/${reportId}/export?format=${format}`
  window.open(url, '_blank')
}

export const PolicyViolations = ({ step }: IPolicyViolations) => {
  const { org } = useOrg()
  const denyViolations =
    (step?.status?.metadata?.deny_violations as PolicyViolation[]) || []
  const warnViolations =
    (step?.status?.metadata?.warn_violations as PolicyViolation[]) || []
  const violations = [...denyViolations, ...warnViolations]
  const reportId = step?.step_target_id
  const ownerType = step?.step_target_type

  if (!reportId || !ownerType || !org?.id) {
    return null
  }

  if (violations.length === 0) {
    return null
  }

  return (
    <div className="flex flex-col gap-2">
      {denyViolations.length > 0 && (
        <Banner theme="error">
          <div className="flex flex-col gap-2 w-full">
            <div className="flex items-center justify-between">
              <Text weight="strong">
                Policy Violations ({denyViolations.length})
              </Text>
              <PolicyReportExportDropdown
                orgId={org.id}
                reportId={reportId}
                ownerType={ownerType}
              />
            </div>
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
            <div className="flex items-center justify-between">
              <Text weight="strong">
                Policy Warnings ({warnViolations.length})
              </Text>
              {denyViolations.length === 0 && (
                <PolicyReportExportDropdown
                  orgId={org.id}
                  reportId={reportId}
                  ownerType={ownerType}
                />
              )}
            </div>
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

const PolicyReportExportDropdown = ({
  orgId,
  reportId,
  ownerType,
}: {
  orgId: string
  reportId: string
  ownerType: string
}) => {
  const handleExport = (format: TPolicyReportExportFormat) => {
    const params = new URLSearchParams({
      owner_type: ownerType,
      owner_id: reportId,
      limit: '1',
      offset: '0',
    })
    const url = `/api/orgs/${orgId}/policy-reports?${params.toString()}`

    const exportWindow = window.open('', '_blank')

    fetch(url)
      .then((res) => res.json())
      .then((payload) => {
        const report = payload?.data?.[0]
        if (!report?.id) {
          exportWindow?.close()
          return
        }
        const downloadUrl = `/api/orgs/${orgId}/policy-reports/${report.id}/export?format=${format}`
        if (exportWindow) {
          exportWindow.location.href = downloadUrl
          return
        }
        downloadReport(orgId, report.id, format)
      })
      .catch(() => {
        exportWindow?.close()
      })
  }

  return (
    <Dropdown
      id={`policy-report-export-${reportId}`}
      buttonText={
        <>
          <Icon variant="DownloadSimple" /> Export
        </>
      }
      alignment="right"
      variant="ghost"
    >
      <Menu>
        <Text>Export Format</Text>
        <Button onClick={() => handleExport('opa')}>
          OPA JSON
        </Button>
        <Button onClick={() => handleExport('sarif')}>
          SARIF
        </Button>
        <Button onClick={() => handleExport('pdf')}>
          PDF Report
        </Button>
      </Menu>
    </Dropdown>
  )
}
