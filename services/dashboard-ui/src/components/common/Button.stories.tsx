import { Button } from "./Button";

export const Variants = () => (
  <div className="flex gap-4 items-center">
    <Button>Default button</Button>
    <Button variant="primary">Primary button</Button>
    <Button variant="danger">Danger button</Button>
    <Button variant="ghost">Ghost button</Button>
    <Button variant="tab">Tab button</Button>
  </div>
);

export const Sizes = () => (
  <div className="flex gap-4 items-center">
    <Button size="xs">XS button</Button>
    <Button size="sm">SM button</Button>
    <Button>MD button</Button>
    <Button size="lg">LG button</Button>
  </div>
);

export const Links = () => (
  <div className="flex gap-4 items-center">
    <Button href="/">Internal link</Button>
    <Button href="https://nuon.co" target="_blank">
      External link
    </Button>
  </div>
);

export const Disabled = () => (
  <div className="flex gap-4 items-center">
    <Button disabled>Default button</Button>
    <Button variant="primary" disabled>
      Primary button
    </Button>
    <Button variant="danger" disabled>
      Danger button
    </Button>
    <Button variant="ghost" disabled>
      Ghost button
    </Button>
  </div>
);

export const TabButtons = () => (
  <div className="flex border-b items-center">
    <Button isActive variant="tab" href="#">Tab button</Button>
    <Button variant="tab">Tab button</Button>
    <Button variant="tab">Tab button</Button>
  </div>
);
