import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import { DetailStatusIcon } from '../../shared/icons'
import { cacheBadgeTheme } from '../../shared/format'

interface IBuildStep {
  metadata: Record<string, any>
  status?: string
}

export const BuildStep = ({ metadata, status }: IBuildStep) => {
  const builds = (metadata.builds as any[]) || []
  const [expandedId, setExpandedId] = useState<string | null>(null)

  if (builds.length === 0) {
    return (
      <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
        <Text variant="subtext" theme="neutral">
          {status === 'in-progress' ? 'Starting component builds...' : 'Waiting to start builds...'}
        </Text>
      </div>
    )
  }

  const succeededCount = builds.filter((b: any) => b.status === 'success' || b.status === 'skipped').length
  const totalDuration = builds.reduce((acc: number, b: any) => acc + (b.duration || 0), 0)

  return (
    <div className="space-y-3">
      {/* Summary row */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-[13px] text-cool-grey-700 dark:text-cool-grey-200">
            <span className="font-semibold">{builds.length}</span> components built
          </span>
          <span className="text-[12px] text-cool-grey-400">·</span>
          <span className="text-[13px] font-semibold text-green-600 dark:text-green-400">
            {succeededCount} succeeded
          </span>
        </div>
        {totalDuration > 0 && (
          <span className="font-mono text-[12px] text-cool-grey-500 dark:text-cool-grey-400">
            {totalDuration.toFixed(1)}s total
          </span>
        )}
      </div>

      {/* Component build rows */}
      <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-[10px] divide-y divide-cool-grey-100 dark:divide-dark-grey-800 overflow-hidden">
        {builds.map((build: any, i: number) => {
          const buildStatus = build.status || 'pending'
          const isExpanded = expandedId === (build.component_id || i)

          return (
            <div key={build.component_id || i}>
              <button
                className="flex items-center gap-3 px-4 py-3 w-full text-left hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800 transition-colors"
                onClick={() => setExpandedId(isExpanded ? null : (build.component_id || i))}
              >
                <svg
                  width="12" height="12" viewBox="0 0 12 12" fill="none"
                  className={`text-cool-grey-400 shrink-0 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                >
                  <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                </svg>

                <DetailStatusIcon status={buildStatus === 'skipped' ? 'success' : buildStatus} />

                <span className="text-[13.5px] font-semibold text-cool-grey-900 dark:text-white">
                  {build.component_name || build.component_id}
                </span>

                {build.cache_status && (
                  <Badge theme={cacheBadgeTheme(build.cache_status)} size="sm">
                    {build.cache_status}
                  </Badge>
                )}

                <div className="flex-1" />

                {build.image_digest && (
                  <span className="font-mono text-[11.5px] text-cool-grey-400 dark:text-cool-grey-500 shrink-0">
                    {build.image_digest.length > 20 ? build.image_digest.substring(0, 20) : build.image_digest}
                  </span>
                )}

                {build.duration && (
                  <span className="font-mono text-[12.5px] text-cool-grey-500 dark:text-cool-grey-400 shrink-0 ml-2">
                    {Number(build.duration).toFixed(1)}s
                  </span>
                )}
              </button>
            </div>
          )
        })}
      </div>
    </div>
  )
}
