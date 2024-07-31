import { Alert } from "@material-tailwind/react";

export const ErrorAlert = ({ children, setError = (error = {}) => {} }) => (
  <div className="fixed w-full right-0 bottom-0 left-0 p-4">
    <Alert
      color="red"
      onClose={() =>
        setError({
          description: "",
          error: "",
          user_error: false,
        })
      }
    >
      {children}
    </Alert>
  </div>
);
