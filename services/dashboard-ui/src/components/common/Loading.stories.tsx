import { Loading } from './Loading'

export const Default = () => <Loading />

export const Large = () => <Loading variant="large" />

export const AllVariants = () => (
  <div className="flex flex-col gap-6">
    <div className="flex items-center gap-4">
      <Loading variant="default" />
      <span>Default variant (h-5 w-5)</span>
    </div>
    <div className="flex items-center gap-4">
      <Loading variant="large" strokeWidth="thick" />
      <span>Large variant (h-10 w-10)</span>
    </div>
  </div>
)

export const WithDifferentColors = () => (
  <div className="flex flex-col gap-6">
    <div className="flex items-center gap-4">
      <div className="text-blue-500">
        <Loading />
      </div>
      <span>Blue loading spinner</span>
    </div>
    <div className="flex items-center gap-4">
      <div className="text-green-500">
        <Loading variant="large" />
      </div>
      <span>Green large loading spinner</span>
    </div>
    <div className="flex items-center gap-4">
      <div className="text-red-500">
        <Loading />
      </div>
      <span>Red loading spinner</span>
    </div>
    <div className="flex items-center gap-4">
      <div className="text-purple-500">
        <Loading variant="large" />
      </div>
      <span>Purple large loading spinner</span>
    </div>
  </div>
)

export const InDifferentContexts = () => (
  <div className="flex flex-col gap-8">
    <div className="border rounded-lg p-4">
      <h3 className="text-lg font-semibold mb-4">Loading in a card</h3>
      <div className="flex items-center justify-center py-8">
        <Loading variant="large" />
      </div>
    </div>

    <div className="flex items-center gap-2">
      <Loading />
      <span>Loading inline with text...</span>
    </div>

    <div className="bg-gray-100 rounded-lg p-6 flex items-center justify-center">
      <div className="text-center">
        <Loading variant="large" />
        <p className="mt-2 text-sm text-gray-600">Loading content...</p>
      </div>
    </div>

    <div className="flex items-center justify-between p-4 border rounded-lg">
      <span>Processing your request</span>
      <Loading />
    </div>
  </div>
)

export const WithCustomStyling = () => (
  <div className="flex flex-col gap-6">
    <div className="flex items-center gap-4">
      <div className="p-4 bg-gray-50 rounded-lg">
        <Loading variant="large" />
      </div>
      <span>With background container</span>
    </div>

    <div className="flex items-center gap-4">
      <div className="text-6xl">
        <Loading variant="large" />
      </div>
      <span>With larger text size context</span>
    </div>

    <div className="flex items-center gap-4">
      <div className="text-xs">
        <Loading />
      </div>
      <span>With smaller text size context</span>
    </div>
  </div>
)

export const InButtonContext = () => (
  <div className="flex flex-col gap-4">
    <button className="flex items-center gap-2 bg-blue-500 text-white px-4 py-2 rounded-lg">
      <Loading />
      <span>Loading...</span>
    </button>

    <button className="flex items-center gap-2 bg-green-500 text-white px-6 py-3 rounded-lg">
      <Loading variant="large" />
      <span>Processing...</span>
    </button>

    <button className="flex items-center justify-center bg-gray-500 text-white px-4 py-2 rounded-lg w-32">
      <Loading />
    </button>
  </div>
)
