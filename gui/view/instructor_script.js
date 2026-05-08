let notifTimeout;
function showNotification(msg, type) {
    const notif = document.getElementById('notification');
    const msgEl = document.getElementById('notif-msg');

    clearTimeout(notifTimeout);
    notif.className = '';
    notif.classList.add(type);
    notif.classList.add('show');
    msgEl.innerText = msg;

    notifTimeout = setTimeout(() => {
        notif.classList.remove('show');
    }, 4000);
}

function switchTab(e, tabId) {
    document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));

    document.getElementById(tabId).classList.add('active');
    if (e && e.currentTarget) {
        e.currentTarget.classList.add('active');
    }
}

function setValue(id, val) { document.getElementById(id).value = val; }
function getValue(id) { return document.getElementById(id).value; }

function renderLiveResult(b64) {
    const bin = atob(b64);
    const blob = new Blob([bin], { type: 'text/html' });
    const url = URL.createObjectURL(blob);
    const iframe = document.getElementById("live-iframe");

    iframe.style.display = "block";
    iframe.onload = () => { URL.revokeObjectURL(url); };
    iframe.src = url;
    document.getElementById("loading").style.display = "none";
}

// Event forwarders
function goRunLive() { goExecLive(getValue('live-sol'), getValue('live-stu')); }
function goRunBatch() { goExecBatch(getValue('batch-sol'), getValue('batch-dir'), getValue('batch-out')); }
function goRunEncrypt() { goExecEncrypt(getValue('sec-in'), getValue('sec-out')); }
function goRunExamBuild() { goExecExam(getValue('exam-dir'), getValue('exam-out')); }

// Config Tab Logic
function onThresholdChange(val) {
    document.getElementById('threshold-val').innerText = parseFloat(val).toFixed(2);
}

function onAIChange(checked) {
    // Logic handled on apply
}

function applyConfig() {
    const th = parseFloat(document.getElementById('config-threshold').value);
    const ai = document.getElementById('config-ai').checked;
    goUpdateConfig(th, ai);
}

function updateConfigUI(th, ai, avail) {
    document.getElementById('config-threshold').value = th;
    document.getElementById('threshold-val').innerText = th.toFixed(2);
    document.getElementById('config-ai').checked = ai;

    const aiStatus = document.getElementById('ai-status');
    const aiCheckbox = document.getElementById('config-ai');

    if (avail) {
        aiStatus.innerText = "AI model loaded successfully (MiniLM).";
        aiStatus.style.color = "var(--success)";
        aiCheckbox.disabled = false;
    } else {
        aiStatus.innerText = "AI model not found. Falling back to fuzzy matching.";
        aiStatus.style.color = "var(--error)";
        aiCheckbox.checked = false;
        aiCheckbox.disabled = true;
    }
}

// --- GENERATOR LOGIC ---

// ── Mermaid init ──────────────────────────────────
mermaid.initialize({
    startOnLoad: false,
    theme: 'base',
    themeVariables: {
        primaryColor: '#ffffff',
        primaryTextColor: '#16202b',
        primaryBorderColor: '#114665',
        lineColor: '#114665',
        secondaryColor: '#f4f7f6',
        tertiaryColor: '#ffffff',
        fontFamily: 'Inter, system-ui, sans-serif'
    },
    themeCSS: `
        g.classGroup rect {
          stroke-width: 1.5;
          rx: 6px !important;
          ry: 6px !important;
        }
        g.classGroup text .classTitle {
          font-weight: 700;
          fill: #720f32 !important;
        }
        g.classGroup line {
          stroke: #d1d1d6;
          stroke-width: 1;
        }
        .edgeTerminals, .edgePath .path {
          stroke-width: 1.5;
        }
      `,
    securityLevel: 'loose'
});

let lastRenderedCode = '';
let renderTimeout;

async function renderGenMermaid(code, targetId) {
    const container = document.getElementById(targetId);
    if (!code || !code.trim()) {
        container.innerHTML = '<span style="color:var(--text-dim);font-size:0.8rem;">Kết quả UML sẽ hiển thị ở đây...</span>';
        return;
    }

    container.removeAttribute('data-processed');
    container.innerHTML = code;
    lastRenderedCode = code;

    try {
        await mermaid.run({ nodes: [container] });
    } catch (e) {
        container.innerHTML = '<p style="color:var(--error);padding:20px;font-size:0.85rem;">Lỗi cú pháp: ' + e.message + '</p>';
    }
}

// Called by Go backend after AI generation succeeds (from generator_view.go bindings)
function showGeneratedUML(mermaidCode) {
    document.getElementById('gen-mermaid-editor').value = mermaidCode;
    onGenEditorChange();
    showNotification('UML được tạo thành công!', 'success');
}

function onGenEditorChange() {
    const currentCode = document.getElementById('gen-mermaid-editor').value;
    clearTimeout(renderTimeout);
    renderTimeout = setTimeout(() => {
        renderGenMermaid(currentCode, 'gen-live-mermaid-render');
    }, 500);
}

function insertGenTemplate() {
    const tmpl = `classDiagram\n    ClassA <|-- ClassB : __1__\n    \n    class ClassA {\n      + "name" : "String" __1__\n    }\n    \n    class ClassB {\n      + "action()" "void" __1__\n    }`;
    document.getElementById('gen-mermaid-editor').value = tmpl;
    onGenEditorChange();
}

// Modal and Config
async function openGenConfigModal() {
    const jsonStr = await goLoadGenConfig();
    const cfg = JSON.parse(jsonStr);
    document.getElementById('gen-cfg-endpoint').value = cfg.api_endpoint || '';
    document.getElementById('gen-cfg-model').value = cfg.model || '';
    document.getElementById('gen-cfg-key').value = cfg.api_key || '';
    document.getElementById('genConfigModal').style.display = 'flex';
}

function closeGenConfigModal() {
    document.getElementById('genConfigModal').style.display = 'none';
}

async function saveGenConfig() {
    const endpoint = document.getElementById('gen-cfg-endpoint').value.trim();
    const model = document.getElementById('gen-cfg-model').value.trim();
    const key = document.getElementById('gen-cfg-key').value.trim();

    const resJSON = await goSaveGenConfig(endpoint, model, key);
    const res = JSON.parse(resJSON);
    if (res.error) {
        showNotification(res.error, 'error');
    } else {
        showNotification('Cấu hình đã lưu!', 'success');
        closeGenConfigModal();
    }
}

// Actions
function goGenGenerate() {
    const problem = document.getElementById('gen-problem-input').value.trim();
    if (!problem) { showNotification('Vui lòng nhập đề bài.', 'error'); return; }
    goExecGenerate(problem);
}

async function saveGenFile() {
    const code = document.getElementById('gen-mermaid-editor').value.trim();
    if (!code) { showNotification('Không có UML để lưu.', 'error'); return; }
    const resJSON = await goExecSave(code);
    const res = JSON.parse(resJSON);
    if (res.error) {
        showNotification(res.error, 'error');
    } else {
        showNotification('Lưu file thành công!', 'success');
    }
}
