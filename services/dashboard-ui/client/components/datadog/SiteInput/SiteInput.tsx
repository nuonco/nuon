import { useEffect, useMemo, useState } from 'react'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'

// KNOWN_SITES mirrors the regional keys accepted on the backend
// (services/ctl-api/internal/app/datadog_connection.go). The dropdown
// shows the host as a sub-label so users picking from US/EU regions can
// double-check they're targeting the right tenant before saving.
export const KNOWN_SITES: { key: string; label: string; host: string }[] = [
  { key: 'us1', label: 'US1 (Datadog default)', host: 'datadoghq.com' },
  { key: 'us3', label: 'US3', host: 'us3.datadoghq.com' },
  { key: 'us5', label: 'US5', host: 'us5.datadoghq.com' },
  { key: 'eu1', label: 'EU1', host: 'datadoghq.eu' },
  { key: 'ap1', label: 'AP1', host: 'ap1.datadoghq.com' },
  { key: 'gov', label: 'US1-Gov', host: 'ddog-gov.com' },
]

const CUSTOM_VALUE = '__custom__'

// SiteInput accepts a region key (us1/us3/...) OR a full https URL, just
// like the backend's validateDatadogSite. We render a dropdown for the
// common case AND a free-form input that activates when the user picks
// "Custom URL" — this matches the dashboard UX agreed on in the spec.
//
// Single onChange contract: emits the canonical string value the backend
// expects (region key or full URL). The parent never has to think about
// "is this a key or a URL" — the input handles the dichotomy.
export const SiteInput = ({
  value,
  onChange,
  disabled,
}: {
  value: string
  onChange: (next: string) => void
  disabled?: boolean
}) => {
  const knownKeys = useMemo(() => new Set(KNOWN_SITES.map((s) => s.key)), [])
  const initialMode = value && !knownKeys.has(value) ? 'custom' : 'known'

  const [mode, setMode] = useState<'known' | 'custom'>(initialMode)
  const [customUrl, setCustomUrl] = useState<string>(
    initialMode === 'custom' ? value : ''
  )

  // Sync the local custom-URL buffer if the parent resets `value`
  // (e.g., the modal switching between create/edit instances).
  useEffect(() => {
    if (!knownKeys.has(value)) {
      setMode('custom')
      setCustomUrl(value || '')
    } else {
      setMode('known')
    }
  }, [value, knownKeys])

  const dropdownValue = mode === 'custom' ? CUSTOM_VALUE : value || 'us1'

  const options = [
    ...KNOWN_SITES.map((s) => ({
      value: s.key,
      label: `${s.label} — ${s.host}`,
    })),
    { value: CUSTOM_VALUE, label: 'Custom URL (self-hosted Datadog)' },
  ]

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-2">
        <Label htmlFor="dd-site-region">Region</Label>
        <Select
          id="dd-site-region"
          options={options}
          value={dropdownValue}
          disabled={disabled}
          onChange={(e) => {
            const next = e.target.value
            if (next === CUSTOM_VALUE) {
              setMode('custom')
              onChange(customUrl)
              return
            }
            setMode('known')
            onChange(next)
          }}
        />
        <Text variant="subtext" theme="neutral">
          Pick the regional tenant your team uses. The Nuon → Datadog hook
          will route events to that tenant's API host.
        </Text>
      </div>

      {mode === 'custom' ? (
        <div className="flex flex-col gap-2">
          <Label htmlFor="dd-site-url">Custom URL</Label>
          <Input
            id="dd-site-url"
            type="url"
            placeholder="https://datadog.internal.example.com"
            value={customUrl}
            disabled={disabled}
            onChange={(e) => {
              const next = e.target.value
              setCustomUrl(next)
              onChange(next.trim())
            }}
          />
          <Text variant="subtext" theme="neutral">
            For self-hosted or private Datadog. Must be an https URL with no
            path — e.g. <code>https://datadog.example.com</code>.
          </Text>
        </div>
      ) : null}
    </div>
  )
}
