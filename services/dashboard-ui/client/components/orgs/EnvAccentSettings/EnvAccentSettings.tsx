import { useState, useRef, type FormEvent } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Input } from '@/components/common/form/Input'
import { Icon } from '@/components/common/Icon'
import { Select } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { EnvAccentBadge } from '@/components/installs/EnvAccentBadge'
import type { TEnvAccentColor, TEnvAccentConfig } from '@/types'

const COLOR_OPTIONS: { value: TEnvAccentColor; label: string }[] = [
  { value: 'error', label: 'Red (production)' },
  { value: 'warn', label: 'Amber (staging)' },
  { value: 'success', label: 'Green (dev)' },
  { value: 'info', label: 'Blue (qa)' },
  { value: 'brand', label: 'Violet' },
  { value: 'neutral', label: 'Neutral' },
]

interface IEnvAccentSettings {
  config: TEnvAccentConfig
  isPending: boolean
  error: unknown
  onSubmit: (next: { label_key: string; values: Record<string, TEnvAccentColor> }) => void
}

type Row = { id: number; value: string; color: TEnvAccentColor }

/**
 * Presentational form for editing the org's env_accent_config. Stores a
 * row list locally and emits a PUT-style replacement on submit so the
 * backend can validate and persist.
 */
export const EnvAccentSettings = ({
  config,
  isPending,
  error,
  onSubmit,
}: IEnvAccentSettings) => {
  const nextId = useRef(0)
  const initialRows: Row[] = Object.entries(config.values ?? {})
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([value, color]) => ({
      id: nextId.current++,
      value,
      color: color as TEnvAccentColor,
    }))

  const [labelKey, setLabelKey] = useState(config.label_key ?? 'env')
  const [rows, setRows] = useState<Row[]>(initialRows)

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    const values: Record<string, TEnvAccentColor> = {}
    for (const row of rows) {
      const trimmed = row.value.trim()
      if (trimmed) values[trimmed] = row.color
    }
    onSubmit({ label_key: labelKey.trim(), values })
  }

  return (
    <form className="flex flex-col gap-6" onSubmit={handleSubmit}>
      <Card className="flex flex-col gap-4 p-6">
        <div className="flex flex-col gap-1">
          <Text variant="base" weight="strong">
            Environment label key
          </Text>
          <Text variant="subtext" theme="neutral">
            The label key on each install that drives the accent (default <code className="font-mono">env</code>).
          </Text>
        </div>
        <Input
          name="label_key"
          value={labelKey}
          onChange={(e) => setLabelKey(e.target.value)}
          placeholder="env"
        />
      </Card>

      <Card className="flex flex-col gap-4 p-6">
        <div className="flex items-start justify-between gap-4">
          <div className="flex flex-col gap-1">
            <Text variant="base" weight="strong">
              Value to color mapping
            </Text>
            <Text variant="subtext" theme="neutral">
              Installs whose <code className="font-mono">{labelKey || 'env'}</code> label matches a value below render with the matching accent across the dashboard.
            </Text>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() =>
              setRows((r) => [
                ...r,
                { id: nextId.current++, value: '', color: 'neutral' },
              ])
            }
          >
            <Icon variant="PlusIcon" size={16} />
            Add value
          </Button>
        </div>

        {error ? (
          <Banner theme="error">
            {(error as { error?: string })?.error || 'Unable to update env accent config'}
          </Banner>
        ) : null}

        {rows.length === 0 ? (
          <Text variant="subtext">No values configured. Installs will render without an accent.</Text>
        ) : (
          <div className="flex flex-col gap-3">
            {rows.map((row) => (
              <fieldset
                key={row.id}
                className="grid grid-cols-[1fr_1fr_auto_auto] gap-3 items-end"
              >
                <label className="flex flex-col gap-1">
                  <Text variant="label">Label value</Text>
                  <Input
                    type="text"
                    placeholder="e.g. production"
                    value={row.value}
                    onChange={(e) =>
                      setRows((rs) =>
                        rs.map((r) =>
                          r.id === row.id ? { ...r, value: e.target.value } : r,
                        ),
                      )
                    }
                  />
                </label>
                <label className="flex flex-col gap-1">
                  <Text variant="label">Color</Text>
                  <Select
                    name={`color-${row.id}`}
                    value={row.color}
                    options={COLOR_OPTIONS}
                    onChange={(e) =>
                      setRows((rs) =>
                        rs.map((r) =>
                          r.id === row.id
                            ? { ...r, color: e.target.value as TEnvAccentColor }
                            : r,
                        ),
                      )
                    }
                  />
                </label>
                <div className="pb-1.5">
                  <EnvAccentBadge
                    size="md"
                    accent={{
                      value: row.value || 'preview',
                      color: row.color,
                      labelKey: labelKey || 'env',
                    }}
                  />
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() =>
                    setRows((rs) => rs.filter((r) => r.id !== row.id))
                  }
                  className="mb-1"
                >
                  <Icon variant="XIcon" size={16} />
                </Button>
              </fieldset>
            ))}
          </div>
        )}
      </Card>

      <div className="flex items-center justify-end">
        <Button type="submit" variant="primary" disabled={isPending}>
          {isPending ? (
            <span className="flex items-center gap-2">
              <Icon variant="Loading" /> Saving...
            </span>
          ) : (
            'Save changes'
          )}
        </Button>
      </div>
    </form>
  )
}
