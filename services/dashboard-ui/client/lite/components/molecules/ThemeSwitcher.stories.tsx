import { ThemeSwitcher } from './ThemeSwitcher'

export default {
  title: 'lite/molecules/ThemeSwitcher',
}

export const Default = () => <ThemeSwitcher />

export const OnSurface = () => (
  <div className="flex flex-col gap-4 rounded-xl border border-divider bg-surface-01 p-6">
    <p className="text-sm text-tertiary">
      The switcher sits on a raised surface, so the selected pill reads against
      it rather than the page background.
    </p>
    <ThemeSwitcher />
  </div>
)
