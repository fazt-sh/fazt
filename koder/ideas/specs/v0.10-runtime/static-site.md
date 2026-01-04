# Static Site Generator

## Summary

Zero-config static site generator with profile-based detection. Drop into any
Jekyll, Docusaurus, VitePress, or Docsify project and run `fazt serve` - it
just works.

## Why

Static sites are the simplest deployment model:
- Fast (pre-rendered HTML)
- Secure (no server-side code)
- Cheap (just file serving)

But users have existing sites in different formats. Rather than force migration,
Fazt detects the format and builds appropriately.

**Philosophy alignment:**
- Single binary: All profiles built-in (~500KB additional)
- Zero-config: Auto-detect format, no setup needed
- Convention over configuration: Sensible defaults per profile

## Quick Start

```bash
# Any existing project - just works
cd my-docusaurus-docs
fazt serve              # Auto-detects Docusaurus, serves locally

cd my-jekyll-blog
fazt serve              # Auto-detects Jekyll, serves locally

cd my-plain-html
fazt serve              # Auto-detects static, serves as-is
```

Open http://localhost:4000 - you see your site with a beautiful theme.

```bash
fazt deploy             # Ship it (server builds, no local _site needed)
```

### New Project

```bash
fazt ssg init           # Creates minimal Jekyll blog
fazt serve              # Preview at localhost:4000
fazt deploy             # Ship it
```

**Note:** The welcome post and index page should be fun and delightful -
something that makes users smile when they first run `serve`. Design TBD.

## Directory Structure

```
my-site/
├── _config.yml           # Site configuration
├── _posts/               # Blog posts (processed by SSG)
│   └── 2025-01-15-hello-world.md
├── _layouts/             # Page templates
│   ├── default.html
│   └── post.html
├── _includes/            # Reusable snippets
│   └── nav.html
├── _data/                # Data files (YAML/JSON)
│   └── menu.yml
├── assets/               # Static files (CSS, JS, images)
│   └── style.css
├── index.md              # Homepage
└── about.md              # Static page
```

### Mixing Static and SSG Content

Regular HTML folders work alongside Jekyll content:

```
my-site/
├── _posts/               # SSG processes these
├── _layouts/
├── index.md              # SSG processes this
├── wedding-invite/       # Regular static folder - served as-is
│   ├── index.html        # → /wedding-invite/
│   └── style.css         # → /wedding-invite/style.css
└── apps/
    └── calculator.html   # → /apps/calculator.html
```

**Detection rule:** SSG only activates when `_posts/` or `_config.yml` exists.
It processes `_posts/*.md` and `*.md` files with frontmatter. Everything else
(HTML, CSS, JS, images, folders) is served unchanged.

## Writing Posts

Create files in `_posts/` with format `YYYY-MM-DD-slug.md`:

```markdown
---
title: My Post Title
description: A brief summary
date: 2025-01-03
tags: [tech, tutorial]
---

Your content in **markdown** here.

## Subheading

More content...
```

### Frontmatter

| Field         | Type     | Description                   |
|---------------|----------|-------------------------------|
| `title`       | string   | Post title (required)         |
| `description` | string   | Summary for SEO/previews      |
| `date`        | date     | Publish date                  |
| `tags`        | []string | Categorization tags           |
| `author`      | string   | Author name                   |
| `layout`      | string   | Template (default: `post`)    |
| `published`   | bool     | Set `false` for drafts        |
| `image`       | string   | Featured image path           |

### Drafts

Keep unpublished posts in `_drafts/`:

```
_drafts/
└── upcoming-feature.md    # No date prefix needed
```

Preview drafts locally:
```bash
fazt ssg serve --drafts
```

## Static Pages

Any `.md` or `.html` file outside `_posts/` becomes a page:

```markdown
---
title: About Me
layout: page
---

I'm a developer who loves building things.
```

Routes are derived from file paths:
- `about.md` → `/about/`
- `docs/getting-started.md` → `/docs/getting-started/`
- `contact.html` → `/contact/`

## Layouts

Layouts wrap your content. Create in `_layouts/`:

