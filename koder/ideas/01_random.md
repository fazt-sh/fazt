# Fazt.sh Roadmap

- [ ] 01 `core-backup`
  - CLI command `fazt backup create`
  - Snapshot `data.db` to backup file
  - CLI command `fazt backup restore`
  - Disaster recovery for single-db systems

- [ ] 02 `core-upgrade`
  - Self-update via `fazt upgrade`
  - Download latest binary from GitHub
  - Retain `setcap` permissions (port 80/443)
  - Restart systemd service post-update

- [x] 03 `core-cache`
  - In-memory LRU cache for VFS
  - Store frequently accessed files (css/js)
  - Invalidate cache on new deployment
  - Improve latency and reduce DB IO

- [x] 04 `core-config-db`
  - Migrate `config.json` to SQLite table
  - Remove dependency on filesystem config
  - Achieve true "One Binary + One DB" state

- [x] 05 `ops-install`
  - One-line installer: `curl | bash`
  - Automate user creation & systemd setup
  - Host `install.sh` at `fazt.sh/install`

- [ ] 06 `js-cron` <next-to-implement>
  - Scheduled serverless functions
  - Define triggers in `fazt.json`
  - Support intervals (e.g., "every 1h")
  - Enable background data syncing

- [ ] 07 `app-drop`
  - Ephemeral file sharing application
  - Upload files -> Share Link -> Auto-expire
  - Built-in drag-and-drop web interface
  - "Snapdrop" clone for personal use

- [ ] 08 `app-pad`
  - Collaborative real-time scratchpad
  - WebSocket-based text synchronization
  - Universal clipboard across devices

- [ ] 09 `app-textdb`
  - Simple NoSQL-like JSON store
  - HTTP API: `POST /db/:collection`
  - Backend for static site forms/data

- [ ] 10 `proto-webdav`
  - Expose VFS via WebDAV protocol
  - Mount Fazt storage as local disk
  - Direct file management from OS

- [ ] 11 `proto-email`
  - Receive emails via SMTP (port 25)
  - Store messages in `inbox` SQL table
  - Send via relay (Postmark/SES/etc)

- [x] 12 `route-reserved`
  - Dashboard shouldn't be in root, but in admin.<DOMAIN>
  - Global `404.*` for missing sites (so we can just change 404 site to provide
    universal 404)
  - Root domain content strategy, it should point to root.<domain> ; which
    should be a "Welcome to Fazt" site, make it fun

- [x] 13 `meta-locations`
  - Update official URLs and docs
  - Lander: `fazt.sh`
  - Social: `x.com/fazt_sh`

- [x] 14 `gh-pages`
  - do a simple landing page
  - have the install script from there
  - ideally: https://fazt.github.io/ or if only this is possible:
    https://fazt-sh.github.io

- [ ] 15 `oauth` <to-discuss>
  - since we have persistent DB, should auth be possible?
  - if so, metamask & google auth should technically be possible for sites that
    need it? (calendar.zyt.app could have the ability to ask users to signin?)

- [ ] 16 `encrypted_content` <to-discuss>
  - consider a note takin app with fazt
  - can't it use metamask login and encrypt the data it stores using say the
    public key, so even if the DB gets compromised its impossible to compromise
    the data? But the webapp can decrypt the data with the user's private key
    and use for local manipulation and encrypt back while saving?

- [ ] 17 `app-store` <to-discuss>
  - can't build apps in git repos in the following format:
    ```
    static-webapp/
      <files>
    api/
      <serverless-functions>
    ```
  - so that we can have a "fazt" site to be able to install this app (repo); so
    that it will be cloned & be assigned a subdomain to work
  - hence we can do something like: have a todo app in
    https://github.com/abc/todo ; which follows that structure; and we just need
    to go to the site console (or cli) and provide the repo + subdomain; to have
    the app "installed" into that subdomain, with proper persistence (and even
    login if above auth idea works?)

- [ ] 18 `fazt-meet` <to-discuss>
  - can't we have a video conferencing app through fazt, as we can setup a
    stunserver?

- [ ] 19 `app-ideas` <to-discuss>
  - fazt meet
  - calendar
  - todo
  - docs
  - sheets
  - kanban / trello
  - micro blogging? / twitter
  - YT
  - albums

- [ ] 20 `ai-sdk` <to-discuss>
  - can't we compile the vercel ai sdk into a JS bundle and make available
    inside the serverless handlers?
  - won;t this ai.() capability be super handy to build many ideas?

- [ ] 21 `local-dev` <to-discuss>
  - since the executable is super easy to use, can't we use it locally for
    developing the "apps" and then publish?

- [ ] 22 `full-gui-capability` <to-discuss>
  - implement all management post installation be handled via GUI too

- [ ] 23 `static-server` <to-discuss>
  - do we need to replicate the sql db stored files in a location in server too,
    so that data is definitely in DB, but the system can choose to serve flat
    files, if its more efficient? Say for streaming/large blobs?
    
