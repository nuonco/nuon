import { Text } from '@/components/common/Text'

const NAV_GROUPS = [
  { title: 12, items: [18, 14] },
  { title: 10, items: [22, 16, 20] },
  { title: 14, items: [16, 24] },
]

const SECTIONS = [
  { title: 18, lines: [36, 28, 42] },
  { title: 14, lines: [30, 22] },
  { title: 20, lines: [40, 34, 26] },
]

export const ConfigChangesLoading = () => (
  <div className="flex flex-col gap-6 md:min-h-0 md:flex-1">
    <header className="flex shrink-0 flex-wrap items-center justify-between gap-3">
      <span className="flex flex-wrap items-center gap-x-4 gap-y-1">
        <Text variant="subtext" family="mono" loading loadingWidth={8} />
        <Text variant="subtext" family="mono" loading loadingWidth={16} />
      </span>
      <Text variant="subtext" loading loadingWidth={24} />
    </header>

    <div className="grid gap-6 md:min-h-0 md:flex-1 md:grid-cols-[15rem_minmax(0,1fr)]">
      <div className="flex flex-col gap-4">
        {NAV_GROUPS.map((group, index) => (
          <div key={index} className="flex flex-col gap-2">
            <div className="px-2 py-1.5">
              <Text variant="subtext" loading loadingWidth={group.title} />
            </div>
            <div className="flex flex-col gap-2 pl-8">
              {group.items.map((width, itemIndex) => (
                <Text
                  key={itemIndex}
                  variant="subtext"
                  loading
                  loadingWidth={width}
                />
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="flex min-w-0 flex-col gap-3">
        <div className="flex items-center gap-2 pb-2">
          <Text loading loadingWidth={40} />
        </div>
        {SECTIONS.map((section, index) => (
          <div
            key={index}
            className="flex flex-col gap-3 rounded-r-md border-l-4 border-l-cool-grey-500/30 bg-cool-grey-500/5 p-3"
          >
            <Text weight="strong" loading loadingWidth={section.title} />
            <div className="flex flex-col gap-2 rounded-md bg-white p-3 dark:bg-dark-grey-800">
              {section.lines.map((width, lineIndex) => (
                <Text
                  key={lineIndex}
                  variant="subtext"
                  family="mono"
                  loading
                  loadingWidth={width}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  </div>
)
