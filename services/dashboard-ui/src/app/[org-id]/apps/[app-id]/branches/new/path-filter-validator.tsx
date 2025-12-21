'use client'

import { useState, useMemo } from 'react'
import { Input } from '@/components/old/Input'
import { Text } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { Badge } from '@/components/common/Badge'
import { Dropdown } from '@/components/old/Dropdown'
import { RadioInput } from '@/components/old/Input'

interface IPathFilterValidatorProps {
  value: string
  onChange: (value: string) => void
  disabled?: boolean
}

const COMMON_PATTERNS = [
  {
    label: 'All files',
    pattern: '',
    description: 'Match all file changes',
  },
  {
    label: 'Source code only',
    pattern: '^(src/|lib/)',
    description: 'Only trigger on changes in src/ or lib/ directories',
  },
  {
    label: 'Ignore docs',
    pattern: '^(?!docs/)',
    description: 'Trigger on all changes except in docs/ directory',
  },
  {
    label: 'Specific file types',
    pattern: '\\.(ts|tsx|js|jsx)$',
    description: 'Only trigger on TypeScript and JavaScript files',
  },
  {
    label: 'Monorepo package',
    pattern: '^packages/frontend/',
    description: 'Only trigger on changes in a specific package',
  },
  {
    label: 'Config files',
    pattern: '^(config/|.env|\\.yaml$|\\.json$)',
    description: 'Only trigger on configuration file changes',
  },
]

const SAMPLE_PATHS = [
  'src/components/Button.tsx',
  'src/utils/helpers.ts',
  'lib/api/client.ts',
  'docs/README.md',
  'docs/guides/getting-started.md',
  'config/app.yaml',
  'config/database.json',
  '.env.production',
  'packages/frontend/src/App.tsx',
  'packages/backend/src/server.ts',
  'tests/unit/button.test.ts',
  'package.json',
  'Dockerfile',
]

export const PathFilterValidator = ({
  value,
  onChange,
  disabled = false,
}: IPathFilterValidatorProps) => {
  const [isValid, setIsValid] = useState(true)

  // Test the regex pattern
  const testResults = useMemo(() => {
    if (!value) {
      return SAMPLE_PATHS.map((path) => ({ path, matches: true }))
    }

    try {
      const regex = new RegExp(value)
      setIsValid(true)
      return SAMPLE_PATHS.map((path) => ({
        path,
        matches: regex.test(path),
      }))
    } catch (err) {
      setIsValid(false)
      return SAMPLE_PATHS.map((path) => ({ path, matches: false }))
    }
  }, [value])

  const matchCount = testResults.filter((r) => r.matches).length

  const handlePatternSelect = (pattern: string) => {
    onChange(pattern)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault()
    }
  }

  return (
    <div className="space-y-4">
      {/* Pattern Input */}
      <div className="space-y-2">
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1">
            <label className="block">
              <Text variant="sm" weight="strong">
                Path Filter (regex)
              </Text>
              <Text
                variant="xs"
                className="text-cool-grey-600 dark:text-cool-grey-400 mb-2"
              >
                Only trigger builds when files matching this pattern change
              </Text>
            </label>
            <Input
              type="text"
              value={value}
              onChange={(e) => onChange(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="^(src/|config/)"
              disabled={disabled}
            />
          </div>

          {/* Common Patterns Dropdown */}
          <div className="pt-6">
            <Dropdown
              id="common-patterns"
              text="Examples"
              variant="secondary"
            >
              {COMMON_PATTERNS.map((pattern, index) => (
                <button
                  key={index}
                  onClick={() => handlePatternSelect(pattern.pattern)}
                  className="w-full text-left px-4 py-2 hover:bg-cool-grey-100 dark:hover:bg-dark-grey-700"
                >
                  <Text variant="sm" weight="strong">
                    {pattern.label}
                  </Text>
                  <Text
                    variant="xs"
                    className="text-cool-grey-600 dark:text-cool-grey-400 mt-1"
                  >
                    {pattern.description}
                  </Text>
                  {pattern.pattern && (
                    <code className="text-xs bg-cool-grey-100 dark:bg-dark-grey-700 px-2 py-0.5 rounded mt-1 inline-block">
                      {pattern.pattern}
                    </code>
                  )}
                </button>
              ))}
            </Dropdown>
          </div>
        </div>

        {/* Validation Status */}
        {!isValid && (
          <div className="flex items-center gap-2 p-2 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-900 rounded">
            <Icon variant="Warning" size={16} className="text-red-600" />
            <Text variant="xs" className="text-red-700 dark:text-red-400">
              Invalid regular expression syntax
            </Text>
          </div>
        )}

        {isValid && value && (
          <div className="flex items-center gap-2 p-2 bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-900 rounded">
            <Icon variant="Check" size={16} className="text-green-600" />
            <Text variant="xs" className="text-green-700 dark:text-green-400">
              Valid pattern - matching {matchCount} of {SAMPLE_PATHS.length} sample paths
            </Text>
          </div>
        )}
      </div>

      {/* Pattern Tester */}
      <div className="border rounded-lg p-4 bg-cool-grey-50 dark:bg-dark-grey-800">
        <div className="flex items-center gap-2 mb-3">
          <Icon variant="Flask" size={16} className="text-primary-600" />
          <Text variant="sm" weight="strong">
            Pattern Tester
          </Text>
        </div>

        <Text
          variant="xs"
          className="text-cool-grey-600 dark:text-cool-grey-400 mb-3"
        >
          Test your pattern against common file paths. Green indicates a match (build will trigger), grey indicates no match (build will be skipped).
        </Text>

        <div className="grid grid-cols-1 gap-1 max-h-64 overflow-y-auto">
          {testResults.map((result, index) => (
            <div
              key={index}
              className="flex items-center gap-2 px-2 py-1 rounded hover:bg-white dark:hover:bg-dark-grey-900"
            >
              <Icon
                variant={result.matches ? 'CheckCircle' : 'Circle'}
                size={14}
                className={
                  result.matches
                    ? 'text-green-600'
                    : 'text-cool-grey-400 dark:text-cool-grey-600'
                }
              />
              <Text
                variant="xs"
                className={
                  result.matches
                    ? 'text-cool-grey-900 dark:text-cool-grey-100 font-mono'
                    : 'text-cool-grey-500 dark:text-cool-grey-500 font-mono'
                }
              >
                {result.path}
              </Text>
            </div>
          ))}
        </div>
      </div>

      {/* Help Text */}
      <div className="flex items-start gap-2 p-3 bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-900 rounded">
        <Icon variant="Info" size={16} className="text-blue-600 mt-0.5" />
        <div className="flex-1">
          <Text variant="xs" className="text-blue-800 dark:text-blue-300">
            <strong>Vercel-style monorepo support:</strong> Use path filters to trigger builds only when specific packages or directories change. Leave empty to trigger on any file change.
          </Text>
        </div>
      </div>
    </div>
  )
}