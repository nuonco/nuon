import { LabelBadge } from './LabelBadge'

export default {
  title: 'Common/LabelBadge',
}

export const Default = () => <LabelBadge label="env:production" />

export const ExplicitKeyValue = () => (
  <LabelBadge labelKey="region" labelValue="us-east-1" />
)

export const AllThemes = () => (
  <div className="flex flex-col gap-2">
    <LabelBadge label="theme:success" theme="success" />
    <LabelBadge label="theme:brand" theme="brand" />
    <LabelBadge label="theme:default" theme="default" />
    <LabelBadge label="theme:neutral" theme="neutral" />
    <LabelBadge label="theme:warn" theme="warn" />
    <LabelBadge label="theme:error" theme="error" />
    <LabelBadge label="theme:info" theme="info" />
  </div>
)

export const Sizes = () => (
  <div className="flex flex-col gap-2">
    <LabelBadge label="size:sm" size="sm" />
    <LabelBadge label="size:md" size="md" />
    <LabelBadge label="size:lg" size="lg" />
  </div>
)

export const MultipleLabels = () => (
  <div className="flex flex-wrap gap-2">
    <LabelBadge label="env:production" theme="success" />
    <LabelBadge label="region:us-east-1" theme="info" />
    <LabelBadge label="team:platform" theme="brand" />
    <LabelBadge label="tier:critical" theme="error" />
  </div>
)

export const ColonInValue = () => (
  <LabelBadge label="url:https://example.com" theme="info" />
)

export const CustomKeyTheme = () => (
  <div className="flex flex-col gap-2">
    <LabelBadge label="env:production" keyTheme="info" theme="success" />
    <LabelBadge label="status:degraded" keyTheme="neutral" theme="warn" />
    <LabelBadge label="alert:critical" keyTheme="error" theme="error" />
  </div>
)

export const CodeVariant = () => (
  <div className="flex flex-wrap gap-2">
    <LabelBadge label="image:nginx:latest" variant="code" theme="info" />
    <LabelBadge label="sha:a1b2c3d" variant="code" />
  </div>
)