- [ ] 24 `db-sync` <to-discuss>
  - is there a possibility to "sync" data between different redundant fazt servers? 
  - use case: a user can use the fazt server locally in his computer & have a
    remote instance, if files sync in the background; the person gets a live
    system where ever he goes transparently; while having a great local version

- [ ] 25 `qr-code-setup` <to-discuss>
  - is it possible to show a QR code option post install to collect the user input?
  - something like <IP>/5-char-random-string/ will be QRcode; it will collect:
    domain, username, password & email & finish the setup?

- [ ] 26 `cloudflare cdn` <to-discuss>
  - is it better to have cloudflare cdn as a front for the site

- [ ] 27 `simplify-client-auth-token` <to-discuss>
  - shall we encode domain & authtoken to a long string, say base64 so that it
    can be simply pasted to get autheticated without having to specify the site?

- [ ] 28 `telegram-bot-server` <to-discuss>
  - what about a telegram bot server so that it is easy to set up a telegram bot
    easily

- [ ] 29 `multiple-domains` <to-discuss>
  - **Custom Domain Mapping**: Map any external domain (e.g., `blog.example.com`) to a Fazt site (`my-blog`) via a `domain_mappings` SQL table.
  - **Dynamic Routing**: Update the HTTP router to resolve incoming Host headers against the DB before falling back to subdomain logic.
  - **On-Demand HTTPS**: Configure CertMagic to allow certificate issuance for any domain found in the mapping table (Zero-Config SSL).
  - **White Labeling**: Enables "Agency Mode" where end-users interact with custom domains and never see the underlying `fazt.sh` infrastructure.

  - [ ] 30 `core-serverless-v2` <to-discuss>
  - **Structure**: Move serverless entry point to `api/main.js` to avoid
    conflicts with static frontend assets (e.g., client-side `js/main.js`).
  - **Imports**: Implement a `require()` shim to enable code splitting within
    the `api/` folder (e.g., `const db = require('./db.js')`).
  - **Standard Library**: Embed optimized, ES5-compatible builds of essential
    utilities (`lodash`, `cheerio`, `marked`, `uuid`) directly into binary.
  - **Zero-Build**: Allow Users and AI Agents to deploy powerful logic without
    `npm install` or bundlers by simply calling `require('lodash')`.
  - **Virtual Modules**: Runtime intercepts `require` calls for standard libs
    and serves pre-compiled sources from memory for sub-millisecond load.

- [ ] 31 `core-ai-shim` <to-discuss>
  - **Standardized Interface**: Embed a `require('fazt-ai')` module that
    normalizes request/response shapes for OpenAI, Anthropic, and Gemini.
  - **Env Var Integration**: Auto-load `OPENAI_API_KEY` from system env so
    user code doesn't need to handle secrets or config boilerplate.
  - **Streaming**: Support chunked responses to allow Agentic Apps to stream
    text back to the calling client or UI.

- [ ] 32 `proto-email-in` <to-discuss>
  - **SMTP Sink**: Bind to Port 25 to receive incoming emails (Requires VPS
    provider to allow inbound traffic).
  - **Routing**: Map private addresses (e.g., `agent-xyz@domain.com`) to
    specific serverless apps using the App UUID.
  - **Trigger**: Execute `api/main.js` with `{ event: 'email' }` payload.
  - **Use Cases**: Enable "Email-to-Agent" workflows like parsing receipts,
    summarizing newsletters, or triggering deploys via email.

- [ ] 33 `core-app-identity` <to-discuss>
  - **Stable UUIDs**: Decouple data storage from subdomains. Generate a unique
    `app_id` (e.g., `app_x9z2`) for every site creation.
  - **Private Routing**: Derive sensitive endpoints like Inbound Email
    (`x9z2@fazt.sh`) and Webhooks from the UUID to prevent public guessing.
  - **Data Resilience**: Allow renaming subdomains without losing KV Store
    data or logs, as they are keyed to the immutable `app_id`.
  - **Peer Access**: Enable "IPC" where App A can grant read access to its
    data bucket to App B via UUID references.

- [ ] 34 `core-marketplace` <to-discuss>
  - **Repositories**: Adopt "Linux Distro" model where Marketplaces are just
    Git URLs (e.g., `github.com/fazt-sh/store`) acting as package sources.
  - **Discovery**: Cache remote `registry.json` in DB to browse available
    apps locally without installing them (`apt-cache search` style).
  - **Installation**: `fazt app install <name>` fetches specific app assets
    from the Git Repo via Zip stream and hydrates them into the VFS.
  - **Personal**: Any app not from a Marketplace (CLI deploy, MCP push) is
    classified as "Personal" source.
  - **Manifest**: `app.json` defines metadata, versioning, and required `env`
    variables, enabling interactive installation wizards.
  - **Tracking**: Add `deployed_by` column to track the Actor (Admin CLI,
    MCP Agent Token) who installed the app.
  - **Updates**: `fazt app update` checks the `source_marketplace` URL for
    newer versions defined in `registry.json`.
