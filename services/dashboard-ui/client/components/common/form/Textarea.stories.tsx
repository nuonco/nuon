/* eslint-disable react/no-unescaped-entities */
import { useState } from 'react'
import { Textarea } from './Textarea'

export default {
  title: 'Forms/Textarea',
  component: Textarea,
}

export const BasicUsage = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Basic Textarea Usage</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Multi-line text input with different configurations. The textarea 
        automatically handles styling, focus states, and accessibility attributes.
      </p>
    </div>

    <div className="space-y-4">
      <Textarea placeholder="Enter your text here..." />
      <Textarea
        placeholder="Textarea with label"
        labelProps={{ labelText: 'Description' }}
      />
      <Textarea
        placeholder="Textarea with helper text"
        helperText="Provide a detailed description"
        labelProps={{ labelText: 'Project Details' }}
      />
    </div>
  </div>
)

export const Sizes = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Textarea Sizes</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        The{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          size
        </code>{' '}
        prop controls the padding and font size of the textarea. Default size is{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          md
        </code>{' '}
        if no size prop is provided.
      </p>
    </div>

    <div className="space-y-4">
      <Textarea
        size="sm"
        placeholder="Small textarea"
        labelProps={{ labelText: 'Small' }}
        rows={3}
      />
      <Textarea
        size="md"
        placeholder="Medium textarea (default)"
        labelProps={{ labelText: 'Medium' }}
        rows={3}
      />
      <Textarea
        size="lg"
        placeholder="Large textarea"
        labelProps={{ labelText: 'Large' }}
        rows={3}
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-4">
      <div>
        <strong>sm:</strong> 8px padding, 12px font size
      </div>
      <div>
        <strong>md:</strong> 12px padding, 14px font size (default)
      </div>
      <div>
        <strong>lg:</strong> 16px padding, 16px font size
      </div>
    </div>
  </div>
)

export const AutoResize = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Auto-Resizing Textarea</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        When{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          autoResize
        </code>{' '}
        is enabled, the textarea automatically grows with content between{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          minRows
        </code>{' '}
        and{' '}
        <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
          maxRows
        </code>.
      </p>
    </div>

    <div className="space-y-4">
      <Textarea
        autoResize
        minRows={2}
        maxRows={6}
        placeholder="Start typing and watch me grow..."
        labelProps={{ labelText: 'Auto-resize (2-6 rows)' }}
        helperText="This textarea will grow as you type, up to 6 rows"
      />
      
      <Textarea
        autoResize
        minRows={3}
        maxRows={8}
        placeholder="Type a longer message here..."
        labelProps={{ labelText: 'Auto-resize (3-8 rows)' }}
        helperText="Larger range for longer content"
      />
      
      <Textarea
        autoResize={false}
        rows={4}
        placeholder="Fixed height textarea"
        labelProps={{ labelText: 'Fixed Height' }}
        helperText="This textarea has a fixed height and manual resize handle"
      />
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Auto-resize Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Grows automatically with content</li>
        <li>Respects minimum and maximum row constraints</li>
        <li>Shows scrollbar when content exceeds maxRows</li>
        <li>Disables manual resize handle when auto-resize is on</li>
        <li>Smooth transitions for better user experience</li>
      </ul>
    </div>
  </div>
)

export const States = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Textarea States</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Textarea components support various states including error states, disabled
        states, and helper text for better user experience and validation
        feedback.
      </p>
    </div>

    <div className="space-y-4">
      <Textarea
        placeholder="Normal textarea"
        labelProps={{ labelText: 'Normal State' }}
        helperText="This is a normal textarea field"
        rows={3}
      />
      
      <Textarea
        placeholder="Error textarea"
        labelProps={{ labelText: 'Error State' }}
        error
        errorMessage="This field is required and cannot be empty"
        rows={3}
      />
      
      <Textarea
        placeholder="Disabled textarea"
        labelProps={{ labelText: 'Disabled State' }}
        disabled
        helperText="This field is currently disabled"
        rows={3}
      />
      
      <Textarea
        placeholder="Required field"
        labelProps={{ labelText: 'Required Field' }}
        required
        helperText="This field is required"
        rows={3}
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
        <strong>Required:</strong> Semantic HTML required attribute with validation
      </div>
    </div>
  </div>
)

export const WithLabelsAndText = () => (
  <div className="space-y-6 max-w-md">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Labels and Helper Text</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Textareas support labels and helper text to provide context and guidance
        to users. Labels are automatically associated with textareas for
        accessibility.
      </p>
    </div>

    <div className="space-y-4">
      <Textarea
        placeholder="Describe your project in detail..."
        labelProps={{ 
          labelText: 'Project Description',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="Provide a comprehensive overview of your project goals"
        rows={4}
      />
      
      <Textarea
        placeholder="Enter feedback or comments..."
        labelProps={{ labelText: 'Additional Comments' }}
        helperText="Any additional information or special requirements"
        rows={3}
      />
      
      <Textarea
        placeholder="Describe technical requirements..."
        labelProps={{ labelText: 'Technical Specifications' }}
        helperText="Include technical details, frameworks, or constraints"
        rows={4}
      />
      
      <Textarea
        placeholder="Enter notes (optional)..."
        labelProps={{ labelText: 'Notes (Optional)' }}
        rows={2}
      />
    </div>

    <div className="grid grid-cols-1 gap-3 text-sm mt-6">
      <div>
        <strong>Labels:</strong> Automatically linked to textareas with htmlFor
        attribute
      </div>
      <div>
        <strong>Helper Text:</strong> Provides additional context below the
        textarea
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
      <Textarea
        placeholder="Enter a description..."
        labelProps={{ labelText: 'Description' }}
        error
        errorMessage="Description is required"
        rows={3}
      />
      
      <Textarea
        placeholder="Enter your message..."
        labelProps={{ labelText: 'Message' }}
        error
        errorMessage="Message must be at least 10 characters long"
        rows={4}
      />
      
      <Textarea
        placeholder="Enter comments..."
        labelProps={{ labelText: 'Comments' }}
        error
        errorMessage="Please provide valid input without special characters"
        rows={3}
      />
      
      <Textarea
        placeholder="Enter feedback..."
        labelProps={{ labelText: 'Feedback' }}
        error
        errorMessage="Feedback cannot exceed 500 characters"
        rows={4}
      />
    </div>

    <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
      <strong>Error State Features:</strong>
      <ul className="mt-2 space-y-1 list-disc list-inside">
        <li>Red border and focus ring for visual indication</li>
        <li>Error messages displayed below the textarea</li>
        <li>Proper ARIA attributes for screen readers</li>
        <li>Error messages override helper text when both are present</li>
        <li>Custom styling support for error messages</li>
      </ul>
    </div>
  </div>
)

export const RealWorldExamples = () => (
  <div className="space-y-6">
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Real-World Examples</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">
        Common use cases for Textarea in application forms, showing practical
        examples you might encounter in a development or content management
        environment.
      </p>
    </div>

    <div className="space-y-6">
      <Textarea
        autoResize
        minRows={4}
        maxRows={8}
        placeholder="Describe the issue you're experiencing..."
        labelProps={{
          labelText: 'Bug Report Description',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="Please provide detailed steps to reproduce the issue"
      />
      
      <Textarea
        rows={6}
        placeholder="Enter your commit message..."
        labelProps={{
          labelText: 'Commit Message',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="First line should be a concise summary (max 50 chars)"
      />
      
      <Textarea
        autoResize
        minRows={3}
        maxRows={10}
        placeholder="Document your API endpoint..."
        labelProps={{
          labelText: 'API Documentation',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="Include endpoint description, parameters, and example responses"
      />

      <Textarea
        rows={5}
        placeholder="Add release notes..."
        labelProps={{
          labelText: 'Release Notes',
          labelTextProps: { weight: 'strong' }
        }}
        helperText="Describe new features, bug fixes, and breaking changes"
      />
    </div>
  </div>
)

export const FormIntegration = () => {
  const [formData, setFormData] = useState({
    title: 'Project Alpha',
    description: '',
    notes: '',
    requirements: ''
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    alert(`Form submitted!\n\nTitle: ${formData.title}\nDescription: ${formData.description.slice(0, 50)}...\nNotes: ${formData.notes.slice(0, 50)}...`)
  }

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Form Integration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Textarea integrates seamlessly with forms and controlled components.
          It supports standard form attributes like{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            name
          </code>,{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            required
          </code>, and{' '}
          <code className="px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded text-xs">
            onChange
          </code>{' '}
          handlers.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6 max-w-3xl">
        <div>
          <label htmlFor="project-title" className="block text-sm font-medium mb-1">
            Project Title *
          </label>
          <input
            id="project-title"
            type="text"
            value={formData.title}
            onChange={(e) => setFormData(prev => ({ ...prev, title: e.target.value }))}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 bg-white dark:bg-gray-900"
            placeholder="Enter project title"
            required
          />
        </div>
        
        <Textarea
          id="project-description"
          name="description"
          autoResize
          minRows={4}
          maxRows={8}
          value={formData.description}
          onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
          labelProps={{
            labelText: 'Project Description *',
          }}
          helperText="Provide a detailed description of your project goals and scope"
          required
        />

        <Textarea
          id="project-requirements"
          name="requirements"
          rows={4}
          value={formData.requirements}
          onChange={(e) => setFormData(prev => ({ ...prev, requirements: e.target.value }))}
          labelProps={{
            labelText: 'Technical Requirements',
          }}
          helperText="List any specific technical requirements or constraints"
        />

        <Textarea
          id="project-notes"
          name="notes"
          autoResize
          minRows={2}
          maxRows={6}
          value={formData.notes}
          onChange={(e) => setFormData(prev => ({ ...prev, notes: e.target.value }))}
          labelProps={{
            labelText: 'Additional Notes',
          }}
          helperText="Any additional comments or special considerations"
        />
        
        <div className="flex gap-3">
          <button
            type="submit"
            className="px-6 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            Create Project
          </button>
          <button
            type="button"
            onClick={() => setFormData({
              title: '',
              description: '',
              notes: '',
              requirements: ''
            })}
            className="px-6 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-50 dark:hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-gray-500"
          >
            Reset
          </button>
        </div>

        <div className="text-sm text-gray-600 dark:text-gray-400 mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
          <strong>Form Data:</strong>
          <pre className="mt-2 text-xs overflow-x-auto">
            {JSON.stringify(formData, null, 2)}
          </pre>
        </div>
      </form>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Form Integration Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Controlled and uncontrolled component support</li>
          <li>Standard form attributes (name, required, etc.)</li>
          <li>onChange event handling for real-time validation</li>
          <li>Form submission with proper value serialization</li>
          <li>Integration with form libraries like React Hook Form</li>
          <li>Auto-resize capability for dynamic content</li>
        </ul>
      </div>
    </div>
  )
}