import { Button } from "@material-tailwind/react";

export const NavArrows = ({
  handlePrev = () => {},
  isFirstStep = false,
  handleNext = () => {},
  isLastStep = false,
}) => (
  <div className="flex justify-between">
    <Button
      className="rounded-md flex items-center gap-4 pl-3 pr-7 bg-black dark:bg-white text-white dark:text-black"
      onClick={handlePrev}
      disabled={isFirstStep}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        strokeWidth={1.5}
        stroke="currentColor"
        className="size-6"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M15.75 19.5 8.25 12l7.5-7.5"
        />
      </svg>
      Previous
    </Button>
    <Button
      className="rounded-md flex items-center gap-4 pl-7 pr-3 bg-black dark:bg-white text-white dark:text-black"
      onClick={handleNext}
      disabled={isLastStep}
    >
      Next
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        strokeWidth={1.5}
        stroke="currentColor"
        className="size-6"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="m8.25 4.5 7.5 7.5-7.5 7.5"
        />
      </svg>
    </Button>
  </div>
);
