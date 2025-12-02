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

- [ ] 03 `core-cache`
  - In-memory LRU cache for VFS
  - Store frequently accessed files (css/js)
  - Invalidate cache on new deployment
  - Improve latency and reduce DB IO

- [ ] 04 `core-config-db`
  - Migrate `config.json` to SQLite table
  - Remove dependency on filesystem config
  - Achieve true "One Binary + One DB" state

- [ ] 05 `ops-install`
  - One-line installer: `curl | bash`
  - Automate user creation & systemd setup
  - Host `install.sh` at `fazt.sh/install`

- [ ] 06 `js-cron`
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

- [ ] 12 `route-reserved`
  - Special handling for `dashboard.*`
  - Global `404.*` for missing sites
  - Root domain content strategy

- [ ] 13 `meta-locations`
  - Update official URLs and docs
  - Lander: `fazt.sh`
  - Social: `x.com/fazt_sh`

- [ ] 14 `gh-pages`
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
