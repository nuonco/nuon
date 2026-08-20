import { Banner } from '@/components/common/Banner'

export const RunnerStatusBanner = ({
  warnings,
  status,
}: {
  warnings?: string[]
  status?: string
}) => {
  if (status === 'disabled') {
    return (
      <Banner theme="neutral">
        This runner is disabled, so it runs no jobs and reports no health.
        Re-enable it in the install stack to start running jobs.
      </Banner>
    )
  }

  if (!warnings?.length) return null

  return (
    <div className="flex flex-col gap-2">
      {warnings.map((warning, i) => (
        <Banner key={i} theme="warn">
          {warning}
        </Banner>
      ))}
    </div>
  )
}
