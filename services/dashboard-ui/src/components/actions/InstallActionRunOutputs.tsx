'use client'

import { Duration } from '@/components/common/Duration'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { hydrateActionRunSteps } from '@/utils/action-utils'

export const InstallActionRunOutputs = () => {
  const { installActionRun } = useInstallActionRun()
  const steps = hydrateActionRunSteps({
    steps: installActionRun?.steps,
    stepConfigs: installActionRun?.config?.steps,
  })

  return (
    <div className="flex flex-col gap-4">
      {steps
        ?.sort((a, b) => {
          if (a.idx === undefined && b.idx === undefined) return 0
          if (a.idx === undefined) return -1
          if (b.idx === undefined) return 1
          return a.idx - b.idx
        })
        .map((step) => {
          const outputs = installActionRun?.outputs?.steps?.[step?.name]
          const outputCount = Object.keys(outputs)?.length
          const quickRef = extractQuickReference(outputs)

          return outputs ? (
            <Expand
              className="rounded-md border"
              id={`${step?.id}-outputs`}
              key={step?.id}
              heading={
                <div className="flex flex-col gap-1 items-start">
                  <Text
                    className="!flex items-center gap-4"
                    variant="base"
                    weight="strong"
                  >
                    {step?.name} outputs{' '}
                  </Text>
                  <div className="flex items-center gap-6">
                    <Status status={step?.status} />

                    <Text
                      className="!flex items-center gap-1"
                      variant="subtext"
                    >
                      <Icon variant="Package" />
                      {outputCount} item{outputCount > 1 ? 's' : null}
                    </Text>
                    {step?.execution_duration ? (
                      <Text className="!flex items-center gap-1">
                        <Icon variant="Timer" />
                        <Duration
                          nanoseconds={step?.execution_duration}
                          variant="subtext"
                        />
                      </Text>
                    ) : null}
                  </div>
                </div>
              }
            >
              <div className="flex flex-col gap-6 p-4 border-t">
                {quickRef?.length ? (
                  <div className="flex flex-col gap-2">
                    <Text weight="strong">Quick reference</Text>
                    <KeyValueList values={quickRef} />
                  </div>
                ) : null}
                <div className="flex flex-col gap-2">
                  <Text className="!flex items-center gap-4" weight="strong">
                    Outputs JSON{' '}
                    <Text
                      className="!flex items-center gap-1"
                      variant="label"
                      theme="neutral"
                    >
                      <Icon variant="FloppyDisk" size="13" />{' '}
                      {getOutputSize(outputs)}
                    </Text>
                  </Text>
                  <JSONViewer
                    className="border-none !rounded-t-none"
                    data={outputs}
                    expanded={3}
                  />
                </div>
              </div>
            </Expand>
          ) : (
            <div className="flex flex-col items-start">
              <Text className="!flex items-center gap-2" weight="strong">
                <Status status={step?.status} isWithoutText />
                No outputs for {step?.name} step
              </Text>
              {step?.execution_duration ? (
                <Text variant="subtext" theme="neutral">
                  Finished in{' '}
                  <Duration
                    nanoseconds={step?.execution_duration}
                    variant="subtext"
                  />
                </Text>
              ) : null}
            </div>
          )
        })}
    </div>
  )
}

function extractQuickReference(output, maxItems = 6) {
  const candidates = []

  function traverse(obj, path = []) {
    for (const [key, value] of Object.entries(obj)) {
      const fullPath = [...path, key].join('.')

      // Score each field
      const score = scoreField(key, value, path.length)

      if (score > 0 && isPrimitive(value)) {
        candidates.push({
          path: fullPath,
          key,
          value,
          score,
          depth: path.length,
        })
      }

      // Don't go too deep or into arrays
      if (
        typeof value === 'object' &&
        !Array.isArray(value) &&
        path.length < 2
      ) {
        traverse(value, [...path, key])
      }
    }
  }

  traverse(output)

  // Sort by score and return top N
  return candidates.sort((a, b) => b.score - a.score).slice(0, maxItems)
}

function scoreField(key, value, depth) {
  let score = 100 // Base score

  // Penalize depth
  score -= depth * 30

  // Boost important-sounding names (generic patterns)
  const keyLower = key.toLowerCase()
  if (keyLower.endsWith('arn') || keyLower.endsWith('id')) score += 50
  if (keyLower.includes('name')) score += 40
  if (keyLower.includes('status') || keyLower.includes('state')) score += 45
  if (
    keyLower.includes('url') ||
    keyLower.includes('endpoint') ||
    keyLower.includes('dns')
  )
    score += 40
  if (keyLower.includes('ip') || keyLower.includes('address')) score += 35
  if (keyLower.includes('port')) score += 30
  if (keyLower.includes('zone') || keyLower.includes('region')) score += 25

  // Penalize verbose/metadata fields
  if (keyLower.includes('description')) score -= 20
  if (keyLower.includes('created') || keyLower.includes('updated')) score -= 15
  if (keyLower.includes('enable') || keyLower.includes('prefix')) score -= 25

  // Penalize very long strings (likely not useful for quick ref)
  if (typeof value === 'string' && value.length > 150) score -= 30

  // Boost short, meaningful values
  if (typeof value === 'string' && value.length < 50) score += 10

  return score
}

function isPrimitive(value) {
  return typeof value !== 'object' || value === null
}

function formatBytes(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
  if (bytes < 1024 * 1024 * 1024)
    return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

function getOutputSize(output) {
  const jsonString = JSON.stringify(output)
  const bytes = new TextEncoder().encode(jsonString).length
  return formatBytes(bytes)
}
