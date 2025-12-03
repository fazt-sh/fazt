document.addEventListener('DOMContentLoaded', () => {
    
    // ==========================================
    // 1. DATA: Terminal Flows (JSON Structure)
    // ==========================================
    const terminalData = {
        tabs: [
            {
                id: 'deploy',
                title: 'Deploy Site',
                active: true,
                flow: [
                    { type: 'input', text: 'fazt client deploy --path ./my-app --domain awesome' },
                    { type: 'output', text: "Deploying ./my-app to http://localhost:4698 as 'awesome'..." },
                    { type: 'output', text: "Zipped 12 files (15420 bytes)" },
                    { type: 'pause', duration: 800 },
                    { type: 'output', text: "✓ Deployment successful!", color: '#27c93f' },
                    { type: 'output', text: "  Site: http://awesome.fazt.sh" },
                    { type: 'output', text: "  Files: 12" },
                    { type: 'output', text: "  Size: 15420 bytes" },
                    { type: 'pause', duration: 5000 }
                ]
            },
            {
                id: 'install',
                title: 'Install Server',
                active: false,
                flow: [
                    { type: 'input', text: 'curl -s https://fazt-sh.github.io/fazt/install.sh | bash' },
                    { type: 'output', text: "________            _____" },
                    { type: 'output', text: "___  __/_____ ________  /_" },
                    { type: 'output', text: "__  /_ _  __ `/__  /_  __/" },
                    { type: 'output', text: "_  __/ / /_/ /__  /_/ /_" },
                    { type: 'output', text: "/_/    \__,_/ _____/\__/" },
                    { type: 'output', text: "  Single Binary PaaS & Analytics", color: '#888' },
                    { type: 'output', text: " " },
                    { type: 'output', text: "ℹ Downloading v0.6.0...", color: '#2da4ff' },
                    { type: 'output', text: "✓ Downloaded ./fazt", color: '#27c93f' },
                    { type: 'output', text: " " },
                    { type: 'output', text: "Select Installation Type:" },
                    { type: 'output', text: "1. Headless Server (Daemon)" },
                    { type: 'output', text: "2. Command Line Tool (Portable)" },
                    { type: 'output', text: " " },
                    { type: 'input', text: '1', prompt: '> Select [1/2]: ' },
                    { type: 'output', text: " " },
                    { type: 'input', text: 'my-paas.com', prompt: 'Domain or IP: ' },
                    { type: 'input', text: 'admin@my-paas.com', prompt: 'Email (for HTTPS): ' },
                    { type: 'output', text: " " },
                    { type: 'output', text: "✓ Fazt installed & running!", color: '#27c93f' },
                    { type: 'output', text: "  https://my-paas.com" },
                    { type: 'output', text: "  Login: https://admin.my-paas.com" },
                    { type: 'pause', duration: 5000 }
                ]
            }
        ]
    };

    // ==========================================
    // 2. COMPONENT: Terminal
    // ==========================================
    class TerminalComponent {
        constructor(containerId, data) {
            this.container = document.getElementById(containerId);
            this.data = data;
            this.currentTabId = data.tabs.find(t => t.active).id;
            this.timeouts = [];
            
            this.renderStructure();
            this.bindEvents();
            this.play(this.currentTabId);
        }

        renderStructure() {
            // Tabs Header
            const header = document.createElement('div');
            header.className = 'terminal-tabs-header';
            
            const tabsList = document.createElement('div');
            tabsList.className = 'terminal-tabs-list';
            
            this.data.tabs.forEach(tab => {
                const btn = document.createElement('button');
                btn.className = `tab-btn ${tab.active ? 'active' : ''}`;
                btn.dataset.id = tab.id;
                btn.textContent = tab.title;
                tabsList.appendChild(btn);
            });

            // Controls (Redo)
            const controls = document.createElement('div');
            controls.className = 'terminal-controls';
            controls.innerHTML = `<button class="control-btn" id="term-redo" title="Replay">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.3"/></svg>
                </button>`;

            header.appendChild(tabsList);
            header.appendChild(controls);

            // Terminal Window
            const window = document.createElement('div');
            window.className = 'terminal-window';
            
            const winHeader = document.createElement('div');
            winHeader.className = 'terminal-window-header';
            winHeader.innerHTML = `<div class="dots">
                    <div class="dot red"></div>
                    <div class="dot yellow"></div>
                    <div class="dot green"></div>
                </div>
                <div class="title">bash</div>`;

            const body = document.createElement('div');
            body.className = 'terminal-body';
            body.id = 'term-body';

            window.appendChild(winHeader);
            window.appendChild(body);

            this.container.innerHTML = '';
            this.container.appendChild(header);
            this.container.appendChild(window);
            
            this.bodyElement = body;
            this.tabButtons = tabsList.querySelectorAll('.tab-btn');
        }

        bindEvents() {
            // Tab Clicking
            this.tabButtons.forEach(btn => {
                btn.addEventListener('click', () => {
                    const id = btn.dataset.id;
                    if (id === this.currentTabId) return;
                    this.switchTab(id);
                });
            });

            // Redo Clicking
            const redoBtn = document.getElementById('term-redo');
            if(redoBtn) {
                redoBtn.addEventListener('click', () => {
                    this.replay();
                });
            }
        }

        switchTab(id) {
            this.currentTabId = id;
            
            // Update UI
            this.tabButtons.forEach(btn => {
                if (btn.dataset.id === id) btn.classList.add('active');
                else btn.classList.remove('active');
            });

            this.replay();
        }

        replay() {
            this.clearTimeouts();
            this.bodyElement.innerHTML = '';
            this.play(this.currentTabId);
        }

        clearTimeouts() {
            this.timeouts.forEach(t => clearTimeout(t));
            this.timeouts = [];
        }

        play(tabId) {
            const flow = this.data.tabs.find(t => t.id === tabId).flow;
            let lineIndex = 0;
            let charIndex = 0;

            const typeLine = () => {
                if (lineIndex >= flow.length) {
                    // Loop after delay
                    this.timeouts.push(setTimeout(() => {
                        this.replay();
                    }, 3000));
                    return;
                }

                const line = flow[lineIndex];

                if (line.type === 'pause') {
                    lineIndex++;
                    this.timeouts.push(setTimeout(typeLine, line.duration));
                    return;
                }

                if (line.type === 'input') {
                    // Create line container if starting new line
                    if (charIndex === 0) {
                        const div = document.createElement('div');
                        div.className = 'line';
                        const promptText = line.prompt || '➜ ';
                        // Determine prompt color (default purple/accent, or generic if custom prompt)
                        const promptClass = line.prompt ? 'prompt-custom' : 'prompt';
                        div.innerHTML = `<span class="${promptClass}">${promptText}</span><span class="text"></span>`;
                        this.bodyElement.appendChild(div);
                    }

                    const currentLine = this.bodyElement.lastElementChild.querySelector('.text');
                    currentLine.textContent += line.text[charIndex];
                    charIndex++;

                    if (charIndex < line.text.length) {
                        this.timeouts.push(setTimeout(typeLine, Math.random() * 30 + 30));
                    } else {
                        charIndex = 0;
                        lineIndex++;
                        this.timeouts.push(setTimeout(typeLine, 400));
                    }
                } else {
                    // Output
                    const div = document.createElement('div');
                    div.className = 'line';
                    if (line.color) div.style.color = line.color;
                    div.textContent = line.text;
                    this.bodyElement.appendChild(div);
                    lineIndex++;
                    this.timeouts.push(setTimeout(typeLine, 50));
                }
                
                // Auto Scroll
                this.bodyElement.scrollTop = this.bodyElement.scrollHeight;
            };

            typeLine();
        }
    }

    // Initialize if element exists
    if (document.getElementById('terminal-component')) {
        new TerminalComponent('terminal-component', terminalData);
    }

    // ==========================================
    // 3. UTILS: Other Page Interactivity
    // ==========================================
    
    // Copy Install Command
    const copyBtn = document.getElementById('copy-btn');
    if (copyBtn) {
        copyBtn.addEventListener('click', () => {
            const cmd = document.getElementById('install-cmd').innerText;
            navigator.clipboard.writeText(cmd).then(() => {
                // Success state
                copyBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>`;
                copyBtn.style.color = '#27c93f';
                setTimeout(() => {
                    // Revert
                    copyBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>`;
                    copyBtn.style.color = '';
                }, 2000);
            });
        });
    }

    // Fetch Stats (GitHub)
    fetch('https://api.github.com/repos/fazt-sh/fazt')
        .then(r => r.json())
        .then(data => {
            if(data.stargazers_count) {
                const el = document.getElementById('stars-count');
                if (el) el.innerText = data.stargazers_count + ' ★';
            }
        }).catch(() => {});

    // Fetch Version
    fetch('https://api.github.com/repos/fazt-sh/fazt/releases/latest')
        .then(r => r.json())
        .then(data => {
            if(data.tag_name) {
                const el = document.getElementById('version-tag');
                if(el) el.innerText = data.tag_name;
            }
        }).catch(() => {});
});