import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getAppLabels, updateApp } from '@/lib'
import type { TAppLabelKey } from '@/lib/ctl-api/apps/get-app-labels'
import type { TAPIError } from '@/types/dashboard.types'

const SWATCH_COLORS = [
  '#2563eb', '#dc2626', '#16a34a', '#9333ea', '#ca8a04', '#0891b2',
  '#e11d48', '#4f46e5', '#059669', '#c026d3', '#d97706', '#0284c7',
  '#7c3aed', '#15803d', '#a21caf', '#b45309', '#6366f1', '#ef4444',
  '#22c55e', '#a855f7', '#eab308', '#06b6d4', '#f43f5e', '#818cf8',
]

function ColorSwatch({ color, active, onClick }: { color: string; active?: boolean; onClick?: () => void }) {
  return (
    <button
      type="button"
      className={`w-5 h-5 rounded-sm border-2 cursor-pointer hover:scale-110 transition-transform ${active ? 'border-dark-grey-950 dark:border-white scale-110' : 'border-cool-grey-300 dark:border-dark-grey-500'}`}
      style={{ backgroundColor: color }}
      onClick={onClick}
      title={color}
    />
  )
}

function LabelRow({
  label,
  overrides,
  onOverride,
  onRemoveOverride,
}: {
  label: TAppLabelKey
  overrides: Record<string, string>
  onOverride: (key: string, color: string) => void
  onRemoveOverride: (key: string) => void
}) {
  const [showPicker, setShowPicker] = useState(false)

  return (
    <div className="flex flex-col gap-2 border-b border-cool-grey-200 dark:border-dark-grey-700 pb-4">
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-4 min-w-0">
          <LabelBadge
            labelKey={label.key}
            labelValue={label.values?.[0] ?? ''}
            customColor={label.color}
            size="sm"
            variant="code"
          />
          <span className="flex flex-wrap gap-1">
            {label.entity_types?.sort().map((et) => (
              <Badge key={et} size="sm" theme="info">{et}</Badge>
            ))}
          </span>
          <Text variant="subtext" theme="neutral">{label.usage_count} use{label.usage_count !== 1 ? 's' : ''}</Text>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          {label.is_override ? (
            <>
              <Badge size="sm" theme="brand">override</Badge>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onRemoveOverride(label.key)}
                title="Remove override and use default color"
              >
                <Icon variant="ArrowCounterClockwiseIcon" size="14" />
              </Button>
            </>
          ) : (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowPicker((v) => !v)}
            >
              <Icon variant="PencilSimpleIcon" size="14" />
              Override
            </Button>
          )}
        </div>
      </div>

      <div className="flex items-center gap-3 pl-1">
        <span className="flex items-center gap-2">
          <span
            className="inline-block w-3.5 h-3.5 rounded-sm border border-cool-grey-300 dark:border-dark-grey-500"
            style={{ backgroundColor: label.default_color }}
          />
          <Text variant="subtext" theme="neutral" className="font-mono">{label.default_color}</Text>
          <Text variant="subtext" theme="neutral">default</Text>
        </span>
        {label.is_override && (
          <>
            <Icon variant="ArrowRightIcon" size="12" theme="neutral" />
            <span className="flex items-center gap-2">
              <span
                className="inline-block w-3.5 h-3.5 rounded-sm border border-cool-grey-300 dark:border-dark-grey-500"
                style={{ backgroundColor: label.color }}
              />
              <Text variant="subtext" className="font-mono">{label.color}</Text>
            </span>
          </>
        )}
      </div>

      {label.values?.length > 1 && (
        <div className="flex items-center gap-2 pl-1">
          <Text variant="subtext" theme="neutral">Values:</Text>
          <span className="flex flex-wrap gap-1">
            {label.values.sort().map((v) => (
              <Badge key={v} size="sm" theme="default">{v}</Badge>
            ))}
          </span>
        </div>
      )}

      {showPicker && (
        <div className="flex flex-col gap-2 pl-1 pt-1">
          <Text variant="label" weight="strong">Pick a color</Text>
          <div className="flex flex-wrap gap-1.5">
            {SWATCH_COLORS.map((color) => (
              <ColorSwatch
                key={color}
                color={color}
                onClick={() => {
                  onOverride(label.key, color)
                  setShowPicker(false)
                }}
              />
            ))}
            <label className="flex items-center gap-1 cursor-pointer" title="Custom color">
              <input
                type="color"
                className="w-5 h-5 rounded-sm border border-cool-grey-300 dark:border-dark-grey-500 cursor-pointer"
                defaultValue={label.default_color}
                onChange={(e) => {
                  onOverride(label.key, e.target.value)
                  setShowPicker(false)
                }}
              />
            </label>
          </div>
        </div>
      )}
    </div>
  )
}

export const Labels = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['app-labels', org?.id, app?.id],
    queryFn: () => getAppLabels({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const { mutate: saveLabelColors, isPending } = useMutation({
    mutationFn: (labelColors: Record<string, string>) =>
      updateApp({
        appId: app.id,
        orgId: org.id,
        body: { label_colors: labelColors },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app-labels', org?.id, app?.id] })
      queryClient.invalidateQueries({ queryKey: ['app', org?.id, app?.id] })
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Update failed" theme="error">
          <Text>{err?.error || 'Unable to update label colors.'}</Text>
        </Toast>
      )
    },
  })

  const overrides = data?.label_colors ?? {}

  const handleOverride = (key: string, color: string) => {
    saveLabelColors({ ...overrides, [key]: color })
  }

  const handleRemoveOverride = (key: string) => {
    const next = { ...overrides }
    delete next[key]
    saveLabelColors(next)
  }

  const handleResetAll = () => {
    saveLabelColors({})
  }

  const labels = data?.labels ?? []
  const hasOverrides = labels.some((l) => l.is_override)

  return (
    <PageSection>
      <PageTitle title={`Labels | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/labels`, text: 'Labels' },
        ]}
      />

      <HeadingGroup>
        <div className="flex items-center justify-between">
          <Text variant="base" weight="strong">
            Labels
          </Text>
          {hasOverrides && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleResetAll}
              disabled={isPending}
            >
              <Icon variant="ArrowCounterClockwiseIcon" size="16" />
              Reset all to defaults
            </Button>
          )}
        </div>
        <Text variant="subtext" theme="neutral">
          All label keys used across components, actions, runbooks, and installs. Each key gets a default color automatically. Override any color by clicking the override button.
        </Text>
      </HeadingGroup>

      {isLoading ? (
        <Card className="flex flex-col gap-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} height="48px" width="100%" />
          ))}
        </Card>
      ) : labels.length === 0 ? (
        <EmptyState
          variant="diagram"
          emptyTitle="No labels yet"
          emptyMessage="Add labels to your components, actions, runbooks, or installs to see them here."
        />
      ) : (
        <Card className="flex flex-col gap-4 !p-6">
          {labels.map((label) => (
            <LabelRow
              key={label.key}
              label={label}
              overrides={overrides}
              onOverride={handleOverride}
              onRemoveOverride={handleRemoveOverride}
            />
          ))}
        </Card>
      )}
    </PageSection>
  )
}
