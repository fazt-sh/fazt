document.addEventListener('DOMContentLoaded', () => {
    
    // Scenarios Data
    const scenarios = {
        deploy: [
            { type: 'input', text: 'fazt client deploy --path ./my-app --domain awesome' },
            { type: 'output', text: "Deploying ./my-app to http://localhost:4698 as 'awesome'..." },
            { type: 'output', text: "Zipped 12 files (15420 bytes)" },
            { type: 'pause', duration: 500 },
            { type: 'output', text: "✓ Deployment successful!", color: '#27c93f' },
            { type: 'output', text: "  Site: http://awesome.fazt.sh" },
            { type: 'output', text: "  Files: 12" },
            { type: 'output', text: "  Size: 15420 bytes" },
            { type: 'pause', duration: 3000 }
        ],
        install: [
            { type: 'input', text: 'sudo fazt service install --domain my-paas.com --email admin@my-paas.com --https' },
            { type: 'output', text: "fazt.sh System Installer" },
            { type: 'output', text: "------------------------" },
            { type: 'output', text: "ℹ Installing for domain: my-paas.com" },
            { type: 'output', text: "✓ Ports [80 443] are available", color: '#27c93f' },
            { type: 'output', text: "✓ User 'fazt' exists", color: '#27c93f' },
            { type: 'output', text: "✓ Binary installed to /usr/local/bin/fazt", color: '#27c93f' },
            { type: 'pause', duration: 300 },
            { type: 'output', text: "✓ Capabilities set", color: '#27c93f' },
            { type: 'output', text: "✓ Configuring environment...", color: '#27c93f' },
            { type: 'output', text: "✓ Firewall configured (UFW)", color: '#27c93f' },
            { type: 'output', text: "✓ Systemd service installed", color: '#27c93f' },
            { type: 'output', text: "✓ Service started", color: '#27c93f' },
            { type: 'pause', duration: 500 },
            { type: 'output', text: " " },
            { type: 'output', text: "Installation Complete" },
            { type: 'output', text: "---------------------" },
            { type: 'output', text: "✓ Fazt is now running at https://my-paas.com", color: '#27c93f' },
            { type: 'output', text: " " },
            { type: 'output', text: "╔══════════════════════════════════════════════════════════╗", color: '#ffbd2e' },
            { type: 'output', text: "║ ADMIN CREDENTIALS (SAVE THESE!)                          ║", color: '#ffbd2e' },
            { type: 'output', text: "╠══════════════════════════════════════════════════════════╣", color: '#ffbd2e' },
            { type: 'output', text: "║ Username:  admin                                         ║", color: '#ffbd2e' },
            { type: 'output', text: "║ Password:  fazt-82h9d29dh92d                             ║", color: '#ffbd2e' },
            { type: 'output', text: "╚══════════════════════════════════════════════════════════╝", color: '#ffbd2e' },
            { type: 'output', text: " " },
            { type: 'output', text: "Login at: https://admin.my-paas.com" },
            { type: 'pause', duration: 5000 }
        ]
    };

    const terminalBody = document.getElementById('terminal-content');
    const tabBtns = document.querySelectorAll('.tab-btn');
    
    let currentScenario = 'deploy'; // Default
    let lineIndex = 0;
    let charIndex = 0;
    let timeouts = []; // To clear pending timeouts on switch

    function clearTimeouts() {
        timeouts.forEach(t => clearTimeout(t));
        timeouts = [];
    }

    function typeLine() {
        if (!terminalBody) return;
        const lines = scenarios[currentScenario];

        if (lineIndex >= lines.length) {
            // Loop scenario
            timeouts.push(setTimeout(() => {
                terminalBody.innerHTML = '';
                lineIndex = 0;
                typeLine();
            }, 3000));
            return;
        }

        const line = lines[lineIndex];
        
        if (line.type === 'pause') {
            lineIndex++;
            timeouts.push(setTimeout(typeLine, line.duration));
            return;
        }

        if (line.type === 'input') {
            // Add prompt if starting new line
            if (charIndex === 0) {
                const div = document.createElement('div');
                div.className = 'line';
                div.innerHTML = `<span class="prompt">➜</span><span class="text"></span>`;
                terminalBody.appendChild(div);
            }
            
            const currentLine = terminalBody.lastElementChild.querySelector('.text');
            currentLine.textContent += line.text[charIndex];
            charIndex++;

            if (charIndex < line.text.length) {
                timeouts.push(setTimeout(typeLine, Math.random() * 30 + 30)); // Typing speed
            } else {
                charIndex = 0;
                lineIndex++;
                timeouts.push(setTimeout(typeLine, 400)); // Pause after typing
            }
        } else {
            // Output appears instantly
            const div = document.createElement('div');
            div.className = 'line';
            if (line.color) div.style.color = line.color;
            div.textContent = line.text;
            terminalBody.appendChild(div);
            lineIndex++;
            timeouts.push(setTimeout(typeLine, 50));
        }
        
        // Auto scroll
        terminalBody.scrollTop = terminalBody.scrollHeight;
    }

    // Tab Switching Logic
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const flow = btn.getAttribute('data-flow');
            if (flow === currentScenario) return;

            // UI Update
            tabBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');

            // Reset Animation
            clearTimeouts();
            currentScenario = flow;
            lineIndex = 0;
            charIndex = 0;
            terminalBody.innerHTML = '';
            
            // Start new flow
            typeLine();
        });
    });

    // Start initial animation
    typeLine();

    // Copy Functionality
    const copyBtn = document.getElementById('copy-btn');
    if (copyBtn) {
        copyBtn.addEventListener('click', () => {
            const cmd = document.getElementById('install-cmd').innerText;
            navigator.clipboard.writeText(cmd).then(() => {
                copyBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>`;
                copyBtn.style.color = '#27c93f';
                setTimeout(() => {
                    copyBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>`;
                    copyBtn.style.color = '';
                }, 2000);
            });
        });
    }

    // GitHub Stars
    fetch('https://api.github.com/repos/fazt-sh/fazt')
        .then(r => r.json())
        .then(data => {
            if(data.stargazers_count) {
                document.getElementById('stars-count').innerText = data.stargazers_count;
            }
        })
        .catch(() => {});
        
    // Release Version
    fetch('https://api.github.com/repos/fazt-sh/fazt/releases/latest')
        .then(r => r.json())
        .then(data => {
            if(data.tag_name) {
                document.getElementById('version-tag').innerText = data.tag_name;
            }
        })
        .catch(() => {});
});