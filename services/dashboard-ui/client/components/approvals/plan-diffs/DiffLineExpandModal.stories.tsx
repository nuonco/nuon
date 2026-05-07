export default {
  title: 'Approvals/PlanDiffs/DiffLineExpandModal',
}

import { DiffLineExpandButton } from './DiffLineExpandModal'

const DiffRow = ({
  prefix,
  style,
  label,
  beforeDisplay,
  afterDisplay,
  before,
  after,
}: {
  prefix: string
  style: string
  label: string
  beforeDisplay: string
  afterDisplay: string
  before: unknown
  after: unknown
}) => (
  <div className={`flex items-center whitespace-pre ${style} p-1`}>
    <span className="inline-block w-[2ch] shrink-0 select-none text-right mr-2 opacity-70">
      {prefix}
    </span>
    <span>
      <span className="font-semibold">{label}:</span>
      {'  '}
      <span className="text-red-800 dark:text-red-400 line-through opacity-70 inline-block max-w-[300px] truncate align-bottom">
        {beforeDisplay}
      </span>
      <span className="opacity-50">{' -> '}</span>
      <span className="inline-block max-w-[300px] truncate align-bottom">
        {afterDisplay}
      </span>
      <DiffLineExpandButton label={label} prefix={prefix as '~' | '+' | '-'} before={before} after={after} />
    </span>
  </div>
)

const changedStyle =
  'bg-orange-500/15 dark:bg-orange-500/5 text-orange-800 dark:text-orange-400'

export const LongStringValue = () => (
  <div className="p-4 bg-code font-mono text-[13px] leading-6">
    <DiffRow
      prefix="~"
      style={changedStyle}
      label="allowed_origins"
      beforeDisplay="https://app.example.com,https://staging.example.com,https://dev.example.com"
      afterDisplay="https://app.example.com,https://staging.example.com,https://dev.example.com,https://preview.example.com,https://canary.example.com"
      before="https://app.example.com,https://staging.example.com,https://dev.example.com"
      after="https://app.example.com,https://staging.example.com,https://dev.example.com,https://preview.example.com,https://canary.example.com"
    />
  </div>
)

export const JsonPolicyValue = () => (
  <div className="p-4 bg-code font-mono text-[13px] leading-6">
    <DiffRow
      prefix="~"
      style={changedStyle}
      label="policy"
      beforeDisplay='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:Get*"],"Resource":"*"}]}'
      afterDisplay='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:Get*","s3:Put*","s3:List*"],"Resource":"arn:aws:s3:::my-bucket/*"},{"Effect":"Allow","Action":["kms:Decrypt"],"Resource":"*"}]}'
      before='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:Get*"],"Resource":"*"}]}'
      after='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:Get*","s3:Put*","s3:List*"],"Resource":"arn:aws:s3:::my-bucket/*"},{"Effect":"Allow","Action":["kms:Decrypt"],"Resource":"*"}]}'
    />
  </div>
)

export const YamlConfigValue = () => (
  <div className="p-4 bg-code font-mono text-[13px] leading-6">
    <DiffRow
      prefix="~"
      style={changedStyle}
      label="nginx_conf"
      beforeDisplay="upstream backend { server backend-1:8080; server backend-2:8080; }"
      afterDisplay="upstream backend { server backend-1:8080; server backend-2:8080; server backend-3:8080; } server { listen 80; listen 443 ssl; }"
      before="upstream backend { server backend-1:8080; server backend-2:8080; }"
      after="upstream backend { server backend-1:8080; server backend-2:8080; server backend-3:8080; } server { listen 80; listen 443 ssl; }"
    />
  </div>
)

export const ObjectValue = () => (
  <div className="p-4 bg-code font-mono text-[13px] leading-6">
    <DiffRow
      prefix="~"
      style={changedStyle}
      label="tags"
      beforeDisplay='{"Environment":"staging","Team":"platform","ManagedBy":"terraform"}'
      afterDisplay='{"Environment":"production","Team":"platform","ManagedBy":"terraform","CostCenter":"eng-123","DataClassification":"internal"}'
      before={{
        Environment: 'staging',
        Team: 'platform',
        ManagedBy: 'terraform',
      }}
      after={{
        Environment: 'production',
        Team: 'platform',
        ManagedBy: 'terraform',
        CostCenter: 'eng-123',
        DataClassification: 'internal',
      }}
    />
  </div>
)

const yamlBodyBefore = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: restate-cloud-ingress
  namespace: restate-cloud-ingress
spec:
  replicas: 2
  selector:
    matchLabels:
      app: restate-cloud-ingress
  template:
    spec:
      containers:
      - name: ingress
        image: registry.example.com/restate-cloud-ingress:v1.2.0
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 256Mi`

const yamlBodyAfter = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: restate-cloud-ingress
  namespace: restate-cloud-ingress
spec:
  replicas: 3
  selector:
    matchLabels:
      app: restate-cloud-ingress
  template:
    spec:
      containers:
      - name: ingress
        image: registry.example.com/restate-cloud-ingress:v1.3.0
        ports:
        - containerPort: 8080
        - containerPort: 9090
          name: metrics
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: "1"
            memory: 512Mi
        env:
        - name: OTEL_ENABLED
          value: "true"`

export const YamlBody = () => (
  <div className="p-4 bg-code font-mono text-[13px] leading-6">
    <DiffRow
      prefix="~"
      style={changedStyle}
      label="yaml_body"
      beforeDisplay={`"apiVersion": "apps/v1" "kind": "Depl…`}
      afterDisplay={`"apiVersion": "apps/v1" "kind": "Depl…`}
      before={yamlBodyBefore}
      after={yamlBodyAfter}
    />
  </div>
)
