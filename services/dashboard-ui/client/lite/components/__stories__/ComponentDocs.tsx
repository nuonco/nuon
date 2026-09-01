import type { ReactNode } from 'react'
import { Text } from '../atoms/Text'

export type TDocProp = {
  name: string
  type: string
  default?: string
  description: string
}

export type TDocSection = {
  heading: string
  body: ReactNode
}

export interface IComponentDocs {
  name: string
  tier: 'atom' | 'molecule' | 'organism' | 'template' | 'page'
  summary: string
  use?: string[]
  avoid?: string[]
  rules?: string[]
  props?: TDocProp[]
  sections?: TDocSection[]
}

const List = ({ items, marker }: { items: string[]; marker: string }) => (
  <ul className="flex flex-col gap-1.5">
    {items.map((item) => (
      <li key={item} className="flex gap-2">
        <span aria-hidden className="select-none text-tertiary">
          {marker}
        </span>
        <Text variant="body" color="secondary">
          {item}
        </Text>
      </li>
    ))}
  </ul>
)

const Section = ({ heading, children }: { heading: string; children: ReactNode }) => (
  <section className="flex flex-col gap-3">
    <Text as="h2" variant="label" color="tertiary">
      {heading}
    </Text>
    {children}
  </section>
)

export const ComponentDocs = ({
  name,
  tier,
  summary,
  use,
  avoid,
  rules,
  props,
  sections,
}: IComponentDocs) => (
  <article className="flex max-w-3xl flex-col gap-10 p-8">
    <header className="flex flex-col gap-3">
      <div className="flex items-baseline gap-3">
        <Text as="h1" variant="title">
          {name}
        </Text>
        <Text variant="label" color="tertiary" family="mono">
          {tier}
        </Text>
      </div>
      <Text variant="body" color="secondary">
        {summary}
      </Text>
    </header>

    {use?.length ? (
      <Section heading="Use it for">
        <List items={use} marker="—" />
      </Section>
    ) : null}

    {avoid?.length ? (
      <Section heading="Don't use it for">
        <List items={avoid} marker="—" />
      </Section>
    ) : null}

    {rules?.length ? (
      <Section heading="Rules">
        <List items={rules} marker="—" />
      </Section>
    ) : null}

    {props?.length ? (
      <Section heading="Props">
        <div className="overflow-hidden rounded-lg border border-divider">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="bg-surface-02">
                {['Prop', 'Type', 'Default', 'Description'].map((h) => (
                  <th key={h} className="px-3 py-2">
                    <Text variant="label" color="tertiary">
                      {h}
                    </Text>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {props.map((prop) => (
                <tr key={prop.name} className="border-t border-divider align-top">
                  <td className="px-3 py-2">
                    <Text variant="caption" family="mono" color="primary">
                      {prop.name}
                    </Text>
                  </td>
                  <td className="px-3 py-2">
                    <Text variant="caption" family="mono" color="tertiary">
                      {prop.type}
                    </Text>
                  </td>
                  <td className="px-3 py-2">
                    <Text variant="caption" family="mono" color="tertiary">
                      {prop.default ?? '—'}
                    </Text>
                  </td>
                  <td className="px-3 py-2">
                    <Text variant="caption" color="secondary">
                      {prop.description}
                    </Text>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Section>
    ) : null}

    {sections?.map((section) => (
      <Section key={section.heading} heading={section.heading}>
        {typeof section.body === 'string' ? (
          <Text variant="body" color="secondary">
            {section.body}
          </Text>
        ) : (
          section.body
        )}
      </Section>
    ))}
  </article>
)
