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

export const CodeValues = () => (
  <div className="flex flex-wrap gap-2">
    <LabelBadge label="image:nginx:latest" theme="info" />
    <LabelBadge label="sha:a1b2c3d" />
  </div>
)

export const CustomColors = () => (
  <div className="flex flex-wrap gap-2">
    <LabelBadge label="env:production" customColor="#16a34a" />
    <LabelBadge label="region:us-east-1" customColor="#2563eb" />
    <LabelBadge label="team:platform" customColor="#9333ea" />
    <LabelBadge label="tier:critical" customColor="#dc2626" />
  </div>
)

export const MixedCustomAndDefault = () => (
  <div className="flex flex-wrap gap-2">
    <LabelBadge label="env:production" customColor="#16a34a" size="sm" />
    <LabelBadge label="region:us-east-1" size="sm" />
    <LabelBadge label="team:platform" customColor="#9333ea" size="sm" />
  </div>
)

export const KeyOnly = () => (
  <div className="flex flex-col gap-2">
    <LabelBadge label="standalone" />
    <LabelBadge labelKey="standalone" />
  </div>
)

export const EmptyValue = () => <LabelBadge label="env:" theme="info" />

export const EmptyKey = () => <LabelBadge label=":production" theme="info" />

export const NoProps = () => <LabelBadge />

export const LongValue = () => (
  <div className="flex flex-col gap-2">
    <LabelBadge label="path:/very/long/filesystem/path/that/keeps/going/and/going/and/going/and/going/and/going/and/going" />
    <LabelBadge label="description:this is an unusually long free-form value that a user might paste in without thinking about width, so it should truncate and reveal the full value on hover" />
  </div>
)

export const LongKey = () => (
  <div className="flex flex-col gap-2">
    <LabelBadge label="this-is-an-unusually-long-label-key-that-nobody-should-ever-use-but-somehow-someone-will:value" />
  </div>
)

export const LongWithRemove = () => (
  <LabelBadge
    label="description:this is an unusually long free-form value that a user might paste in without thinking about width"
    theme="info"
    onRemove={() => {}}
  />
)

export const WithRemove = () => (
  <div className="flex flex-wrap gap-2">
    <LabelBadge label="env:production" theme="success" onRemove={() => {}} />
    <LabelBadge label="region:us-east-1" customColor="#2563eb" onRemove={() => {}} />
    <LabelBadge label="size:sm" size="sm" onRemove={() => {}} />
  </div>
)

export const WithRemoveDisabled = () => (
  <LabelBadge label="env:production" theme="success" disabled onRemove={() => {}} />
)
