/* eslint-disable react/no-unescaped-entities */
import { Input } from './Input'

export default {
  title: 'Forms/Input',
  component: Input,
}

export const BasicUsage = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Input Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Basic input field with different configurations. The input automatically
        handles styling, focus states, and accessibility attributes.
      </p>
    </div>

    <div className="space-y-4">
      <Input placeholder="Enter your text here..." />
      <Input
        placeholder="Input with label"
        labelProps={{ labelText: 'Username' }}
      />
      <Input
        placeholder="Input with helper text"
        helperText="This is some helpful information"
        labelProps={{ labelText: 'Email Address' }}
      />
    </div>
  </div>
)

export const Sizes = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Input Sizes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          size
        </code>{' '}
        prop controls the dimensions and padding of the input. Default size is{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          md
        </code>{' '}
        if no size prop is provided.
      </p>
    </div>

    <div className="space-y-4">
      <Input
        size="sm"
        placeholder="Small input"
        labelProps={{ labelText: 'Small' }}
      />
      <Input
        size="md"
        placeholder="Medium input (default)"
        labelProps={{ labelText: 'Medium' }}
      />
      <Input
        size="lg"
        placeholder="Large input"
        labelProps={{ labelText: 'Large' }}
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-4">
      <div>
        <strong>sm:</strong> 32px height, 8px vertical padding
      </div>
      <div>
        <strong>md:</strong> 40px height, 8px vertical padding (default)
      </div>
      <div>
        <strong>lg:</strong> 48px height, 12px vertical padding
      </div>
    </div>
  </div>
)

export const InputTypes = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Input Types</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The input component supports all standard HTML input types with
        appropriate styling and behavior for each type.
      </p>
    </div>

    <div className="space-y-4">
      <Input
        type="text"
        placeholder="Text input"
        labelProps={{ labelText: 'Text' }}
      />
      <Input
        type="email"
        placeholder="user@example.com"
        labelProps={{ labelText: 'Email' }}
      />
      <Input
        type="password"
        placeholder="••••••••"
        labelProps={{ labelText: 'Password' }}
      />
      <Input
        type="number"
        placeholder="123"
        labelProps={{ labelText: 'Number' }}
      />
      <Input
        type="url"
        placeholder="https://example.com"
        labelProps={{ labelText: 'URL' }}
      />
      <Input
        type="tel"
        placeholder="+1 (555) 123-4567"
        labelProps={{ labelText: 'Phone' }}
      />
      <Input
        type="search"
        placeholder="Search..."
        labelProps={{ labelText: 'Search' }}
      />
    </div>
  </div>
)

export const States = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Input States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Input components support various states including error states, disabled
        states, and helper text for better user experience and validation
        feedback.
      </p>
    </div>

    <div className="space-y-4">
      <Input
        placeholder="Normal input"
        labelProps={{ labelText: 'Normal State' }}
        helperText="This is a normal input field"
      />
      
      <Input
        placeholder="Error input"
        labelProps={{ labelText: 'Error State' }}
        error
        errorMessage="This field is required"
      />
      
      <Input
        placeholder="Disabled input"
        labelProps={{ labelText: 'Disabled State' }}
        disabled
        helperText="This field is currently disabled"
      />
      
      <Input
        placeholder="Required field"
        labelProps={{ labelText: 'Required Field' }}
        required
        helperText="This field is required"
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Normal:</strong> Default styling with focus states
      </div>
      <div>
        <strong>Error:</strong> Red border and error message styling
      </div>
      <div>
        <strong>Disabled:</strong> Muted colors and no interaction
      </div>
      <div>
        <strong>Required:</strong> Semantic HTML required attribute
      </div>
    </div>
  </div>
)

export const WithLabelsAndText = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Labels and Helper Text</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Inputs support labels and helper text to provide context and guidance
        to users. Labels are automatically associated with inputs for
        accessibility.
      </p>
    </div>

    <div className="space-y-4">
      <Input
        placeholder="Enter your full name"
        labelProps={{ 
          labelText: 'Full Name',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="Enter your first and last name"
      />
      
      <Input
        type="email"
        placeholder="user@company.com"
        labelProps={{ labelText: 'Work Email' }}
        helperText="We'll use this for account notifications"
      />
      
      <Input
        type="password"
        placeholder="••••••••••"
        labelProps={{ labelText: 'Password' }}
        helperText="Must be at least 8 characters long"
      />
      
      <Input
        placeholder="Enter company name"
        labelProps={{ labelText: 'Company (Optional)' }}
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Labels:</strong> Automatically linked to inputs with htmlFor
        attribute
      </div>
      <div>
        <strong>Helper Text:</strong> Provides additional context below the
        input
      </div>
      <div>
        <strong>Styling:</strong> Labels support all Text component props for
        customization
      </div>
      <div>
        <strong>Accessibility:</strong> Proper ARIA attributes and associations
      </div>
    </div>
  </div>
)

export const ErrorHandling = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Error Handling</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Error states provide visual feedback and error messages to guide users
        in correcting their input. Error messages replace helper text when
        present.
      </p>
    </div>

    <div className="space-y-4">
      <Input
        placeholder="Enter username"
        labelProps={{ labelText: 'Username' }}
        error
        errorMessage="Username is required"
      />
      
      <Input
        type="email"
        placeholder="user@example.com"
        labelProps={{ labelText: 'Email Address' }}
        error
        errorMessage="Please enter a valid email address"
      />
      
      <Input
        type="password"
        placeholder="••••••••"
        labelProps={{ labelText: 'Password' }}
        error
        errorMessage="Password must be at least 8 characters"
      />
      
      <Input
        placeholder="Enter a number"
        labelProps={{ labelText: 'Age' }}
        error
        errorMessage="Age must be between 18 and 120"
      />
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Error State Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Red border and focus ring for visual indication</li>
        <li>Error messages displayed below the input</li>
        <li>Proper ARIA attributes for screen readers</li>
        <li>Error messages override helper text when both are present</li>
        <li>Custom styling support for error messages</li>
      </ul>
    </div>
  </div>
)