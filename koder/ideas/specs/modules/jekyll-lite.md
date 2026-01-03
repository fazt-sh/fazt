# Jekyll-lite

## Summary

Minimal Jekyll-compatible static site generator in pure Go. Supports the core
Jekyll conventions (`_posts`, `_layouts`, `_includes`, Liquid templates) without
Ruby dependencies. Designed as extractable module for use in Fazt and standalone.

## Goals

1. **Zero-config Jekyll sites**: Deploy existing Jekyll blogs without changes
2. **Pure Go**: No Ruby, no CGO, embeddable in single binary
3. **Extractable**: Separate module, usable outside Fazt
4. **Plugin-ready**: Optional JS plugins via Goja for extensibility

## Non-Goals

- 100% Jekyll compatibility (we target ~80% of real-world usage)
- Ruby plugin support (JS alternative provided)
- Complex features: pagination v2, collections beyond posts

## Directory Convention

```
my-blog/
├── _config.yml           # Site configuration
├── _posts/               # Blog posts (date-prefixed markdown)
│   ├── 2025-01-15-hello.md
│   └── 2025-03-27-world.md
├── _drafts/              # Unpublished posts (optional)
│   └── upcoming-post.md
├── _layouts/             # Page templates (Liquid)
│   ├── default.html
│   ├── post.html
│   └── page.html
├── _includes/            # Reusable partials
│   ├── head.html
│   ├── header.html
│   └── footer.html
├── _plugins/             # JS plugins (optional, Fazt extension)
│   └── reading-time.js
├── _data/                # Data files (YAML/JSON)
│   └── navigation.yml
├── assets/               # Static files (copied as-is)
│   ├── css/
│   ├── js/
│   └── images/
├── index.html            # Homepage (or index.md)
├── about.md              # Static page
└── feed.xml              # RSS template (optional)
```

## Post Format

Filename: `YYYY-MM-DD-slug.md`

```markdown
---
layout: post
title: My First Post
description: A brief intro
date: 2025-03-27
author: Jikku Jose
tags: [tech, ai]
categories: [blog]
published: true
---

Post content in **markdown** here.
```

### Frontmatter Fields

| Field         | Type     | Description                    |
|---------------|----------|--------------------------------|
| `layout`      | string   | Layout template to use         |
| `title`       | string   | Post/page title                |
| `description` | string   | Meta description               |
| `date`        | date     | Publish date (overrides file)  |
| `author`      | string   | Author name                    |
| `tags`        | []string | Tags for categorization        |
| `categories`  | []string | Categories                     |
| `published`   | bool     | false = draft                  |
| `permalink`   | string   | Custom URL path                |
| `*`           | any      | Custom fields accessible       |

## Liquid Templating (Subset)

### Variables

```liquid
{{ page.title }}
{{ page.date }}
{{ page.content }}
{{ site.title }}
{{ site.posts }}
{{ content }}
```

### Objects Available

| Object        | Description                          |
|---------------|--------------------------------------|
| `site`        | Site-wide config and collections     |
| `page`        | Current page/post data               |
| `content`     | Rendered content (in layouts)        |
| `paginator`   | Pagination data (if enabled)         |

### Site Object

```liquid
{{ site.title }}
{{ site.description }}
{{ site.url }}
{{ site.baseurl }}
{{ site.posts }}        # All posts, newest first
{{ site.pages }}        # All pages
{{ site.data.nav }}     # Data from _data/nav.yml
{{ site.time }}         # Build time
```

### Control Flow

```liquid
{% if page.title %}
  <h1>{{ page.title }}</h1>
{% endif %}

{% if page.tags contains "tech" %}
  <span class="tech-post">
{% endif %}

{% unless page.draft %}
  Published content
{% endunless %}
```

### Loops

```liquid
{% for post in site.posts %}
  <a href="{{ post.url }}">{{ post.title }}</a>
{% endfor %}

{% for post in site.posts limit:5 %}
  ...
{% endfor %}

{% for post in site.posts offset:5 %}
  ...
{% endfor %}

{% for tag in page.tags %}
  <span>{{ tag }}</span>
{% endfor %}
```

### Loop Variables

```liquid
{% for post in site.posts %}
  {{ forloop.index }}      # 1, 2, 3...
  {{ forloop.index0 }}     # 0, 1, 2...
  {{ forloop.first }}      # true on first
  {{ forloop.last }}       # true on last
  {{ forloop.length }}     # total items
{% endfor %}
```

### Includes

```liquid
{% include head.html %}
{% include nav.html active="home" %}
{% include post-card.html post=post %}
```

### Filters

**String filters:**
```liquid
{{ "hello" | upcase }}              # HELLO
{{ "HELLO" | downcase }}            # hello
{{ "hello" | capitalize }}          # Hello
{{ "  hello  " | strip }}           # hello
{{ "hello" | truncate: 3 }}         # hel...
{{ "hello" | truncatewords: 2 }}    # hello...
{{ "a,b,c" | split: "," }}          # ["a","b","c"]
{{ page.url | prepend: site.baseurl }}
{{ page.url | append: ".html" }}
{{ text | escape }}                 # HTML escape
{{ text | strip_html }}             # Remove HTML tags
{{ text | newline_to_br }}          # \n → <br>
{{ text | markdownify }}            # Markdown → HTML
```

