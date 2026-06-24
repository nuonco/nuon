import { useState } from 'react'
import { Text } from '@/components/common/Text'
import { DiffMarker } from '../../shared/icons'
import { diffRowBg } from '../../shared/format'
import type { DiffSectionData } from './lib'

type DiffSummary = { added?: number; removed?: number; changed?: number }

interface IConfigStep {
  appConfigId?: string
  status?: string
  sections: DiffSectionData[]
  summary: DiffSummary | null
  diffResolved: boolean
  metadata: Record<string, any>
}

export const ConfigStep = ({ appConfigId, status, sections, summary, diffResolved, metadata }: IConfigStep) => {
  if (!appConfigId) {
    return (
      <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
        <Text variant="subtext" theme="neutral">
          {status === 'in-progress' ? 'Cloning repository and parsing configuration...' : 'Waiting to fetch app configuration...'}
        </Text>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Header: config file + summary counts */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div className="flex items-center gap-3">
          <span className="font-mono text-[12.5px] px-2.5 py-1 rounded-[6px] border border-cool-grey-300 dark:border-dark-grey-600 bg-cool-grey-50 dark:bg-dark-grey-800 text-cool-grey-700 dark:text-cool-grey-200">
            nuon.toml
          </span>
          {summary && (
            <>
              {(summary.added ?? 0) > 0 && (
                <span className="text-[13px] text-green-600 dark:text-green-400">
                  <span className="font-semibold">{summary.added}</span> additions
                </span>
              )}
              {(summary.removed ?? 0) > 0 && (
                <span className="text-[13px] text-red-600 dark:text-red-400">
                  <span className="font-semibold">{summary.removed}</span> removals
                </span>
              )}
              {(summary.changed ?? 0) > 0 && (
                <span className="text-[13px] text-yellow-600 dark:text-yellow-400">
                  <span className="font-semibold">{summary.changed}</span> changed
                </span>
              )}
            </>
          )}
        </div>
      </div>

      {/* Diff sections from API */}
      {sections.map((section, i) => (
        <ConfigDiffSectionView key={section.name || i} section={section} />
      ))}

      {/* Loading / error / no-changes fallback */}
      {sections.length === 0 && appConfigId && (
        <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700">
          <Text variant="subtext" theme="neutral">
            {diffResolved
              ? metadata.component_count !== undefined
                ? `Synced ${metadata.component_count} components${metadata.action_count ? `, ${metadata.action_count} actions` : ''}`
                : 'No changes detected'
              : 'Loading diff...'}
          </Text>
        </div>
      )}
    </div>
  )
}

const ConfigDiffSectionView = ({ section }: { section: DiffSectionData }) => {
  const [expanded, setExpanded] = useState(true)

  return (
    <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-[10px] overflow-hidden">
      {/* Section header */}
      <button
        className="flex items-center justify-between w-full px-4 py-2.5 bg-cool-grey-100/70 dark:bg-dark-grey-800 border-b border-cool-grey-200 dark:border-dark-grey-700 hover:bg-cool-grey-100 dark:hover:bg-dark-grey-750 transition-colors text-left"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-2">
          <svg
            width="12" height="12" viewBox="0 0 12 12" fill="none"
            className={`text-cool-grey-400 shrink-0 transition-transform ${expanded ? 'rotate-90' : ''}`}
          >
            <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          <span className="text-[13px] font-semibold text-cool-grey-900 dark:text-white">{section.name}</span>
        </div>
        <div className="flex items-center gap-2">
          {section.additions > 0 && (
            <span className="text-[12px] font-semibold text-green-600 dark:text-green-400">+{section.additions}</span>
          )}
          {section.removals > 0 && (
            <span className="text-[12px] font-semibold text-red-600 dark:text-red-400">−{section.removals}</span>
          )}
          {section.changed > 0 && (
            <span className="text-[12px] font-semibold text-yellow-600 dark:text-yellow-400">~{section.changed}</span>
          )}
        </div>
      </button>
      {/* Entries */}
      {expanded && (
        <div className="divide-y divide-cool-grey-100 dark:divide-dark-grey-800">
          {section.entries.map((entry, j) => (
            <div key={`${entry.name}-${j}`} className={`flex items-center gap-3 px-4 py-2.5 ${diffRowBg(entry.op)}`}>
              <DiffMarker op={entry.op} />
              <span className="font-mono text-[12.5px] font-semibold text-cool-grey-900 dark:text-white">{entry.name}</span>
              <span className="font-mono text-[12px] text-cool-grey-500 dark:text-cool-grey-400 truncate">{entry.description}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
