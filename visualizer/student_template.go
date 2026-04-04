package visualizer

// studentHTMLTemplate is a student-facing HTML report.
// It shows ONLY the student's nodes/relations with color-coded status.
// It does NOT reveal: solution nodes, deduction feedbacks, or detailed summary.
const studentHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>UML Feedback Report — {{printf "%.1f" .Percent}}%</title>
<style>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800&display=swap" rel="stylesheet">
<style>
  /* ── Reset & Base ─────────────────────────────── */
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: 'Inter', system-ui, -apple-system, sans-serif;
    background: #f8fafc;
    color: #16202b;
    line-height: 1.6;
    padding: 3rem 1rem;
    max-width: 960px;
    margin: 0 auto;
  }

  /* ── Header ───────────────────────────────────── */
  .report-header {
    text-align: center;
    padding: 3rem 2rem;
    background: #ffffff;
    border-radius: 20px;
    margin-bottom: 2.5rem;
    border: 1px solid rgba(22, 32, 43, 0.08);
    box-shadow: 0 10px 30px rgba(0,0,0,0.04);
    position: relative;
    overflow: hidden;
  }
  .report-header::before {
    content: '';
    position: absolute;
    top: 0; left: 0; right: 0;
    height: 6px;
    background: linear-gradient(90deg, #720f32, #114665);
  }
  .report-header h1 {
    font-size: 1.8rem;
    font-weight: 800;
    color: #16202b;
    margin-bottom: 0.6rem;
    letter-spacing: -0.02em;
  }
  .report-header .subtitle {
    font-size: 0.9rem;
    color: #7b445a;
    margin-bottom: 1.5rem;
  }
  .score-display {
    font-size: 3.5rem;
    font-weight: 800;
    margin: 0.5rem 0;
    letter-spacing: -0.04em;
  }
  .score-green  { color: #10b981; }
  .score-yellow { color: #d79b00; }
  .score-red    { color: #720f32; }
  .progress-wrap {
    width: 340px;
    height: 12px;
    background: #edf2f7;
    border-radius: 10px;
    margin: 1.5rem auto 0;
    overflow: hidden;
  }
  .progress-fill {
    height: 100%;
    border-radius: 10px;
    transition: width 1s cubic-bezier(0.4, 0, 0.2, 1);
  }
  .fill-green  { background: #10b981; }
  .fill-yellow { background: #d79b00; }
  .fill-red    { background: #720f32; }

  /* ── Legend ────────────────────────────────────── */
  .legend {
    display: flex;
    justify-content: center;
    gap: 2rem;
    margin: 2rem 0;
    font-size: 0.85rem;
    font-weight: 600;
    color: #7b445a;
  }
  .legend-item {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }
  .legend-dot {
    width: 10px;
    height: 10px;
    border-radius: 3px;
  }
  .dot-correct { background: #10b981; }
  .dot-wrong   { background: #720f32; }
  .dot-extra   { background: #114665; }

  /* ── Section Container ────────────────────────── */
  .section {
    background: #ffffff;
    border: 1px solid rgba(22, 32, 43, 0.08);
    border-radius: 16px;
    padding: 2rem;
    margin-bottom: 2rem;
    box-shadow: 0 4px 20px rgba(0,0,0,0.03);
  }
  .section-title {
    font-size: 1.2rem;
    font-weight: 800;
    margin-bottom: 1.5rem;
    padding-bottom: 0.8rem;
    border-bottom: 2px solid #f1f5f9;
    color: #114665;
    letter-spacing: -0.01em;
  }

  /* ── Node Card ────────────────────────────────── */
  .node-card {
    background: #f8fafc;
    border: 1px solid rgba(22, 32, 43, 0.06);
    border-radius: 12px;
    margin-bottom: 1rem;
    overflow: hidden;
    transition: all 0.2s ease;
  }
  .node-card:hover {
    transform: translateY(-3px);
    box-shadow: 0 10px 25px rgba(0,0,0,0.06);
    border-color: #114665;
  }
  .node-name {
    padding: 0.6rem 1rem;
    font-weight: 700;
    font-size: 0.95rem;
    border-bottom: 1px solid #2a2a4a;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .node-type-badge {
    display: inline-block;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .badge-class     { background: #114665; color: white; }
  .badge-interface { background: #720f32; color: white; }
  .badge-abstract  { background: #7b445a; color: white; }
  .badge-enum      { background: #16202b; color: white; }

  .member-list {
    padding: 0.4rem 0;
    font-size: 0.82rem;
  }
  .member-group-title {
    padding: 0.2rem 1rem;
    font-size: 0.72rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: #666;
  }
  .member-item {
    padding: 0.25rem 1rem 0.25rem 1.4rem;
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: 0.8rem;
    border-left: 3px solid transparent;
    transition: background 0.15s;
  }
  .member-item:hover {
    background: rgba(255,255,255,0.03);
  }
  .status-correct { border-left-color: #10b981; color: #065f46; background: rgba(16, 185, 129, 0.04); }
  .status-wrong   { border-left-color: #720f32; color: #720f32; background: rgba(114, 15, 50, 0.04); }
  .status-missing { border-left-color: #720f32; color: #720f32; background: rgba(114, 15, 50, 0.04); }
  .status-extra   { border-left-color: #114665; color: #114665; font-style: italic; background: rgba(17, 70, 101, 0.04); }
  .status-neutral { border-left-color: #cbd5e1; color: #64748b; }

  /* ── Relation Row ─────────────────────────────── */
  .relation-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    margin-bottom: 0.4rem;
    font-size: 0.85rem;
    font-family: 'Cascadia Code', 'Consolas', monospace;
    transition: background 0.15s;
  }
  .relation-row:hover { background: rgba(255,255,255,0.03); }
  .rel-correct { border-left: 4px solid #10b981; }
  .rel-wrong   { border-left: 4px solid #720f32; }
  .rel-extra   { border-left: 4px solid #114665; font-style: italic; }
  .rel-icon { font-size: 1rem; }
  .rel-type-tag {
    display: inline-block;
    padding: 0.1rem 0.4rem;
    border-radius: 3px;
    font-size: 0.7rem;
    background: #2a2a4a;
    color: #aaa;
  }

  /* ── Hint Box ─────────────────────────────────── */
  .hint-box {
    background: #f1f5f9;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.5rem;
    margin-bottom: 2rem;
    font-size: 0.85rem;
    color: #475569;
    line-height: 1.7;
  }
  .hint-box strong { color: #114665; }

  /* ── Footer ───────────────────────────────────── */
  .report-footer {
    text-align: center;
    padding: 1rem;
    font-size: 0.75rem;
    color: #555;
    margin-top: 1rem;
  }

  @media (max-width: 600px) {
    body { padding: 1rem; }
    .legend { flex-wrap: wrap; gap: 0.8rem; }
  }
</style>
</head>
<body>

<!-- ═══ HEADER ═══ -->
<div class="report-header">
  <h1> UML Diagram Feedback</h1>
  <div class="subtitle">Your submission has been automatically checked</div>
  <div class="score-display {{.ScoreClass}}">
    {{printf "%.1f" .Percent}}%
  </div>
  <div class="progress-wrap">
    <div class="progress-fill {{.FillClass}}" style="width:{{printf "%.1f" .Percent}}%"></div>
  </div>
</div>

<!-- ═══ LEGEND ═══ -->
<div class="legend">
  <div class="legend-item"><div class="legend-dot dot-correct"></div> Correct</div>
  <div class="legend-item"><div class="legend-dot dot-wrong"></div> Needs Fix</div>
  <div class="legend-item"><div class="legend-dot dot-extra"></div> Not Required</div>
</div>

<!-- ═══ HINT ═══ -->
<div class="hint-box">
  <strong>How to read this report:</strong> Each of your classes, attributes, methods, and relations is shown below. 
  Items marked in <span style="color:#a8d98a">green</span> are correct. 
  Items in <span style="color:#f0c040">yellow</span> have issues (wrong type, scope, or name). 
  Items in <span style="color:#9ed0ee">blue italic</span> were not expected in the solution.
  Items you are <strong>missing</strong> are not shown here, review your diagram carefully.
</div>

<!-- ═══ YOUR CLASSES ═══ -->
<div class="section">
  <div class="section-title">Your Classes & Members</div>
  {{range .StuNodes}}
  <div class="node-card">
    <div class="node-name">
      <span class="node-type-badge badge-{{.BadgeClass}}">{{.Type}}</span>
      {{.Name}}
    </div>
    <div class="member-list">
      {{if .Attributes}}
      <div class="member-group-title">Attributes</div>
      {{range .Attributes}}
      <div class="member-item status-{{.Status}}">{{.Display}}</div>
      {{end}}
      {{end}}
      {{if .Methods}}
      <div class="member-group-title">Methods</div>
      {{range .Methods}}
      <div class="member-item status-{{.Status}}">{{.Display}}</div>
      {{end}}
      {{end}}
    </div>
  </div>
  {{else}}
  <div style="color:#666; text-align:center; padding:1rem;">No classes found in your submission</div>
  {{end}}
</div>

<!-- ═══ YOUR RELATIONS ═══ -->
<div class="section">
  <div class="section-title">Your Relations</div>
  {{range .Relations}}
  <div class="relation-row rel-{{.Status}}">
    <span class="rel-icon">
      {{if eq .Status "correct"}}<svg viewBox="0 0 24 24" width="18" height="18" stroke="currentColor" stroke-width="3" fill="none"><polyline points="20 6 9 17 4 12"></polyline></svg>
      {{else if eq .Status "wrong"}}<svg viewBox="0 0 24 24" width="18" height="18" stroke="currentColor" stroke-width="3" fill="none"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
      {{else if eq .Status "extra"}}<svg viewBox="0 0 24 24" width="18" height="18" stroke="currentColor" stroke-width="3" fill="none"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="12" x2="16" y2="12"></line><line x1="12" y1="8" x2="12" y2="16"></line></svg>
      {{end}}
    </span>
    <span>{{.Source}}</span>
    <span style="color:#cbd5e1">──▷</span>
    <span>{{.Target}}</span>
    <span class="rel-type-tag">{{.RelType}}</span>
  </div>
  {{else}}
  <div style="color:#666; text-align:center; padding:1rem;">No relations found in your submission</div>
  {{end}}
</div>

<div class="report-footer">
  Generated by UML Comparator {{.Timestamp}}
</div>

</body>
</html>`
