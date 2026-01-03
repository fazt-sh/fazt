# Static Site Generator

## Summary

Built-in static site generator for blogs, documentation, and content sites.
Uses battle-tested Jekyll-style conventions. Powered by jekyll-lite module
for compatibility with existing Jekyll sites.

## Why

Static sites are the simplest deployment model:
- Fast (pre-rendered HTML)
- Secure (no server-side code)
- Cheap (just file serving)

Fazt should make static sites trivial while supporting dynamic features when
needed. The Jekyll convention is well-known, well-documented, and has thousands
of existing themes and sites.

**Philosophy alignment:**
- Single binary: SSG is ~500KB additional, pure Go
- Zero-build: No npm/bundler, just markdown files
- Convention over configuration: Sensible defaults

## Quick Start

```bash
mkdir my-blog && cd my-blog
fazt ssg init           # Creates minimal working blog
fazt ssg serve          # Preview at localhost:4000
```

Open http://localhost:4000 - you have a blog.

```bash
fazt ssg build          # Build to _site/
fazt deploy ./_site     # Ship it
```

### What `init` Creates

```
my-blog/
├── _config.yml         # Site title, URL
├── _layouts/
│   └── default.html    # Minimal layout
├── _posts/
│   └── 2025-01-03-welcome.md
├── assets/
│   └── favicon.svg     # Simple favicon
└── index.html          # Post listing
```

Minimal, readable files. Learn by modifying them.

**Note:** The welcome post and index page should be fun and delightful -
something that makes users smile when they first run `serve`. Design TBD.

## Directory Structure

```
my-site/
├── _config.yml           # Site configuration
├── _posts/               # Blog posts
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

## CLI Commands

```bash
fazt ssg init                         # Scaffold minimal blog
fazt ssg serve [--port 4000] [--drafts]  # Local preview
fazt ssg build [--destination _site]  # Build static files
fazt deploy ./_site                   # Deploy built output
```

## Themes

Themes are just `_layouts/` + `_includes/` + `assets/`. Clone or copy any
Jekyll-compatible theme:

```bash
# Clone a theme as starting point
git clone https://github.com/example/minimal-theme my-blog
cd my-blog
fazt ssg serve
```

Theme structure:
```
my-theme/
├── _layouts/
│   ├── default.html
│   ├── post.html
│   └── page.html
├── _includes/
│   ├── head.html
│   ├── nav.html
│   └── footer.html
└── assets/
    └── style.css
```

## Plugins (Optional)

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