**Date filters:**
```liquid
{{ page.date | date: "%Y-%m-%d" }}
{{ page.date | date: "%B %d, %Y" }}
{{ page.date | date_to_xmlschema }}
{{ page.date | date_to_rfc822 }}
{{ page.date | date_to_string }}
```

**Array filters:**
```liquid
{{ site.posts | size }}
{{ site.posts | first }}
{{ site.posts | last }}
{{ site.posts | sort: "title" }}
{{ site.posts | reverse }}
{{ site.posts | where: "category", "tech" }}
{{ site.posts | group_by: "category" }}
{{ tags | join: ", " }}
```

**Number filters:**
```liquid
{{ 4.5 | round }}
{{ 4 | plus: 2 }}
{{ 4 | minus: 2 }}
{{ 4 | times: 2 }}
{{ 4 | divided_by: 2 }}
```

## Config File

`_config.yml`:

```yaml
title: My Blog
description: A personal blog
url: https://example.com
baseurl: ""

# Author defaults
author:
  name: Jikku Jose
  email: jikku@example.com

# Build settings
markdown: goldmark
permalink: /:year/:month/:day/:title/

# Pagination (optional)
paginate: 10
paginate_path: /page/:num/

# Exclude from processing
exclude:
  - README.md
  - Gemfile
  - node_modules

# Include hidden files
include:
  - .htaccess

# Plugins to enable (built-in)
plugins:
  - feed          # RSS/Atom feed
  - sitemap       # sitemap.xml
  - seo           # Meta tags

# Custom variables
social:
  twitter: jikkujose
  github: jikkujose
```

## Built-in Features (No Plugins)

### RSS Feed

Automatically generated at `/feed.xml` when `feed` plugin enabled:

```yaml
plugins:
  - feed
```

Or create custom `feed.xml`:

```liquid
---
layout: null
---
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>{{ site.title }}</title>
    {% for post in site.posts limit:10 %}
    <item>
      <title>{{ post.title | escape }}</title>
      <link>{{ post.url | prepend: site.url }}</link>
      <pubDate>{{ post.date | date_to_rfc822 }}</pubDate>
    </item>
    {% endfor %}
  </channel>
</rss>
```

### Sitemap

Automatically generated at `/sitemap.xml` when enabled:

```yaml
plugins:
  - sitemap
```

### SEO Tags

Auto-inject meta tags when enabled:

```yaml
plugins:
  - seo
```

Then in layout:
```liquid
{% seo %}
```

Generates:
```html
<title>Post Title | Site Title</title>
<meta name="description" content="...">
<meta property="og:title" content="...">
<meta property="og:description" content="...">
<link rel="canonical" href="...">
```

## JS Plugins (Fazt Extension)

Optional Goja-powered plugins for extensibility.

### Plugin Types

| Type       | Purpose                    | Example                    |
|------------|----------------------------|----------------------------|
| `filter`   | Custom Liquid filters      | `reading_time`             |
| `tag`      | Custom Liquid tags         | `{% youtube id %}`         |
| `generator`| Create pages dynamically   | Tag archive pages          |
| `hook`     | Lifecycle callbacks        | Post-process HTML          |

### Filter Plugin

`_plugins/reading-time.js`:

```javascript
// Register a custom filter
jekyll.filter('reading_time', function(content) {
  const words = content.split(/\s+/).length;
  const minutes = Math.ceil(words / 200);
  return minutes + ' min read';
});
```

Usage:
```liquid
{{ page.content | reading_time }}
```

### Tag Plugin

`_plugins/youtube.js`:

```javascript
// Register a custom tag
jekyll.tag('youtube', function(args) {
  const id = args[0];
  return `<iframe src="https://youtube.com/embed/${id}"
          allowfullscreen></iframe>`;
});
```

Usage:
```liquid
{% youtube dQw4w9WgXcQ %}
```

### Generator Plugin

`_plugins/tag-pages.js`:

```javascript
// Generate a page for each tag
jekyll.generator('tag_pages', function(site) {
  const tags = new Set();

  site.posts.forEach(post => {
    (post.tags || []).forEach(tag => tags.add(tag));
  });

  return Array.from(tags).map(tag => ({
    url: `/tags/${tag}/`,
    layout: 'tag',
    title: `Posts tagged "${tag}"`,
    tag: tag,
    posts: site.posts.filter(p => (p.tags || []).includes(tag))
  }));
});
```

### Hook Plugin

`_plugins/minify.js`:

```javascript
// Post-process rendered HTML
jekyll.hook('post_render', function(page) {
  if (page.output_ext === '.html') {
    // Simple minification
    page.content = page.content
      .replace(/\s+/g, ' ')
      .replace(/>\s+</g, '><');
  }
  return page;
});
```

### Plugin API

