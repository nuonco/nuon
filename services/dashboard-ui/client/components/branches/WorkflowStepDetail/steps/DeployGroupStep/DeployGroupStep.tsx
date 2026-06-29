import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TInstallWorkflowStep } from '@/types'
import { InstallStatusIcon } from '../../shared/icons'

interface IDeployGroupStep {
  step: TInstallWorkflowStep
  metadata: Record<string, any>
}

export const DeployGroupStep = ({ step, metadata }: IDeployGroupStep) => {
  const groupName = step.name?.replace(/^deploy install group:\s*/i, '') || 'unknown'
  const installs = (metadata.installs as any[]) || []
  const totalInstalls = installs.length || (metadata.install_count as number) || 0
  const deployedCount = installs.filter((i: any) => i.status === 'success' || i.status === 'deployed').length
  const currentActivity = metadata.current_activity as string | undefined
  const showActivity = currentActivity || (step.status?.status === 'in-progress' && step.status?.status_human_description)
  const activityText = currentActivity || step.status?.status_human_description

  return (
    <div className="space-y-3">
      {/* Deploy head row */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Icon variant="PackageIcon" size={16} className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
          <span className="text-[13px] text-cool-grey-600 dark:text-cool-grey-300">
            install group:{' '}
            <span className="font-semibold text-cool-grey-900 dark:text-white">{groupName}</span>
          </span>
          <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">
            {totalInstalls} {totalInstalls === 1 ? 'install' : 'installs'}
          </span>
        </div>
        {totalInstalls > 0 && (
          <span className="font-mono text-[12px] text-cool-grey-500 dark:text-cool-grey-400">
            {deployedCount} / {totalInstalls} deployed
          </span>
        )}
      </div>

      {/* Activity bar */}
      {showActivity && activityText && (
        <div
          className="flex items-center gap-3 px-4 py-3 rounded-[10px] border"
          style={{
            background: 'rgba(63,116,224,0.07)',
            borderColor: 'rgba(63,116,224,0.32)',
          }}
        >
          <div className="w-[18px] h-[18px] rounded-full bg-blue-500 flex items-center justify-center shrink-0">
            <svg className="animate-spin" width="12" height="12" viewBox="0 0 12 12" fill="none">
              <circle cx="6" cy="6" r="4.5" stroke="rgba(255,255,255,0.3)" strokeWidth="1.5" />
              <path d="M6 1.5 A4.5 4.5 0 0 1 10.5 6" stroke="white" strokeWidth="1.5" strokeLinecap="round" />
            </svg>
          </div>
          <span className="font-mono text-[12.5px] text-blue-700 dark:text-blue-300 flex-1 truncate">
            {activityText}
          </span>
          <div className="w-[120px] h-[6px] rounded-full bg-blue-100 dark:bg-blue-900/40 overflow-hidden shrink-0">
            <div className="h-full bg-blue-500 rounded-full transition-all" style={{ width: '40%' }} />
          </div>
        </div>
      )}

      {/* Install list */}
      {installs.length > 0 && (
        <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-[10px] divide-y divide-cool-grey-100 dark:divide-dark-grey-800 overflow-hidden">
          {installs.map((inst: any, i: number) => {
            const instStatus = inst.status || 'pending'
            const isInstInProgress = instStatus === 'in-progress'
            const isPending = instStatus === 'pending'

            return (
              <div
                key={inst.install_id || i}
                className={`px-4 py-3 transition-colors ${isInstInProgress ? 'bg-blue-50/60 dark:bg-[rgba(63,116,224,0.06)]' : ''
                  } ${isPending ? 'opacity-[0.62]' : ''}`}
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <InstallStatusIcon status={instStatus} />
                    <span className="text-[14px] font-semibold text-cool-grey-900 dark:text-white truncate">
                      {inst.install_name || inst.install_id}
                    </span>
                    {inst.region && (
                      <div className="flex items-center gap-1 shrink-0">
                        <Icon variant="GlobeIcon" size={12} className="text-cool-grey-400 dark:text-cool-grey-500" />
                        <span className="text-[12px] text-cool-grey-400 dark:text-cool-grey-500">{inst.region}</span>
                      </div>
                    )}
                    {inst.version && (
                      <span className="text-[11.5px] font-mono px-1.5 py-0.5 rounded-[6px] border border-cool-grey-200 dark:border-dark-grey-700 bg-cool-grey-50 dark:bg-dark-grey-800 text-cool-grey-500 dark:text-cool-grey-400 shrink-0">
                        {inst.version}
                      </span>
                    )}
                  </div>
                  <span className="font-mono text-[12.5px] text-cool-grey-400 dark:text-cool-grey-500 shrink-0">
                    {isPending ? '—' : (inst.duration || '')}
                  </span>
                </div>
                {isInstInProgress && (
                  <div className="flex items-center gap-3 mt-2 pl-[26px]">
                    <div className="w-[180px] h-[5px] rounded-full bg-cool-grey-200 dark:bg-dark-grey-700 overflow-hidden shrink-0">
                      <div className="h-full bg-blue-500 rounded-full transition-all" style={{ width: `${inst.progress || 30}%` }} />
                    </div>
                    {inst.activity && (
                      <span className="text-[11.5px] text-cool-grey-500 dark:text-cool-grey-400 truncate">
                        {inst.activity}
                      </span>
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      {installs.length === 0 && step.status?.status === 'in-progress' && !activityText && (
        <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
          <Text variant="subtext" theme="neutral">Deploying to install group...</Text>
        </div>
      )}
    </div>
  )
}
