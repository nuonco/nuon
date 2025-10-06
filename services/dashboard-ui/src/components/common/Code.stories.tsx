import { Code } from './Code'

export const Default = () => <Code>This is a default code block.</Code>

export const Preformatted = () => (
  <Code variant="preformated">
    {`
      {
        formatted: true
      }
    `}
  </Code>
)

export const Inline = () => (
  <p>
    This is an example of an <Code variant="inline">inline code</Code> snippet.
  </p>
)