```javascript
// Available in plugins
jekyll.filter(name, fn)           // Register filter
jekyll.tag(name, fn)              // Register tag
jekyll.generator(name, fn)        // Register generator
jekyll.hook(event, fn)            // Register hook

// Hook events
'pre_render'                      // Before Liquid processing
'post_render'                     // After HTML generation
'post_write'                      // After file written

// Site data access
jekyll.site                       // Site config and collections
jekyll.site.posts                 // All posts
jekyll.site.pages                 // All pages
jekyll.site.data                  // _data files
```

## CLI Interface

### Standalone Usage

```bash
# Build site
jekyll-lite build [--source .] [--destination _site]

# Serve with live reload
jekyll-lite serve [--port 4000] [--drafts]

# New post
jekyll-lite new post "My New Post"

# Validate site
jekyll-lite doctor
```

### Fazt Integration

```bash
# Deploy Jekyll site (auto-detected)
fazt deploy ./my-blog

# Local development
fazt ssg serve --drafts

# Explicit build
fazt ssg build
```

## Build Process

1. **Load config**: Parse `_config.yml`
2. **Scan content**: Find posts, pages, static files
3. **Parse frontmatter**: Extract metadata from each file
4. **Load data**: Parse `_data/*.yml` files
5. **Run generators**: Execute generator plugins
6. **Render content**: Markdown → HTML
7. **Apply layouts**: Wrap content in templates
8. **Process Liquid**: Variable substitution, loops, etc.
9. **Run hooks**: Post-processing
10. **Write output**: Generate `_site/` directory

## Fazt VFS Integration

When deployed to Fazt, Jekyll-lite outputs directly to VFS:

```go
// Instead of writing to _site/
builder.SetOutput(fazt.VFS(appID))

// Posts become routes automatically
// /2025/03/27/my-post/ → served from VFS
```

## Implementation Notes

### Module Structure

```
github.com/anthropic/jekyll-lite/
├── cmd/
│   └── jekyll-lite/      # CLI binary
├── pkg/
│   ├── config/           # Config parsing
│   ├── content/          # Post/page loading
│   ├── liquid/           # Liquid template engine
│   ├── markdown/         # Markdown rendering
│   ├── plugins/          # Plugin system
│   └── builder/          # Build orchestration
├── internal/
│   └── filters/          # Built-in Liquid filters
└── go.mod
```

### Dependencies

| Dependency       | Purpose              | Notes              |
|------------------|----------------------|--------------------|
| `goldmark`       | Markdown rendering   | Pure Go            |
| `yaml.v3`        | YAML parsing         | Pure Go            |
| `goja`           | JS plugins           | Pure Go, optional  |
| (custom)         | Liquid templating    | ~1000 lines        |

### Binary Impact

- Core (no plugins): ~500KB
- With Goja (plugins): ~2MB additional

## Compatibility Matrix

| Feature                 | Jekyll | Jekyll-lite | Notes              |
|-------------------------|--------|-------------|--------------------|
| `_posts` convention     | ✓      | ✓           | Full support       |
| `_layouts`              | ✓      | ✓           | Full support       |
| `_includes`             | ✓      | ✓           | Full support       |
| `_data` (YAML/JSON)     | ✓      | ✓           | Full support       |
| `_drafts`               | ✓      | ✓           | Full support       |
| Liquid basics           | ✓      | ✓           | Core subset        |
| Common filters          | ✓      | ✓           | ~30 filters        |
| Markdown (kramdown)     | ✓      | ~           | goldmark instead   |
| Syntax highlighting     | ✓      | ✓           | chroma             |
| Pagination              | ✓      | ✓           | Basic v1 style     |
| Collections             | ✓      | ✗           | Posts only         |
| Ruby plugins            | ✓      | ✗           | JS alternative     |
| Sass/SCSS               | ✓      | ✗           | Use external tool  |
| CoffeeScript            | ✓      | ✗           | Not supported      |
| Incremental builds      | ✓      | ✗           | Full rebuild only  |

## Migration Guide

### From Jekyll

1. Remove `Gemfile`, `Gemfile.lock`
2. Remove Ruby-specific plugins from `_config.yml`
3. Convert Ruby plugins to JS (if any custom)
4. Replace kramdown-specific syntax with standard markdown
5. Test with `jekyll-lite build`
6. Deploy: `fazt deploy ./my-blog`

### Common Issues

**Kramdown syntax:**
```markdown
# Jekyll (kramdown)
{:.my-class}
Paragraph with class

# Jekyll-lite (goldmark)
<p class="my-class">Paragraph with class</p>
```

**Ruby plugins:**
```ruby
# Jekyll Ruby plugin
module Jekyll
  module ReadingTimeFilter
    def reading_time(input)
      words = input.split.size
      (words / 200.0).ceil.to_s + " min"
    end
  end
end
Liquid::Template.register_filter(Jekyll::ReadingTimeFilter)
```

```javascript
// Jekyll-lite JS equivalent
jekyll.filter('reading_time', function(input) {
  const words = input.split(/\s+/).length;
  return Math.ceil(words / 200) + ' min';
});
```

## Future Considerations

- **Collections**: Support arbitrary collections beyond posts
- **Incremental**: Only rebuild changed files
- **Themes**: Package themes as modules
- **Data cascade**: Nested data file inheritance
