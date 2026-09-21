import { useState, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export type TSectionNavItem = {
  id: string
  label: ReactNode
}

export type TSectionNavSection = {
  id: string
  label: string
  items?: TSectionNavItem[]
  render: (itemId?: string) => ReactNode
}

interface ISectionNav {
  sections: TSectionNavSection[]
  initSectionId?: string
  initItemId?: string
  ariaLabel: string
}

export const SectionNav = ({
  sections,
  initSectionId,
  initItemId,
  ariaLabel,
}: ISectionNav) => {
  const initialSection =
    sections.find((section) => section.id === initSectionId) ?? sections.at(0)
  const [selectedSectionId, setSelectedSectionId] = useState(initialSection?.id)
  const [selectedItems, setSelectedItems] = useState<Record<string, string>>(
    initialSection && initItemId
      ? { [initialSection.id]: initItemId }
      : initialSection?.items?.at(0)
        ? { [initialSection.id]: initialSection.items[0].id }
        : {}
  )
  const selectedSection =
    sections.find((section) => section.id === selectedSectionId) ??
    sections.at(0)

  if (!selectedSection) return null

  const selectSection = (section: TSectionNavSection) => {
    setSelectedSectionId(section.id)
    if (section.items?.length && !selectedItems[section.id]) {
      setSelectedItems((current) => ({
        ...current,
        [section.id]: section.items?.[0]?.id ?? '',
      }))
    }
  }

  const selectedItemId =
    selectedSection.items?.find(
      (item) => item.id === selectedItems[selectedSection.id]
    )?.id ?? selectedSection.items?.at(0)?.id

  return (
    <div className="grid grid-cols-1 md:grid-cols-[15rem_minmax(0,1fr)] min-h-0">
      <nav
        className="p-3 border-b md:border-b-0 md:border-r"
        aria-label={ariaLabel}
      >
        <div
          className="flex flex-col gap-1"
          role="tablist"
          aria-orientation="vertical"
        >
          {sections.map((section) => {
            const isSelected = section.id === selectedSection.id
            const hasItems = Boolean(section.items)

            return (
              <div key={section.id} className="flex flex-col gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  isActive={isSelected}
                  className="justify-start w-full"
                  onClick={() => selectSection(section)}
                  role="tab"
                  aria-selected={isSelected}
                  aria-expanded={hasItems ? isSelected : undefined}
                >
                  <span className="flex items-center justify-between gap-2 w-full">
                    <Text as="span" variant="subtext" weight="strong">
                      {section.label}
                    </Text>
                    {hasItems && (
                      <Icon
                        variant="CaretDownIcon"
                        size={12}
                        className={cn(
                          'shrink-0 text-cool-grey-400 transition-transform duration-fast ease-cubic',
                          !isSelected && '-rotate-90'
                        )}
                      />
                    )}
                  </span>
                </Button>

                {hasItems && isSelected && (
                  <div
                    className="flex flex-col gap-1 ml-3 pl-2 border-l"
                    aria-label={section.label}
                  >
                    {section.items?.map((item) => {
                      const isItemSelected = item.id === selectedItemId

                      return (
                        <Button
                          key={item.id}
                          variant="ghost"
                          size="sm"
                          isActive={isItemSelected}
                          className="justify-start w-full"
                          onClick={() =>
                            setSelectedItems((current) => ({
                              ...current,
                              [section.id]: item.id,
                            }))
                          }
                          role="tab"
                          aria-selected={isItemSelected}
                        >
                          {item.label}
                        </Button>
                      )
                    })}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </nav>

      <div
        className="min-w-0"
        role="tabpanel"
        aria-label={selectedSection.label}
      >
        {selectedSection.render(selectedItemId)}
      </div>
    </div>
  )
}
