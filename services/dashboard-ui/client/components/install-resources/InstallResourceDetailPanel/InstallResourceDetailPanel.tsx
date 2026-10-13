import { useEffect, useId, useRef, useState } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { useCopyState } from '@/components/common/ClickToCopy'
import { Cron } from '@/components/common/Cron'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ResourceKind } from '@/components/install-resources/ResourceKind'
import { resourceSummary } from '@/components/install-resources/resource-utils'
import {
  RemovedFromAppConfigBanner,
  RemovedFromAppConfigBadge,
} from '@/components/installs/RemovedFromAppConfig'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TInstallResource } from '@/types'
import { isIdentityOnlyResource } from '@/utils/health-utils'
import { humanize } from '@/utils/string-utils'
import { isStaleObservation } from '@/utils/time-utils'
import {
  describeResource,
  parseResourceDetails,
  resourceSections,
  type TResourceSection,
} from './resource-details'

const searchPattern = (search: string) =>
  new RegExp(`(${search.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')

const HighlightedText = ({ text, search }: { text: string; search: string }) =>
  search ? (
    <>
      {text.split(searchPattern(search)).map((part, index) =>
        index % 2 ? (
          <mark
            key={index}
            className="bg-orange-100 text-orange-950 dark:bg-orange-900 dark:text-orange-100 rounded-sm"
          >
            {part}
          </mark>
        ) : (
          part
        )
      )}
    </>
  ) : (
    text
  )

const ResourceSection = ({
  section,
  search,
}: {
  section: TResourceSection
  search: string
}) => (
  <section className="flex flex-col gap-3">
    <Text
      level={3}
      variant="body"
      weight="strong"
      theme="brand"
      className="border-b pb-2"
    >
      <HighlightedText text={section.title} search={search} />
    </Text>
    {section.fields?.map(([label, value]) => (
      <LabeledValue
        key={label}
        className="sm:!flex-row items-start !gap-1 sm:!gap-4"
        label={
          <Text
            variant="body"
            family={
              ['Spec', 'Status', 'Diagnosis', 'Details'].includes(section.title)
                ? 'mono'
                : 'sans'
            }
            theme="neutral"
            className="w-32 sm:w-44 shrink-0"
          >
            <HighlightedText text={label} search={search} />
          </Text>
        }
      >
        {label === 'Last scheduled' && !search ? (
          <Time time={value} variant="body" className="min-w-0" />
        ) : label === 'Schedule' && !search ? (
          <Cron cron={value} format="both" variant="body" />
        ) : (
          <Text variant="body" family="mono" className="min-w-0 break-words">
            <HighlightedText text={value} search={search} />
          </Text>
        )}
      </LabeledValue>
    ))}
    {section.table ? (
      <div className="overflow-x-auto">
        <PropertyGrid
          align="start"
          values={section.table.rows.map((row) =>
            Object.fromEntries(
              row.map((value, index) => [String(index), value])
            )
          )}
          gridTemplate={section.table.headers
            .map((header) =>
              header === 'REASON'
                ? 'minmax(8rem, 1fr)'
                : 'minmax(max-content, 1fr)'
            )
            .join(' ')}
          columns={section.table.headers.map((header, index) => ({
            key: String(index),
            header: humanize(header),
            className: index ? '!pl-3 pr-3 min-w-0' : 'pr-3 min-w-0',
            render: (value) =>
              header === 'STATUS' && value !== 'Not reported' ? (
                <Badge variant="code" theme="default" size="md">
                  <HighlightedText text={String(value)} search={search} />
                </Badge>
              ) : (
                <Text
                  variant="body"
                  family="mono"
                  nowrap={header !== 'REASON'}
                  className="min-w-0 break-words"
                >
                  {header === 'LAST TRANSITION' &&
                  value !== 'Not reported' &&
                  !search ? (
                    <Time time={String(value)} variant="body" />
                  ) : (
                    <HighlightedText text={String(value)} search={search} />
                  )}
                </Text>
              ),
          }))}
        />
      </div>
    ) : null}
    {section.text ? (
      <Text
        as="pre"
        family="mono"
        variant="body"
        className="whitespace-pre-wrap break-words"
      >
        <HighlightedText text={section.text} search={search} />
      </Text>
    ) : null}
  </section>
)

export interface IInstallResourceDetailPanel extends IPanel {
  installResource: TInstallResource
}

export const InstallResourceDetailPanel = ({
  installResource: resource,
  ...props
}: IInstallResourceDetailPanel) => {
  const [mode, setMode] = useState<'describe' | 'json'>('describe')
  const [search, setSearch] = useState('')
  const searchRef = useRef<HTMLInputElement>(null)
  const tabRefs = useRef<
    Partial<Record<typeof mode, HTMLButtonElement | HTMLAnchorElement>>
  >({})
  const id = useId()
  const { isCopied, handleCopy } = useCopyState()
  const [, refreshFreshness] = useState(0)
  useEffect(() => {
    const interval = setInterval(
      () => refreshFreshness((tick) => tick + 1),
      30_000
    )
    return () => clearInterval(interval)
  }, [])
  const parsedDetails = parseResourceDetails(resource?.details)
  const json =
    parsedDetails !== undefined
      ? JSON.stringify(parsedDetails, null, 2)
      : undefined
  const sections = resourceSections(resource)
  const query = search.trim()
  const searchable =
    mode === 'json'
      ? json || ''
      : sections
          .map((section) =>
            [
              section.title,
              ...(section.fields?.flat() || []),
              ...(section.table?.rows.flat() || []),
              section.text || '',
            ].join('\n')
          )
          .join('\n')
  const matches = query
    ? (searchable.match(searchPattern(query)) || []).length
    : 0
  const stale =
    !isIdentityOnlyResource(resource) &&
    isStaleObservation(
      resource?.observed_at,
      resource?.stale_after_seconds || undefined
    )

  return (
    <Panel
      size="3/4"
      {...props}
      aria-label="Resource inspector"
      heading={
        <DetailHeader
          backLink={false}
          title={
            <span className="flex flex-wrap items-center gap-2">
              {resource.provider === 'probe' ||
              resource.provider === 'custom' ? (
                <Icon variant="HeartbeatIcon" theme="info" />
              ) : null}
              <ResourceKind resource={resource} compact />
              <Text
                family="mono"
                variant="base"
                weight="strong"
                className="break-all"
              >
                / {resource.namespace ? `${resource.namespace} / ` : ''}
                {resource.name || 'Resource'}
              </Text>
            </span>
          }
          description={
            <Text
              family={resource.provider === 'kubernetes' ? 'mono' : 'sans'}
              variant="subtext"
              theme="neutral"
            >
              {resource.provider === 'kubernetes'
                ? resource.api_group || 'core'
                : humanize(resource.provider)}
            </Text>
          }
        />
      }
      headerClassName="!h-auto min-h-18 py-4 gap-4"
      childrenClassName="!overflow-hidden"
      footer={
        <div className="flex flex-wrap justify-between gap-2 w-full">
          <Text variant="subtext" theme="neutral">
            Reported snapshot · Not a full manifest
          </Text>
          <Text variant="subtext" theme="neutral">
            <kbd>d</kbd> Describe ·{' '}
            {json !== undefined ? (
              <>
                <kbd>j</kbd> JSON ·{' '}
              </>
            ) : null}
            <kbd>/</kbd> Find · <kbd>Esc</kbd> Close
          </Text>
        </div>
      }
      onKeyDown={(event) => {
        props.onKeyDown?.(event)
        if (
          event.defaultPrevented ||
          event.ctrlKey ||
          event.metaKey ||
          event.altKey ||
          (event.target as HTMLElement).closest(
            'input, textarea, select, [contenteditable="true"]'
          )
        )
          return
        if (event.key === '/') {
          event.preventDefault()
          searchRef.current?.focus()
        } else if (
          event.key === 'd' ||
          (event.key === 'j' && json !== undefined)
        ) {
          event.preventDefault()
          setMode(event.key === 'd' ? 'describe' : 'json')
        }
      }}
    >
      <div className="flex shrink-0 flex-wrap items-center gap-3">
        {resource.removed_from_config ? (
          <RemovedFromAppConfigBadge kind="probe" />
        ) : null}
        {stale && !resource.removed_from_config ? (
          <Badge theme="warn" size="sm">
            Stale
          </Badge>
        ) : null}
        {resource?.observed_at ? (
          <Text variant="subtext" theme="neutral">
            Last observed{' '}
            <Time
              variant="subtext"
              time={resource.observed_at}
              format="relative"
              shouldTick
            />
          </Text>
        ) : null}
      </div>
      {resource.removed_from_config ? (
        <RemovedFromAppConfigBanner kind="probe" />
      ) : null}
      {stale ? (
        <Banner theme="warn" className="shrink-0">
          This observation is outside its freshness window. Details below are
          last-reported values, not current state.
        </Banner>
      ) : null}
      <Text family="mono" weight="strong" className="shrink-0 break-words">
        {resourceSummary(resource) ||
          (isIdentityOnlyResource(resource)
            ? 'Identity snapshot'
            : 'Status not reported')}
      </Text>
      <div className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b pb-3">
        <div role="tablist" aria-label="Resource view" className="flex gap-3">
          {(['describe', 'json'] as const)
            .filter((key) => key === 'describe' || json !== undefined)
            .map((key) => (
              <Button
                key={key}
                role="tab"
                id={`${id}-${key}`}
                aria-selected={mode === key}
                aria-controls={`${id}-output`}
                tabIndex={mode === key ? 0 : -1}
                variant="tab"
                isActive={mode === key}
                ref={(el) => {
                  if (el) tabRefs.current[key] = el
                }}
                onClick={() => setMode(key)}
                onKeyDown={(event) => {
                  if (
                    json === undefined ||
                    !['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(
                      event.key
                    )
                  )
                    return
                  event.preventDefault()
                  const next =
                    event.key === 'Home'
                      ? 'describe'
                      : event.key === 'End'
                        ? 'json'
                        : mode === 'describe'
                          ? 'json'
                          : 'describe'
                  setMode(next)
                  tabRefs.current[next]?.focus()
                }}
              >
                {key === 'describe' ? 'Describe' : 'JSON'}
              </Button>
            ))}
        </div>
        <div className="flex min-w-0 flex-wrap items-center gap-3">
          <SearchInput
            ref={searchRef}
            value={search}
            onChange={setSearch}
            aria-label="Find in resource output"
            placeholder="Find in output (/)"
            className="w-full !min-w-0"
            labelClassName="min-w-0 w-48"
          />
          {query ? (
            <Text variant="subtext" theme="neutral" aria-live="polite">
              {matches} {matches === 1 ? 'match' : 'matches'}
            </Text>
          ) : null}
          <Button
            variant="ghost"
            aria-label={mode === 'json' ? 'Copy JSON' : 'Copy description'}
            onClick={() =>
              handleCopy(
                mode === 'json' ? json || '' : describeResource(resource)
              )
            }
          >
            <Icon variant={isCopied ? 'CheckIcon' : 'CopyIcon'} />
            {isCopied ? 'Copied' : 'Copy'}
          </Button>
        </div>
      </div>
      <div
        id={`${id}-output`}
        role="tabpanel"
        aria-labelledby={`${id}-${mode}`}
        tabIndex={0}
        className="flex min-h-0 flex-auto flex-col gap-8 overflow-y-auto pb-4 focus-visible:outline-1 focus-visible:outline-primary-400/80"
      >
        {mode === 'json' ? (
          <Text
            as="pre"
            family="mono"
            variant="body"
            nowrap
            className="overflow-x-auto whitespace-pre shrink-0"
          >
            <HighlightedText text={json || ''} search={query} />
          </Text>
        ) : sections.length ? (
          sections.map((section) => (
            <ResourceSection
              key={section.title}
              section={section}
              search={query}
            />
          ))
        ) : (
          <Text theme="neutral">No additional details reported.</Text>
        )}
      </div>
    </Panel>
  )
}

export interface IInstallResourceDetailPanelButton extends IButtonAsButton {
  onOpen: () => void
}

export const InstallResourceDetailPanelButton = ({
  onOpen,
  children = 'Details',
  ...props
}: IInstallResourceDetailPanelButton) => {
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={onOpen}
      aria-label="View resource details"
      {...props}
    >
      {children}
    </Button>
  )
}
