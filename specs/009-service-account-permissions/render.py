from pathlib import Path
import json
import subprocess
import tempfile


directory = Path(__file__).resolve().parent
source = (directory / "prompt.md").read_text()
template = """<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>009 — Granular service-account access</title>
<style>
:root { color-scheme: light; --ink: #182e3c; --muted: #526674; --line: #d5dfe5; --accent: #13675e; }
* { box-sizing: border-box; }
html { scroll-behavior: smooth; scroll-padding-top: 24px; }
body { margin: 0; color: var(--ink); background: #f4f6f7; font: 16px/1.7 system-ui, sans-serif; }
.masthead { background: #163d3a; color: #fff; padding: 18px max(24px, calc((100vw - 1440px) / 2)); display: flex; gap: 20px; justify-content: space-between; font-size: 13px; letter-spacing: .06em; }
.layout { display: grid; grid-template-columns: 260px minmax(0, 1fr); max-width: 1440px; margin: auto; }
nav { padding: 30px 22px; position: sticky; top: 0; align-self: start; max-height: 100vh; overflow: auto; font-size: 13px; }
nav h2 { font-size: 12px; text-transform: uppercase; letter-spacing: .1em; border: 0; margin: 0 0 16px; padding: 0; }
nav ul { list-style: none; padding: 0; }
nav li { margin: 0 0 12px; }
nav a { color: var(--muted); text-decoration: none; }
nav a:hover { color: var(--accent); text-decoration: underline; }
nav .review-link { display: block; padding-top: 18px; border-top: 1px solid var(--line); font-weight: 650; }
main { background: #fff; padding: 42px 48px 72px; min-width: 0; border-left: 1px solid var(--line); }
h1, h2, h3 { line-height: 1.25; letter-spacing: -.025em; }
h1 { font-size: clamp(30px, 3.2vw, 46px); margin: 0 0 14px; }
h2 { margin: 48px 0 20px; padding-top: 26px; border-top: 1px solid var(--line); font-size: 26px; }
h3 { margin-top: 30px; font-size: 19px; }
p, li { max-width: 90ch; }
li { margin-bottom: 7px; }
a { color: #08685e; text-underline-offset: 3px; overflow-wrap: anywhere; }
code { font: .88em/1.6 ui-monospace, SFMono-Regular, Menlo, monospace; background: #edf3f3; padding: 2px 4px; border-radius: 3px; overflow-wrap: anywhere; }
pre { padding: 22px; border: 1px solid var(--line); background: #f0f6f5; overflow: auto; font-size: 13px; }
pre code { background: none; padding: 0; white-space: pre; }
.architecture { overflow: auto; padding: 16px; border: 1px solid var(--line); }
.architecture svg { display: block; min-width: 700px; margin: auto; }
.table-scroll { overflow-x: auto; margin: 24px 0; border: 1px solid var(--line); border-radius: 5px; }
table { width: 100%; border-collapse: collapse; font-size: 14px; line-height: 1.6; }
th, td { text-align: left; vertical-align: top; padding: 14px 16px; min-width: 150px; border-bottom: 1px solid var(--line); }
th { background: #e9f1ef; font-weight: 650; }
tr:last-child td { border-bottom: 0; }
tbody tr:nth-child(even) { background: #fafcfc; }
button { font: inherit; cursor: pointer; border: 1px solid #a9babd; border-radius: 4px; color: var(--ink); background: #fff; padding: 6px 12px; }
button:hover { background: #e9f1ef; }
button:focus-visible, a:focus-visible, summary:focus-visible, textarea:focus { outline: 3px solid #1b8073; outline-offset: 2px; }
.listing { margin: 20px 0; border: 1px solid var(--line); font: 13px/1.65 ui-monospace, monospace; }
.review-line { display: grid; grid-template-columns: 32px 36px minmax(0, 1fr); align-items: start; gap: 10px; padding: 6px 10px; }
.review-line:nth-child(even) { background: #f3f6f7; }
.review-line.added { background: #e3f3e9; }
.review-line.removed { background: #fbeaea; }
.review-line button { padding: 0; height: 26px; width: 28px; }
.line-number { color: var(--muted); text-align: right; user-select: none; }
.line-text { white-space: pre-wrap; overflow-wrap: anywhere; min-width: 0; }
summary { cursor: pointer; font-weight: 650; }
.note { color: var(--muted); font-size: 14px; }
dialog { color: var(--ink); border: 1px solid var(--line); border-radius: 8px; width: min(660px, 92vw); padding: 26px; }
dialog::backdrop { background: #132c3a88; }
dialog h2 { margin: 0 0 18px; padding: 0; border: 0; font-size: 22px; }
dialog blockquote { margin: 12px 0; padding: 12px; background: #f0f6f5; white-space: pre-wrap; overflow-wrap: anywhere; max-height: 160px; overflow: auto; }
textarea { width: 100%; min-height: 150px; font: inherit; padding: 10px; margin: 8px 0 16px; }
.actions { display: flex; justify-content: flex-end; flex-wrap: wrap; gap: 12px; }
.primary { background: #13675e; color: #fff; border-color: #13675e; }
.primary:hover { background: #0c514a; }
@media (max-width: 950px) { .layout { grid-template-columns: 210px minmax(0, 1fr); } main { padding: 30px 24px; } }
@media (max-width: 700px) { .layout { display: block; } nav { position: static; max-height: none; padding: 20px; } nav > ul > li > ul { columns: 2; column-gap: 22px; } nav li { break-inside: avoid; } main { padding: 28px 20px 50px; border: 0; } .masthead { padding: 16px 20px; flex-wrap: wrap; gap: 6px 20px; } h2 { font-size: 23px; } .review-line { gap: 6px; grid-template-columns: 28px 27px minmax(0, 1fr); padding: 6px; } }
@media (max-width: 700px) { pre code { white-space: pre-wrap; } }
@media (prefers-reduced-motion: reduce) { html { scroll-behavior: auto; } }
@media print { nav, .masthead, #review, dialog, .review-line button { display: none; } body, main { background: #fff; } .layout { display: block; } main { padding: 0; border: 0; } .review-line { grid-template-columns: 36px minmax(0, 1fr); } .table-scroll { overflow: visible; } table { font-size: 10px; } th, td { min-width: 0; padding: 6px; } h2, h3 { break-after: avoid; } }
</style>
</head>
<body>
<header class="masthead"><span>NUON / ENGINEERING PROPOSAL 009</span><span>GRANULAR ACCESS · SERVICE-ACCOUNT UI</span></header>
<div class="layout">
<nav aria-label="Proposal contents"><h2>On this page</h2>$table-of-contents$<a class="review-link" href="#review">Review and comment →</a><p class="note"><a href="prompt.md">Markdown source</a><br>Built on spec 008's native model.</p></nav>
<main id="main">$body$
<section id="review" aria-labelledby="review-heading">
<h2 id="review-heading">Line-by-line review</h2>
<p>Open the source below and click <strong>+</strong> beside any line. The illustrative diff above supports comments too. Each comment downloads as Markdown with its source excerpt.</p>
<p class="note">Save downloads in <code>specs/009-service-account-permissions/</code>. Nothing is submitted to a server. Unsubmitted text is not retained after closing this page.</p>
<details><summary>Review the complete proposal source</summary><div id="source-lines" class="listing"></div></details>
<p id="review-status" role="status" aria-live="polite"></p>
</section>
</main></div>
<dialog id="comment-dialog" aria-labelledby="comment-title">
<form id="comment-form"><h2 id="comment-title">Review comment</h2>
<blockquote id="comment-excerpt"></blockquote>
<label for="comment-text">Your comment</label><textarea id="comment-text" required></textarea>
<div class="actions"><button type="button" id="cancel-comment">Cancel</button><button class="primary" type="submit">Download comment</button></div>
</form></dialog>
<script id="markdown-source" type="application/json">SOURCE_JSON</script>
<script>
const source = JSON.parse(document.getElementById('markdown-source').textContent);
const dialog = document.getElementById('comment-dialog');
const input = document.getElementById('comment-text');
let target;
function listing(root, lines, file, diff = false) {
  lines.forEach((text, index) => {
    const row = document.createElement('div');
    row.className = 'review-line' + (diff && text.startsWith('+') ? ' added' : diff && text.startsWith('-') ? ' removed' : '');
    const button = document.createElement('button');
    button.type = 'button';
    button.textContent = '+';
    button.setAttribute('aria-label', 'Comment on ' + file + ' line ' + (index + 1));
    button.onclick = () => {
      target = {file, line: index + 1, text};
      document.getElementById('comment-title').textContent = file + ':' + target.line;
      document.getElementById('comment-excerpt').textContent = text;
      input.value = '';
      dialog.showModal();
      input.focus();
    };
    const number = document.createElement('span');
    number.className = 'line-number';
    number.textContent = index + 1;
    const content = document.createElement('span');
    content.className = 'line-text';
    content.textContent = text;
    row.append(button, number, content);
    root.append(row);
  });
}
document.querySelectorAll('table').forEach(table => {
  const wrapper = document.createElement('div');
  wrapper.className = 'table-scroll';
  wrapper.tabIndex = 0;
  wrapper.setAttribute('role', 'region');
  wrapper.setAttribute('aria-label', 'Scrollable proposal table');
  table.before(wrapper);
  wrapper.append(table);
});
document.querySelectorAll('pre.diff, pre.sourceCode.diff').forEach(pre => {
  const root = document.createElement('div');
  root.className = 'listing';
  root.setAttribute('aria-label', 'Illustrative authorization contract diff');
  listing(root, pre.textContent.trimEnd().split('\\n'), 'illustrative-contract.diff', true);
  pre.replaceWith(root);
});
listing(document.getElementById('source-lines'), source.trimEnd().split('\\n'), 'prompt.md');
document.getElementById('cancel-comment').onclick = () => dialog.close();
document.getElementById('comment-form').onsubmit = event => {
  event.preventDefault();
  if (!input.value.trim()) return;
  const body = '# Spec 009 review: ' + target.file + ':' + target.line + '\\n\\nSource excerpt:\\n\\n    ' + target.text + '\\n\\n' + input.value.trim() + '\\n';
  const url = URL.createObjectURL(new Blob([body], {type: 'text/markdown;charset=utf-8'}));
  const link = document.createElement('a');
  link.href = url;
  link.download = 'comment-' + target.file + '-L' + target.line + '-' + Date.now() + '.md';
  link.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
  document.getElementById('review-status').textContent = 'Comment download requested. Save it in specs/009-service-account-permissions/.';
  dialog.close();
};
</script>
<script type="module">
try {
  const {default: mermaid} = await import('https://cdn.jsdelivr.net/npm/mermaid@11.4.1/dist/mermaid.esm.min.mjs');
  mermaid.initialize({startOnLoad: false, securityLevel: 'strict', theme: 'neutral'});
  for (const [index, pre] of [...document.querySelectorAll('pre.mermaid')].entries()) {
    const {svg} = await mermaid.render('architecture-' + index, pre.textContent);
    const diagram = document.createElement('div');
    diagram.className = 'architecture';
    diagram.setAttribute('role', 'img');
    diagram.setAttribute('aria-label', 'Permission catalog feeds authoring and enforcement; live policies, token ceilings and trusted ownership constrain reads and writes');
    diagram.innerHTML = svg;
    pre.replaceWith(diagram);
  }
} catch {
  document.querySelectorAll('pre.mermaid').forEach(pre => {
    const note = document.createElement('p');
    note.className = 'note';
    note.textContent = 'Diagram renderer unavailable. Architecture source is shown below; text and review comments still work offline.';
    pre.before(note);
  });
}
</script>
</body></html>
""".replace("SOURCE_JSON", json.dumps(source, ensure_ascii=False).replace("</", "<\\/"))

with tempfile.TemporaryDirectory() as temporary:
    template_file = Path(temporary) / "template.html"
    template_file.write_text(template)
    subprocess.run(
        [
            "pandoc", str(directory / "prompt.md"), "--standalone", "--to=html5",
            "--toc", "--toc-depth=2", f"--template={template_file}",
            "--output", str(directory / "spec.html"),
        ],
        check=True,
    )
