import { Logo } from './Logo'

export default { title: 'Common/Logo' }

export const System = () => <Logo />

export const Light = () => (
  <div className="bg-dark-grey-900 p-4">
    <Logo variant="light" />
  </div>
)

export const Dark = () => (
  <div className="bg-white p-4">
    <Logo variant="dark" />
  </div>
)
