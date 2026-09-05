import { useForm } from '@tanstack/react-form'
import { allEvents } from './defaults'
import { FormInterestsPicker } from './FormInterestsPicker'
import { PRESETS } from './presets'
import type { Interests } from './types'
import type { SubscriptionMatch } from '@/components/match/types'

export default { title: 'Interests/FormInterestsPicker' }

const Demo = ({
  initialInterests,
  initialMatch,
}: {
  initialInterests?: Interests
  initialMatch?: SubscriptionMatch
}) => {
  const form = useForm({
    defaultValues: {
      interests: initialInterests ?? allEvents(),
      match: initialMatch as SubscriptionMatch | undefined,
    },
  })
  return (
    <div className="max-w-lg p-4">
      <form.Field name="interests">
        {(field) => (
          <form.Field name="match">
            {(matchField) => (
              <FormInterestsPicker
                field={field}
                matchField={matchField}
              />
            )}
          </form.Field>
        )}
      </form.Field>
    </div>
  )
}

export const Default = () => <Demo />

export const FailuresWithLabels = () => (
  <Demo
    initialInterests={PRESETS[0].build()}
    initialMatch={{
      installs: {
        selector: { match_labels: { env: 'prod', tier: 'critical' } },
      },
    }}
  />
)

export const CustomWithScope = () => (
  <Demo
    initialInterests={{ resources: { installs: { outcome: 'failures' } } }}
    initialMatch={{
      installs: { selector: { match_labels: { env: 'prod' } } },
    }}
  />
)
