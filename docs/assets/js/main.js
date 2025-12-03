document.addEventListener('DOMContentLoaded', () => {
    // Typing Animation Data
    const lines = [
        { type: 'input', text: 'fazt server init --domain my-paas.com' },
        { type: 'output', text: '✓ Server initialized successfully', color: '#27c93f' },
        { type: 'output', text: '  Database: ~/.config/fazt/data.db' },
        { type: 'pause', duration: 800 },
        { type: 'input', text: 'fazt deploy --domain blog' },
        { type: 'output', text: 'Deploying to blog.my-paas.com...' },
        { type: 'output', text: '✓ Deployment successful! (24 files, 1.2MB)', color: '#27c93f' },
        { type: 'pause', duration: 2000 }
    ];

    const terminalBody = document.getElementById('terminal-content');
    let lineIndex = 0;
    let charIndex = 0;

    function typeLine() {
        if (!terminalBody) return;

        if (lineIndex >= lines.length) {
            setTimeout(() => {
                terminalBody.innerHTML = '';
                lineIndex = 0;
                typeLine();
            }, 3000);
            return;
        }

        const line = lines[lineIndex];
        
        if (line.type === 'pause') {
            lineIndex++;
            setTimeout(typeLine, line.duration);
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
                setTimeout(typeLine, Math.random() * 30 + 30); // Typing speed
            } else {
                charIndex = 0;
                lineIndex++;
                setTimeout(typeLine, 400); // Pause after typing
            }
        } else {
            // Output appears instantly
            const div = document.createElement('div');
            div.className = 'line';
            if (line.color) div.style.color = line.color;
            div.textContent = line.text;
            terminalBody.appendChild(div);
            lineIndex++;
            setTimeout(typeLine, 100);
        }
        
        // Auto scroll
        terminalBody.scrollTop = terminalBody.scrollHeight;
    }

    // Start typing
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
