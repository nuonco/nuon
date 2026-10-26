import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { WorkflowSteps } from '@/components/workflows/WorkflowSteps'
import { useWorkflow } from '@/hooks/use-workflow'
import { WorkflowProvider } from '@/providers/workflow-provider'
import type { TAPIError } from '@/types'
import { getStatusTheme } from '@/utils/status-utils'

interface ITeardownProgress {
  installName: string
  installHref: string
  isForgetting: boolean
  forgetError?: TAPIError | null
  onForget: () => void
}

const TeardownProgress = ({
  installName,
  installHref,
  isForgetting,
  forgetError,
  onForget,
}: ITeardownProgress) => {
  const { workflow } = useWorkflow()
  const status = workflow.status?.status ?? ''
  const theme = getStatusTheme(status)
  const finished =
    !!workflow.finished ||
    theme === 'success' ||
    theme === 'error' ||
    theme === 'warn'
  const succeeded = theme === 'success'

  return (
    <div className="flex flex-col gap-6">
      <FormErrorBanner
        error={forgetError}
        fallback="Unable to forget install"
      />
      {finished ? (
        <Banner
          theme={succeeded ? 'success' : theme === 'warn' ? 'warn' : 'error'}
        >
          {succeeded
            ? `${installName} is removed from the cloud account. Forget the install to remove it from Nuon.`
            : `Teardown ${theme === 'warn' ? 'finished with warnings' : 'failed'}.`}
        </Banner>
      ) : (
        <Text theme="neutral" flex className="gap-2">
          <Icon variant="Loading" />
          Tearing down {installName}
        </Text>
      )}
      <WorkflowSteps />
      <div className="flex flex-wrap justify-end gap-3">
        <Button type="button" variant="secondary" href={installHref}>
          Go to install
        </Button>
        <Button
          type="button"
          variant="danger"
          disabled={!succeeded || isForgetting}
          onClick={onForget}
          tooltipProps={
            succeeded
              ? undefined
              : { tipContent: 'Wait for teardown to complete successfully' }
          }
        >
          {isForgetting ? (
            <>
              <Icon variant="Loading" /> Forgetting...
            </>
          ) : (
            <>
              <Icon variant="TrashIcon" /> Forget install
            </>
          )}
        </Button>
      </div>
    </div>
  )
}

export const TeardownStep = ({
  workflowId,
  ...props
}: ITeardownProgress & { workflowId: string }) => (
  <WorkflowProvider workflowId={workflowId} shouldPoll>
    <TeardownProgress {...props} />
  </WorkflowProvider>
)
