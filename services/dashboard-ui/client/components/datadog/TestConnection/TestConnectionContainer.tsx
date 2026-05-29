import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { testDatadogConnection } from '@/lib'
import type { TAPIError, TDatadogConnection } from '@/types'

// TestConnectionButton fires the backend's `/test` action, which
// re-validates the stored API key against DD AND posts a synthetic event
// to the org's event stream. The success toast carries the event's DD
// link so the user can click straight through to their tenant and
// confirm visually — that closes the "did this actually work?" loop in
// one click, no separate "verify in DD" instructions.
export const TestConnectionButton = ({
  connection,
  ...props
}: { connection: TDatadogConnection } & Omit<IButtonAsButton, 'children'>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { addToast } = useToast()

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
      testDatadogConnection({
        orgId: org.id,
        connectionId: connection.id!,
      }),
    onSuccess: (resp) => {
      // A successful test re-asserts `verified` status on the row, so
      // refresh the list (a previously-revoked connection might flip
      // back to verified after the user fixed their key).
      queryClient.invalidateQueries({
        queryKey: ['datadog-connections', org.id],
      })
      addToast(
        <Toast heading="Test event sent to Datadog" theme="success">
          <Text>
            {resp.posted_event_url ? (
              <>
                Open it in Datadog to confirm: {' '}
                <a
                  href={resp.posted_event_url}
                  target="_blank"
                  rel="noreferrer noopener"
                  className="text-primary-600 dark:text-primary-400 hover:underline"
                >
                  view event
                </a>
              </>
            ) : (
              'Check your Datadog event stream to confirm receipt.'
            )}
          </Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Connection test failed" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Button
      variant="ghost"
      onClick={() => mutate()}
      disabled={isPending}
      {...props}
    >
      <Icon variant={isPending ? 'Loading' : 'PaperPlaneTiltIcon'} size={14} />
    </Button>
  )
}
