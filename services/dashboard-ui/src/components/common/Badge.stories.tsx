import { Badge } from "./Badge";

export const Themes = () => (
  <div className="flex gap-4 items-center">
    <Badge>Neutral badge</Badge>
    <Badge theme="success">Success badge</Badge>
    <Badge theme="warn">Warn badge</Badge>
    <Badge theme="error">Error badge</Badge>
    <Badge theme="info">Info badge</Badge>
  </div>
);

export const Sizes = () => (
  <div className="flex gap-4 items-center">
    <Badge size="sm">SM badge</Badge>
    <Badge size="md">MD badge</Badge>
    <Badge size="lg">LG badge</Badge>
  </div>
);

export const Variants = () => (
  <div className="flex gap-4 items-center">
    <Badge variant="default">Default badge</Badge>
    <Badge variant="code">Code badge</Badge>
  </div>
);
