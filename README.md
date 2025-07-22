# Markdown Parser Service

Go Microservice for markdown text editor 

- Converts markdown text to HTML
- Detects typing patterns like `# ` and converts them to headings instantly
- Supports GitHub Flavored Markdown (tables, checkboxes, code blocks)
- Provides both REST API and WebSocket for real-time updates
- Handles multiple users simultaneously

## Supported Syntax

| Type | Input | Output |
|------|-------|--------|
| Heading | `# text` | `<h1>text</h1>` |
| List | `- item` | `<ul><li>item</li></ul>` |
| Checkbox | `- [ ] task` | `<input type="checkbox">task` |
| Code | ``` | `<pre><code>` |

