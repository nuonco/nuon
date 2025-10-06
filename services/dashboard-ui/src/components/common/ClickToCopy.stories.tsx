import { ClickToCopy, ClickToCopyButton } from './ClickToCopy'

export const Default = () => (
  <div className="flex flex-col gap-4">
    <ClickToCopy>Simple text to copy</ClickToCopy>
    <ClickToCopy>
      <span>Text inside a span element</span>
    </ClickToCopy>
  </div>
)

export const ButtonVariant = () => (
  <div className="flex gap-4 items-center">
    <ClickToCopyButton textToCopy="Hello World" />
    <ClickToCopyButton textToCopy="API_KEY_12345" />
    <ClickToCopyButton
      textToCopy="Custom styled button"
      noticeClassName="!bg-blue-500 text-white"
    />
  </div>
)
