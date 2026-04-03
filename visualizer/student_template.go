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
  /* ── Reset & Base ─────────────────────────────── */
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
    background: #0f0f1a;
    color: #e0e0e0;
    line-height: 1.6;
    padding: 2rem;
    max-width: 860px;
    margin: 0 auto;
  }

  /* ── Header ───────────────────────────────────── */
  .report-header {
    text-align: center;
    padding: 2.5rem 1rem 2rem;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
    border-radius: 16px;
    margin-bottom: 2rem;
    border: 1px solid #2a2a4a;
    box-shadow: 0 8px 32px rgba(0,0,0,0.4);
  }
  .report-header h1 {
    font-size: 1.6rem;
    font-weight: 700;
    background: linear-gradient(90deg, #e0e0e0, #7eb8da);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    margin-bottom: 0.6rem;
  }
  .report-header .subtitle {
    font-size: 0.85rem;
    color: #888;
    margin-bottom: 1.2rem;
  }
  .score-display {
    font-size: 2.8rem;
    font-weight: 800;
    margin: 0.5rem 0;
  }
  .score-green  { color: #82b366; }
  .score-yellow { color: #d79b00; }
  .score-red    { color: #b85450; }
  .progress-wrap {
    width: 300px;
    height: 14px;
    background: #2a2a3a;
    border-radius: 7px;
    margin: 1rem auto 0;
    overflow: hidden;
  }
  .progress-fill {
    height: 100%;
    border-radius: 7px;
    transition: width 1s ease;
  }
  .fill-green  { background: linear-gradient(90deg, #82b366, #a8d98a); }
  .fill-yellow { background: linear-gradient(90deg, #d79b00, #f0c040); }
  .fill-red    { background: linear-gradient(90deg, #b85450, #e07070); }

  /* ── Legend ────────────────────────────────────── */
  .legend {
    display: flex;
    justify-content: center;
    gap: 1.5rem;
    margin: 1.5rem 0;
    font-size: 0.78rem;
    color: #888;
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
  .dot-correct { background: #82b366; }
  .dot-wrong   { background: #d79b00; }
  .dot-extra   { background: #7eb8da; }

  /* ── Section Container ────────────────────────── */
  .section {
    background: #1a1a2e;
    border: 1px solid #2a2a4a;
    border-radius: 12px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
    box-shadow: 0 4px 16px rgba(0,0,0,0.3);
  }
  .section-title {
    font-size: 1.1rem;
    font-weight: 700;
    margin-bottom: 1rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid #2a2a4a;
    color: #7eb8da;
  }

  /* ── Node Card ────────────────────────────────── */
  .node-card {
    background: #12121f;
    border: 1px solid #2a2a4a;
    border-radius: 10px;
    margin-bottom: 0.8rem;
    overflow: hidden;
    transition: transform 0.2s, box-shadow 0.2s;
  }
  .node-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(0,0,0,0.4);
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
  .badge-class     { background: #16213e; color: #7eb8da; }
  .badge-interface { background: #2e1a2e; color: #da7eb8; }
  .badge-abstract  { background: #2e2e1a; color: #b8da7e; }
  .badge-enum      { background: #1a2e2e; color: #7edab8; }

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
  .status-correct { border-left-color: #82b366; color: #a8d98a; }
  .status-wrong   { border-left-color: #d79b00; color: #f0c040; }
  .status-missing { border-left-color: #b85450; color: #e07070; }
  .status-extra   { border-left-color: #7eb8da; color: #9ed0ee; font-style: italic; }
  .status-neutral { border-left-color: #444; color: #999; }

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
  .rel-correct { border-left: 3px solid #82b366; }
  .rel-wrong   { border-left: 3px solid #d79b00; }
  .rel-extra   { border-left: 3px solid #7eb8da; font-style: italic; }
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
    background: rgba(126,184,218,0.08);
    border: 1px solid #2a3a4a;
    border-radius: 10px;
    padding: 1.2rem 1.5rem;
    margin-bottom: 1.5rem;
    font-size: 0.82rem;
    color: #9ab8d0;
    line-height: 1.7;
  }
  .hint-box strong { color: #7eb8da; }

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
  <h1>📋 UML Diagram Feedback</h1>
  <div class="subtitle">Your submission has been automatically checked</div>
  <div class="score-display {{.ScoreClass}}">
    {{printf "%.1f" .Percent}}%
  </div>
  <div style="font-size:0.9rem; color:#888;">{{printf "%.2f" .Score}} / {{printf "%.2f" .MaxScore}} points</div>
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
  Items you are <strong>missing</strong> are not shown here — review your diagram carefully.
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
    <span class="rel-icon">{{.Icon}}</span>
    <span>{{.Source}}</span>
    <span style="color:#555">──▷</span>
    <span>{{.Target}}</span>
    <span class="rel-type-tag">{{.RelType}}</span>
  </div>
  {{else}}
  <div style="color:#666; text-align:center; padding:1rem;">No relations found in your submission</div>
  {{end}}
</div>

<div class="report-footer">
  Generated by UML Comparator • {{.Timestamp}}
</div>

</body>
</html>`