```html
<!-- _layouts/default.html -->
<!DOCTYPE html>
<html>
<head>
  <title>{{ page.title }} | {{ site.title }}</title>
  <link rel="stylesheet" href="/assets/style.css">
</head>
<body>
  {% include nav.html %}
  <main>
    {{ content }}
  </main>
</body>
</html>
```

```html
<!-- _layouts/post.html -->
{% layout default %}

<article>
  <h1>{{ page.title }}</h1>
  <time>{{ page.date | date: "%B %d, %Y" }}</time>
  <div class="content">
    {{ content }}
  </div>
  <div class="tags">
    {% for tag in page.tags %}
      <span class="tag">{{ tag }}</span>
    {% endfor %}
  </div>
</article>
```

### Layout Inheritance

Layouts can extend other layouts:

```html
<!-- _layouts/post.html -->
{% layout default %}
<!-- This layout's content replaces {{ content }} in default.html -->
```

## Includes

Reusable snippets in `_includes/`:

```html
<!-- _includes/nav.html -->
<nav>
  <a href="/">Home</a>
  <a href="/about/">About</a>
  <a href="/blog/">Blog</a>
</nav>
```

Use in layouts or pages:
```html
{% include nav.html %}

<!-- With parameters -->
{% include nav.html active="home" %}
```

## Templating

Simple Liquid-style syntax:

### Variables

```html
{{ page.title }}          <!-- Page variable -->
{{ site.title }}          <!-- Site config -->
{{ content }}             <!-- Rendered content -->
```

### Conditionals

```html
{% if page.image %}
  <img src="{{ page.image }}" alt="{{ page.title }}">
{% endif %}

{% unless page.draft %}
  <article>...</article>
{% endunless %}
```

### Loops

```html
<!-- List all posts -->
{% for post in site.posts %}
  <a href="{{ post.url }}">{{ post.title }}</a>
{% endfor %}

<!-- With limit -->
{% for post in site.posts limit:5 %}
  ...
{% endfor %}

<!-- Loop variables -->
{% for post in site.posts %}
  {% if forloop.first %}<ul>{% endif %}
  <li>{{ post.title }}</li>
  {% if forloop.last %}</ul>{% endif %}
{% endfor %}
```

### Filters

```html
<!-- Dates -->
{{ page.date | date: "%Y-%m-%d" }}
{{ page.date | date: "%B %d, %Y" }}

<!-- Strings -->
{{ page.title | upcase }}
{{ page.title | truncate: 50 }}
{{ page.url | prepend: site.baseurl }}

<!-- Arrays -->
{{ site.posts | size }}
{{ site.posts | first }}
{{ site.posts | where: "tags", "tech" }}
```

## Configuration

### URL Handling

Unlike Jekyll, you usually don't need to configure `url` or `baseurl`:

- **Local dev**: `fazt ssg serve` serves at `http://localhost:4000`
- **Production**: Fazt injects the correct URL at build/deploy time

```bash
# Build for specific URL (optional)
fazt ssg build --url https://myblog.example.com

# Or set in config for custom domains
```

If you need explicit control:

```yaml
# _config.yml
url: https://myblog.example.com   # Only if using custom domain
baseurl: ""                        # Almost always empty in Fazt
```

Links in templates work without prefixes:
```html
<a href="/about/">About</a>        <!-- Just works -->
<a href="{{ post.url }}">...</a>   <!-- Just works -->
```

### Full Config

`_config.yml`:

```yaml
# Site info
title: My Blog
description: Thoughts on technology
url: https://example.com

# Author
author:
  name: Your Name
  email: you@example.com

# Build
permalink: /:year/:month/:title/   # URL structure
paginate: 10                       # Posts per page

# Features (built-in)
feed: true      # Generate RSS at /feed.xml
sitemap: true   # Generate /sitemap.xml
seo: true       # Auto meta tags

# Exclude from build
exclude:
  - README.md
  - node_modules
```

### Permalink Patterns

