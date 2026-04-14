import { Banner } from '@/components/common/Banner'
import { Badge } from '@/components/common/Badge'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import type { TCompositeError, TCompositeErrorOwnerType } from '@/types'

const OWNER_LABELS: Record<TCompositeErrorOwnerType, string> = {
  plan: 'Plan',
  apply: 'Apply',
  'action-run': 'Action run',
  'variable-renderer': 'Variable renderer',
  runner: 'Runner',
  'k8s-diagnostics': 'K8s diagnostics',
}

interface ICompositeErrorBanner {
  errors?: TCompositeError[] | null
}

export const CompositeErrorBanner = ({ errors }: ICompositeErrorBanner) => {
  if (!errors?.length) return null

  if (errors.length === 1) {
    return <SingleErrorBanner error={errors[0]} />
  }

  return <MultipleErrorsBanner errors={errors} />
}

const SingleErrorBanner = ({ error }: { error: TCompositeError }) => {
  return (
    <Banner theme="error">
      <div className="flex flex-col gap-2 w-full">
        <div className="flex items-center gap-2">
          <Text weight="strong">{error.summary}</Text>
          <Badge theme="error" size="sm">
            {OWNER_LABELS[error.owner_type]}
          </Badge>
        </div>
        {error.detail ? <ErrorDetail detail={error.detail} /> : null}
      </div>
    </Banner>
  )
}

const MultipleErrorsBanner = ({ errors }: { errors: TCompositeError[] }) => {
  return (
    <Banner theme="error">
      <div className="flex flex-col gap-3 w-full">
        <Text weight="strong">
          {errors.length} error{errors.length !== 1 ? 's' : ''} occurred
        </Text>
        <div className="flex flex-col gap-1">
          {errors.map((error, index) => (
            <Expand
              key={`${error.owner_id ?? error.owner_type}-${index}`}
              id={`composite-error-${index}`}
              heading={
                <div className="flex items-center gap-2">
                  <Badge theme="error" size="sm">
                    {OWNER_LABELS[error.owner_type]}
                  </Badge>
                  <Text variant="body">{error.summary}</Text>
                </div>
              }
              headerClassName="rounded"
            >
              {error.detail ? (
                <div className="pt-1 pb-2 px-2">
                  <ErrorDetail detail={error.detail} />
                </div>
              ) : null}
            </Expand>
          ))}
        </div>
      </div>
    </Banner>
  )
}

const ErrorDetail = ({ detail }: { detail: string }) => {
  return (
    <pre className="font-mono text-xs leading-5 max-h-64 overflow-y-auto bg-black/10 dark:bg-white/5 rounded p-3 whitespace-pre-wrap break-words">
      {detail}
    </pre>
  )
}
