import { BranchRunCommit } from './BranchRunCommit'

export default {
  title: 'Branches/BranchRunCommit',
}

export const Success = () => (
  <div className="max-w-md">
    <BranchRunCommit
      status="success"
      href="#"
      message="feat: add resources section to customer portal readme (#273)"
      author="Nat Hamilton"
      avatarUrl="https://github.com/nat.png"
      sha="85d067ecafe1234"
      createdAt="2026-08-12T09:00:00Z"
    />
  </div>
)

export const Running = () => (
  <div className="max-w-md">
    <BranchRunCommit
      status="running"
      href="#"
      message="test: testing app branch changes"
      author="Nat Hamilton"
      sha="4e41797abcd"
      createdAt="2026-08-12T10:30:00Z"
    />
  </div>
)

export const Inline = () => (
  <div className="max-w-md">
    <BranchRunCommit
      displayVariant="inline"
      showStatus={false}
      href="#"
      message="feat: add retry budget to the checkout worker"
      author="developer"
      sha="a1b2c3d4e5f6"
      createdAt="2026-08-12T09:00:00Z"
    />
  </div>
)

export const InlineLongMessageTruncates = () => (
  <div className="max-w-xs">
    <BranchRunCommit
      displayVariant="inline"
      showStatus={false}
      href="#"
      message="feat: add retry budget to the checkout worker and rework the queue drain path"
      author="developer"
      sha="a1b2c3d4e5f6"
      createdAt="2026-08-12T09:00:00Z"
    />
  </div>
)

export const NoCommitMeta = () => (
  <div className="max-w-md">
    <BranchRunCommit status="pending" href="#" />
  </div>
)

export const LongMessageTruncates = () => (
  <div className="max-w-md">
    <BranchRunCommit
      status="success"
      href="#"
      message="feat: add resources section to customer portal readme (#273) add Resources section to customer portal readme with a very long trailing description"
      author="Nat Hamilton"
      avatarUrl="https://github.com/nat.png"
      sha="85d067ecafe1234"
      createdAt="2026-08-12T09:00:00Z"
    />
  </div>
)
