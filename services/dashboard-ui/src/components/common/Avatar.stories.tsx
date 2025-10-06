import { Avatar } from './Avatar'

export const Default = () => <Avatar name="Nat Friedman" />

export const Sizes = () => (
  <div className="flex items-center gap-4">
    <Avatar name="Nat Friedman" size="xs" />
    <Avatar name="Nat Friedman" size="sm" />
    <Avatar name="Nat Friedman" size="md" />
    <Avatar name="Nat Friedman" size="lg" />
    <Avatar name="Nat Friedman" size="xl" />
  </div>
)

export const WithImage = () => (
  <Avatar src="https://github.com/nat.png" alt="Nat Friedman" />
)

export const Loading = () => <Avatar isLoading />
