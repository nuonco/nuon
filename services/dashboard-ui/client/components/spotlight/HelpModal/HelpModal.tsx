import { useMemo } from 'react'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Badge } from '@/components/common/Badge'
import { Kbd } from '@/components/common/Kbd'
import { toSentenceCase } from '@/utils/string-utils'
import { COMMANDS_BY_PREFIX, COMMAND_DESCRIPTIONS, FILTER_PREFIXES } from '../types'

const Combo = ({ children }: { children: React.ReactNode }) => (
  <span className="flex items-center gap-1">{children}</span>
)

const isMac = () =>
  typeof navigator !== 'undefined' &&
  /Mac|iPhone|iPad|iPod/.test(navigator.userAgent)

interface IShortcut {
  label: string
  keys: React.ReactNode
}

interface ISection {
  title: string
  shortcuts: IShortcut[]
}

const ShortcutRow = ({ label, keys }: IShortcut) => (
  <div className="flex items-center justify-between gap-4 py-1.5">
    <Text variant="body">{label}</Text>
    <div className="shrink-0">{keys}</div>
  </div>
)

const Section = ({ title, shortcuts }: ISection) => (
  <section>
    <div className="mb-2">
      <Text variant="base" weight="stronger">
        {title}
      </Text>
    </div>
    <div className="flex flex-col">
      {shortcuts.map((s, i) => (
        <ShortcutRow key={`${title}-${i}`} {...s} />
      ))}
    </div>
  </section>
)

export const HelpModal = ({ ...props }: IModal) => {
  const mac = useMemo(isMac, [])
  const mod = mac ? '⌘' : 'Ctrl'
  const alt = mac ? '⌥' : 'Alt'
  const shift = mac ? '⇧' : 'Shift'

  const sections: ISection[] = useMemo(
    () => [
      {
        title: 'General',
        shortcuts: [
          {
            label: 'Open spotlight search',
            keys: (
              <Combo>
                <Kbd>{mod}</Kbd>
                <Kbd>K</Kbd>
              </Combo>
            ),
          },
          {
            label: 'Open this help',
            keys: (
              <Combo>
                <Kbd>{mod}</Kbd>
                <Kbd>H</Kbd>
              </Combo>
            ),
          },
          {
            label: 'Close modal or panel',
            keys: (
              <Combo>
                <Kbd>Esc</Kbd>
              </Combo>
            ),
          },
          {
            label: 'Toggle sidebar',
            keys: (
              <Combo>
                <Kbd>{alt}</Kbd>
                <Kbd>S</Kbd>
              </Combo>
            ),
          },
          {
            label: 'Toggle page sidebar',
            keys: (
              <Combo>
                <Kbd>{alt}</Kbd>
                <Kbd>{shift}</Kbd>
                <Kbd>S</Kbd>
              </Combo>
            ),
          },
        ],
      },
      {
        title: 'Navigation',
        shortcuts: [
          {
            label: 'Go to dashboard',
            keys: (
              <Combo>
                <Kbd>G</Kbd>
                <Kbd>D</Kbd>
              </Combo>
            ),
          },
          {
            label: 'Go to apps',
            keys: (
              <Combo>
                <Kbd>G</Kbd>
                <Kbd>A</Kbd>
              </Combo>
            ),
          },
          {
            label: 'Go to installs',
            keys: (
              <Combo>
                <Kbd>G</Kbd>
                <Kbd>I</Kbd>
              </Combo>
            ),
          },
          {
            label: 'Go to team',
            keys: (
              <Combo>
                <Kbd>G</Kbd>
                <Kbd>T</Kbd>
              </Combo>
            ),
          },
          {
            label: 'Go to build runner',
            keys: (
              <Combo>
                <Kbd>G</Kbd>
                <Kbd>R</Kbd>
              </Combo>
            ),
          },
          {
            label: 'Go to webhooks',
            keys: (
              <Combo>
                <Kbd>G</Kbd>
                <Kbd>W</Kbd>
              </Combo>
            ),
          },
        ],
      },
      {
        title: 'Logs',
        shortcuts: [
          {
            label: 'Cycle log lines (when log panel is open)',
            keys: (
              <Combo>
                <Kbd>↑</Kbd>
                <Kbd>↓</Kbd>
                <Text variant="subtext" theme="neutral" className="text-xs px-1">
                  or
                </Text>
                <Kbd>K</Kbd>
                <Kbd>J</Kbd>
              </Combo>
            ),
          },
        ],
      },
    ],
    [mod, alt, shift]
  )

  const commandEntries = Object.entries(COMMANDS_BY_PREFIX) as [
    string,
    string[],
  ][]

  const shortcutsTab = (
    <div className="flex flex-col gap-8 pt-4">
      {sections.map((section) => (
        <Section key={section.title} {...section} />
      ))}
    </div>
  )

  const searchTipsTab = (
    <div className="flex flex-col gap-8 pt-4">
      <section>
        <div className="mb-2">
          <Text variant="base" weight="stronger">
            Spotlight search
          </Text>
        </div>
        <div className="mb-3">
          <Text variant="subtext" theme="neutral">
            Press <Kbd>{mod}</Kbd> <Kbd>K</Kbd> to open, then type to search.
            Use a prefix to scope results, press <Kbd>Tab</Kbd> to autocomplete.
          </Text>
        </div>
        <div className="flex items-center gap-1.5 flex-wrap">
          {FILTER_PREFIXES.map((prefix) => (
            <Badge key={prefix} size="sm" variant="code" theme="neutral">
              {prefix}
            </Badge>
          ))}
        </div>
      </section>

      <section>
        <div className="mb-2">
          <Text variant="base" weight="stronger">
            Slash commands
          </Text>
        </div>
        <div className="mb-4 flex flex-col gap-2">
          <Text variant="subtext" theme="neutral">
            Type <Kbd>/</Kbd> after a filter to run commands on the matched
            resource, e.g.
          </Text>
          <Badge size="sm" variant="code" theme="neutral">
            install:my-install/deploy all components
          </Badge>
        </div>
        <div className="flex flex-col gap-5">
          {commandEntries.map(([prefix, commands]) => (
            <div key={prefix}>
              <div className="flex items-center gap-2 mb-1.5">
                <Text variant="body" weight="stronger">
                  {toSentenceCase(prefix)}
                </Text>
                <Badge size="sm" variant="code" theme="brand">
                  {prefix}:
                </Badge>
              </div>
              <div className="flex flex-col">
                {commands.map((cmd) => (
                  <ShortcutRow
                    key={cmd}
                    label={COMMAND_DESCRIPTIONS[cmd] ?? cmd}
                    keys={
                      <Badge size="sm" variant="code" theme="neutral">
                        /{cmd}
                      </Badge>
                    }
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )

  return (
    <Modal
      heading="Keyboard shortcuts & commands"
      size="lg"
      showFooter={false}
      {...props}
    >
      <Tabs
        tabs={{
          shortcuts: shortcutsTab,
          'search tips': searchTipsTab,
        }}
      />
    </Modal>
  )
}
