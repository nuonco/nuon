import { LabelColorPicker } from './LabelColorPicker'

export default {
  title: 'Apps/AppLabels/LabelColorPicker',
}

const noop = () => {}

export const Default = () => (
  <LabelColorPicker
    id="story-picker"
    value="#2563eb"
    defaultColor="#2563eb"
    onSelect={noop}
  />
)

export const Overridden = () => (
  <LabelColorPicker
    id="story-picker-override"
    value="#a21caf"
    defaultColor="#16a34a"
    isOverride
    onSelect={noop}
    onReset={noop}
  />
)
