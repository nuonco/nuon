import { Divider } from './Divider'

export const Default = () => (
  <div className="w-full max-w-md space-y-8">
    <div>
      <p>Content above divider</p>
      <Divider />
      <p>Content below divider</p>
    </div>
  </div>
)

export const WithWord = () => (
  <div className="w-full max-w-md space-y-8">
    <div>
      <p>Login with email</p>
      <Divider dividerWord="OR" />
      <p>Login with OAuth</p>
    </div>
  </div>
)

export const InForm = () => (
  <div className="w-full max-w-md space-y-4 p-6 border rounded-lg">
    <div className="space-y-2">
      <label className="block text-sm font-medium">Email</label>
      <input
        type="email"
        className="w-full px-3 py-2 border rounded-md"
        placeholder="Enter your email"
      />
    </div>
    <div className="space-y-2">
      <label className="block text-sm font-medium">Password</label>
      <input
        type="password"
        className="w-full px-3 py-2 border rounded-md"
        placeholder="Enter your password"
      />
    </div>
    <button className="w-full bg-blue-600 text-white py-2 rounded-md">
      Sign In
    </button>

    <Divider dividerWord="OR" />

    <button className="w-full border border-gray-300 py-2 rounded-md">
      Continue with Google
    </button>
  </div>
)
