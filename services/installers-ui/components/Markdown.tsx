import showdown from "showdown";

// load extension to add target="_blank" to links
showdown.extension("targetlink", () => {
  return [
    {
      type: "html",
      regex: /(<a [^>]+?)(>.*<\/a>)/g,
      replace: '$1 target="_blank"$2',
    },
  ];
});

// instantiate converter
const markdown = new showdown.Converter({ extensions: ["targetlink"] });

export const Markdown = ({ content = "" }) => (
  <div
    className="prose dark:prose-invert"
    dangerouslySetInnerHTML={{
      __html: markdown.makeHtml(content),
    }}
  />
);