| Pattern                      | Example URL                |
|------------------------------|----------------------------|
| `/:title/`                   | `/my-post/`                |
| `/:year/:month/:title/`      | `/2025/01/my-post/`        |
| `/:year/:month/:day/:title/` | `/2025/01/03/my-post/`     |
| `/blog/:title/`              | `/blog/my-post/`           |

## Data Files

Store structured data in `_data/`:

```yaml
# _data/team.yml
- name: Alice
  role: Developer
  github: alice

- name: Bob
  role: Designer
  github: bob
```

Access in templates:
```html
{% for member in site.data.team %}
  <div class="member">
    <h3>{{ member.name }}</h3>
    <p>{{ member.role }}</p>
  </div>
{% endfor %}
```

## Built-in Features

### RSS Feed

Enabled by default. Generates `/feed.xml`:

```yaml
# _config.yml
feed: true
```

### Sitemap

Auto-generates `/sitemap.xml`:

```yaml
sitemap: true
```

### SEO Tags

Auto-injects meta tags:

```yaml
seo: true
```

In your layout:
```html
<head>
  {% seo %}
</head>
```

Generates:
```html
<title>Post Title | Site Title</title>
<meta name="description" content="...">
<meta property="og:title" content="...">
<meta property="og:image" content="...">
<link rel="canonical" href="...">
```

### Syntax Highlighting

Code blocks are automatically highlighted:

````markdown
```javascript
function hello() {
  console.log("Hello!");
}
```
````

## Profiles

> **See `specs/v0.11-distribution/profiles.md`** for the complete deployment
> profile system. This section covers SSG-specific aspects only.

SSG profiles are a subset of the broader deployment profile system:

| Profile     | Build Method    | Notes                        |
|-------------|-----------------|------------------------------|
| `jekyll`    | Internal        | Built by Fazt (jekyll-lite)  |
| `docusaurus`| External        | Built by GitHub Actions/Modal|
| `vitepress` | External        | Built by GitHub Actions/Modal|
| `docsify`   | None            | Client-side rendering        |
| `static`    | None            | Serve as-is                  |

**Internal vs External builds:**
- `jekyll`: Fazt builds using embedded jekyll-lite (pure Go)
- `docusaurus/vitepress`: Require npm, built externally via `dev.build`

### SSG Configuration (app.json)

```json
{
  "name": "my-docs",
  "profile": "jekyll",
  "ssg": {
    "drafts": false,
    "future": false
  }
}
```

### MDX Graceful Degradation

