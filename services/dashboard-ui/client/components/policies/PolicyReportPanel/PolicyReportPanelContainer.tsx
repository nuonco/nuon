import { useSurfaces } from '@/hooks/use-surfaces'
import { PolicyReportPanel, PolicyReportPanelButton } from './PolicyReportPanel'
import type { TPolicyReport } from '@/types'
import type { IButtonAsButton } from '@/components/common/Button'

interface IPolicyReportPanelButtonContainer extends IButtonAsButton {
  report: TPolicyReport
  orgId: string
  policyNameMap: Map<string, string>
}

export const PolicyReportPanelButtonContainer = ({
  report,
  orgId,
  policyNameMap,
  ...props
}: IPolicyReportPanelButtonContainer) => {
  const { addPanel } = useSurfaces()

  const handleOpen = () => {
    const panel = (
      <PolicyReportPanel
        report={report}
        orgId={orgId}
        policyNameMap={policyNameMap}
      />
    )
    addPanel(panel)
  }

  return (
    <PolicyReportPanelButton
      report={report}
      orgId={orgId}
      policyNameMap={policyNameMap}
      onOpen={handleOpen}
      {...props}
    />
  )
}
