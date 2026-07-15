import { Banner } from '@/components/common/Banner'

export const RunnerStatusBanner = ({ warnings }: { warnings?: string[] }) => {
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
