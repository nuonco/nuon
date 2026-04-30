import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'

export const SAMPLE_PAYLOAD = `{
  "specversion": "1.0",
  "id": "3487f2f5-ec33-45db-8200-6cc669c044e3",
  "type": "com.nuon.operation.lifecycle.v1",
  "source": "//nuon.co/ctl-api",
  "time": "2026-04-28T04:42:52Z",
  "subject": "orgeovwzku7qimrb9lxhrowckg/component-deploy/plan/plan.finished",
  "datacontenttype": "application/json",
  "nuonorgid": "orgeovwzku7qimrb9lxhrowckg",
  "nuonoperation": "component-deploy",
  "nuonstage": "plan",
  "nuonstatus": "succeeded",
  "data": {
    "event": "plan.finished",
    "operation": "component-deploy",
    "stage": "plan",
    "status": "succeeded",
    "duration_ms": 12345,
    "context": {
      "org_id": "orgeovwzku7qimrb9lxhrowckg",
      "install_id": "ins98e2wpzdxwoey393edtqj45",
      "component_id": "cmp05dfarnb3wt8wvcxanqovby"
    }
  }
}`

export const SamplePayload = () => {
  return (
    <Expand
      id="webhook-sample-payload"
      className="rounded-md border"
      headerClassName="p-4"
      heading={
        <Text weight="strong">
          View sample payload
        </Text>
      }
    >
      <div className="flex flex-col gap-3 p-4 border-t">
        <div className="flex justify-end">
          <ClickToCopyButton textToCopy={SAMPLE_PAYLOAD} title="Copy payload" />
        </div>
        <CodeBlock language="json">{SAMPLE_PAYLOAD}</CodeBlock>
      </div>
    </Expand>
  )
}