For Docusaurus/VitePress MDX files, Fazt gracefully degrades:
- Strip JSX components (can't run React on server)
- Render standard markdown
- Show placeholder for custom components

---

**Note: UX Flow Modeling Needed**

The interaction between these specs needs deeper design work:

1. **Detection priority**: When does profile detection happen? At `deploy` time?
   At `serve` time? Both?

2. **Build triggering**: If external build needed, when is user prompted? Can it
   be fully automatic?

3. **Local vs remote**: `fazt serve` runs locally. Does it also trigger external
   builds? Or only use local artifacts?

4. **Error states**: What happens when detection is ambiguous? When build fails?
   When profile is unsupported?

5. **Configuration cascade**: `app.json` profile vs auto-detection vs CLI flag -
   what wins?

These questions need UX flow diagrams and user journey mapping before
implementation.

## CLI Commands

```bash
# Top-level aliases (most common)
fazt serve                            # Auto-detect profile, serve locally
fazt deploy                           # Deploy source, server builds

# Namespaced commands (full control)
fazt ssg init                         # Scaffold minimal Jekyll blog
fazt ssg serve [--port] [--drafts]    # Local preview (explicit)
fazt ssg build [--destination]        # Build to local folder
fazt ssg profiles                     # List supported profiles
```

`fazt serve` and `fazt deploy` are convenience aliases that auto-detect
the profile. Use namespaced commands when you need explicit control.

## Example Apps & Plugins

The Fazt store (`github.com/fazt-sh/store/`) contains curated example apps.
Each is crafted to be beautiful, purposeful, and educational.

### Haikus (Flagship Blog Example)

A collection of Japanese haikus with meanings. Showcases all SSG capabilities:

```
github.com/fazt-sh/store/haikus/
├── _config.yml
├── _posts/
│   ├── 2025-01-01-old-pond.md      # Bashō's famous frog poem
│   └── ...
├── _layouts/
│   └── haiku.html                   # Elegant minimal layout
├── _plugins/
│   ├── emoji.js                     # :cherry_blossom: → 🌸
│   ├── reading-time.js              # "1 min read"
│   ├── archives.js                  # Tag/season archives
│   ├── vertical-text.js             # Traditional vertical rendering
│   └── romanji.js                   # Japanese → romanji filter
└── assets/
    └── style.css                    # Beautiful typography
```

Install it:
```bash
fazt deploy github.com/fazt-sh/store/haikus
```

The `repo/folder` pattern works with any GitHub repo - projects can include
Fazt apps in subfolders alongside their main code:
```bash
fazt deploy github.com/someproject/repo/examples/fazt-demo
```

Or copy just the plugins you need:
```bash
curl -O https://raw.githubusercontent.com/fazt-sh/store/main/haikus/_plugins/emoji.js
mv emoji.js _plugins/
```

### Learning from Examples

Each store app demonstrates different capabilities:

| App         | Showcases                                    |
|-------------|----------------------------------------------|
| `haikus`    | Blog plugins, elegant design, custom filters |
| `devlog`    | Code embedding, gists, syntax highlighting   |
| `portfolio` | Static pages, image galleries                |
| `docs`      | Multi-section navigation, search             |

Browse, install, or copy parts. Learn by reading real code.

## Plugins

Extend with JavaScript plugins in `_plugins/`:

```javascript
// _plugins/reading-time.js
fazt.filter('reading_time', (content) => {
  const words = content.split(/\s+/).length;
  return Math.ceil(words / 200) + ' min read';
});
```

```html
<span class="meta">{{ page.content | reading_time }}</span>
```

### Custom Tags

```javascript
// _plugins/youtube.js
fazt.tag('youtube', (args) => {
  const id = args[0];
  return `
    <div class="video">
      <iframe src="https://youtube.com/embed/${id}"
              allowfullscreen></iframe>
    </div>
  `;
});
```

```markdown
{% youtube dQw4w9WgXcQ %}
```

## Mixing Static + Dynamic

Static sites can include dynamic `.ejs` pages:

```
my-site/
├── _posts/           # Static blog posts
├── _layouts/
├── index.md          # Static homepage
├── about.md          # Static page
└── contact.ejs       # Dynamic form handler
```

The static content is pre-built, while `.ejs` pages run server-side.

## Jekyll Compatibility

Fazt's SSG is powered by jekyll-lite, so existing Jekyll sites work:

```bash
# Deploy existing Jekyll site
cd my-jekyll-blog
fazt deploy
```

**What works:**
- `_posts`, `_layouts`, `_includes`, `_data`
- Liquid templating (core subset)
- Common filters and tags
- Frontmatter

**What doesn't:**
- Ruby plugins (use JS instead)
- Sass/SCSS (use external tool)
- Collections (only `_posts`)

See `specs/modules/jekyll-lite.md` for compatibility details.

## Example: Minimal Blog

```
my-blog/
├── _config.yml
├── _layouts/
│   └── default.html
├── _posts/
│   └── 2025-01-03-hello.md
└── index.html
```

`_config.yml`:
```yaml
title: My Blog
```

`_layouts/default.html`:
```html
<!DOCTYPE html>
<html>
<head><title>{{ page.title }}</title></head>
<body>
  {{ content }}
</body>
</html>
```

`_posts/2025-01-03-hello.md`:
```markdown
---
title: Hello World
layout: default
---
My first post!
```

`index.html`:
```html
---
layout: default
title: Home
---
<h1>Posts</h1>
{% for post in site.posts %}
  <a href="{{ post.url }}">{{ post.title }}</a>
{% endfor %}
```

Deploy:
```bash
fazt deploy
```

Done. Your blog is live.
