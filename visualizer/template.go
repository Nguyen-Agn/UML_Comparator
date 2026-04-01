package visualizer

// htmlTemplate is the self-contained HTML+CSS template for the grading report.
// Color scheme sourced from skills/Visualizer-AI-Context/knowledge.md:
//   - Correct:  #d5e8d4 bg / #82b366 border
//   - Error:    #f8cecc bg / #b85450 border
//   - Warning:  #ffe6cc bg / #d79b00 border
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>UML Grading Report — {{.Percent}}%</title>
<style>
  /* ── Reset & Base ─────────────────────────────── */
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
    background: #0f0f1a;
    color: #e0e0e0;
    line-height: 1.6;
    padding: 2rem;
  }
  a { color: #7eb8da; }

  /* ── Header ───────────────────────────────────── */
  .report-header {
    text-align: center;
    padding: 2rem 1rem 1.5rem;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
    border-radius: 16px;
    margin-bottom: 2rem;
    border: 1px solid #2a2a4a;
    box-shadow: 0 8px 32px rgba(0,0,0,0.4);
  }
  .report-header h1 {
    font-size: 1.8rem;
    font-weight: 700;
    background: linear-gradient(90deg, #e0e0e0, #7eb8da);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    margin-bottom: 0.8rem;
  }
  .score-display {
    font-size: 2.4rem;
    font-weight: 800;
    margin: 0.5rem 0;
  }
  .score-green  { color: #82b366; }
  .score-yellow { color: #d79b00; }
  .score-red    { color: #b85450; }

  /* ── Progress Bar ─────────────────────────────── */
  .progress-wrap {
    width: 320px;
    height: 12px;
    background: #2a2a3a;
    border-radius: 6px;
    margin: 0.8rem auto 0;
    overflow: hidden;
  }
  .progress-fill {
    height: 100%;
    border-radius: 6px;
    transition: width 1s ease;
  }
  .fill-green  { background: linear-gradient(90deg, #82b366, #a8d98a); }
  .fill-yellow { background: linear-gradient(90deg, #d79b00, #f0c040); }
  .fill-red    { background: linear-gradient(90deg, #b85450, #e07070); }

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
    font-size: 1.2rem;
    font-weight: 700;
    margin-bottom: 1rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid #2a2a4a;
    color: #7eb8da;
  }

  /* ── Two-Column Grid ──────────────────────────── */
  .columns {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
  }
  .col-header {
    font-size: 1rem;
    font-weight: 700;
    padding: 0.6rem 1rem;
    border-radius: 8px;
    margin-bottom: 0.8rem;
    text-align: center;
  }
  .col-student  .col-header { background: #16213e; color: #7eb8da; }
  .col-solution .col-header { background: #1a2e1a; color: #82b366; }

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
  .status-extra   { border-left-color: #b85450; color: #e07070; font-style: italic; }
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
  .rel-missing { border-left: 3px solid #b85450; }
  .rel-wrong   { border-left: 3px solid #d79b00; }
  .rel-extra   { border-left: 3px solid #b85450; font-style: italic; }
  .rel-icon { font-size: 1rem; }
  .rel-type-tag {
    display: inline-block;
    padding: 0.1rem 0.4rem;
    border-radius: 3px;
    font-size: 0.7rem;
    background: #2a2a4a;
    color: #aaa;
  }

  /* ── Summary Cards ────────────────────────────── */
  .summary-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
    margin-bottom: 1.2rem;
  }
  .stat-card {
    text-align: center;
    padding: 1rem;
    border-radius: 10px;
    border: 1px solid #2a2a4a;
    transition: transform 0.2s;
  }
  .stat-card:hover { transform: scale(1.03); }
  .stat-card .stat-value {
    font-size: 1.8rem;
    font-weight: 800;
  }
  .stat-card .stat-label {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: #888;
    margin-top: 0.3rem;
  }
  .card-correct  { background: rgba(130,179,102,0.1); }
  .card-correct  .stat-value { color: #82b366; }
  .card-missing  { background: rgba(184,84,80,0.1); }
  .card-missing  .stat-value { color: #b85450; }
  .card-wrong    { background: rgba(215,155,0,0.1); }
  .card-wrong    .stat-value { color: #d79b00; }
  .card-extra    { background: rgba(126,184,218,0.1); }
  .card-extra    .stat-value { color: #7eb8da; }

  /* ── Feedbacks ────────────────────────────────── */
  .feedback-list {
    list-style: none;
    padding: 0;
  }
  .feedback-item {
    padding: 0.5rem 1rem;
    margin-bottom: 0.3rem;
    border-radius: 6px;
    background: rgba(184,84,80,0.08);
    border-left: 3px solid #b85450;
    font-size: 0.82rem;
    font-family: 'Cascadia Code', 'Consolas', monospace;
  }

  /* ── Footer ───────────────────────────────────── */
  .report-footer {
    text-align: center;
    padding: 1rem;
    font-size: 0.75rem;
    color: #555;
    margin-top: 1rem;
  }

  /* ── Responsive ───────────────────────────────── */
  @media (max-width: 768px) {
    .columns { grid-template-columns: 1fr; }
    .summary-grid { grid-template-columns: repeat(2, 1fr); }
    body { padding: 1rem; }
  }
</style>
</head>
<body>

<!-- ═══ HEADER ═══ -->
<div class="report-header">
  <h1>📊 UML Grading Report</h1>
  <div class="score-display {{.ScoreClass}}">
    {{printf "%.2f" .Score}} / {{printf "%.2f" .MaxScore}}
  </div>
  <div style="font-size:1rem; color:#888;">{{printf "%.1f" .Percent}}% correct</div>
  <div class="progress-wrap">
    <div class="progress-fill {{.FillClass}}" style="width:{{printf "%.1f" .Percent}}%"></div>
  </div>
</div>

<!-- ═══ NODES SIDE-BY-SIDE ═══ -->
<div class="section">
  <div class="section-title">🏗️ Nodes Comparison</div>
  <div class="columns">
    <!-- Student Column -->
    <div class="col-student">
      <div class="col-header">📘 Student (bài nộp)</div>
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
      <div style="color:#666; text-align:center; padding:1rem;">No nodes found</div>
      {{end}}
    </div>

    <!-- Solution Column -->
    <div class="col-solution">
      <div class="col-header">📗 Solution (đáp án)</div>
      {{range .SolNodes}}
      <div class="node-card">
        <div class="node-name">
          <span class="node-type-badge badge-{{.BadgeClass}}">{{.Type}}</span>
          {{.Name}}
        </div>
        <div class="member-list">
          {{if .Attributes}}
          <div class="member-group-title">Attributes</div>
          {{range .Attributes}}
          <div class="member-item status-neutral">{{.Display}}</div>
          {{end}}
          {{end}}
          {{if .Methods}}
          <div class="member-group-title">Methods</div>
          {{range .Methods}}
          <div class="member-item status-neutral">{{.Display}}</div>
          {{end}}
          {{end}}
        </div>
      </div>
      {{else}}
      <div style="color:#666; text-align:center; padding:1rem;">No nodes found</div>
      {{end}}
    </div>
  </div>
</div>

<!-- ═══ RELATIONS ═══ -->
<div class="section">
  <div class="section-title">🔗 Relations</div>
  {{range .Relations}}
  <div class="relation-row rel-{{.Status}}">
    <span class="rel-icon">{{.Icon}}</span>
    <span>{{.Source}}</span>
    <span style="color:#555">──▷</span>
    <span>{{.Target}}</span>
    <span class="rel-type-tag">{{.RelType}}</span>
    {{if .Note}}<span style="color:#888; font-size:0.75rem; margin-left:auto;">{{.Note}}</span>{{end}}
  </div>
  {{else}}
  <div style="color:#666; text-align:center; padding:1rem;">No relations found</div>
  {{end}}
</div>

<!-- ═══ SUMMARY ═══ -->
<div class="section">
  <div class="section-title">📝 Summary</div>
  <div class="summary-grid">
    <div class="stat-card card-correct">
      <div class="stat-value">{{.Stats.Correct}}</div>
      <div class="stat-label">Correct</div>
    </div>
    <div class="stat-card card-missing">
      <div class="stat-value">{{.Stats.Missing}}</div>
      <div class="stat-label">Missing</div>
    </div>
    <div class="stat-card card-wrong">
      <div class="stat-value">{{.Stats.Wrong}}</div>
      <div class="stat-label">Wrong</div>
    </div>
    <div class="stat-card card-extra">
      <div class="stat-value">{{.Stats.Extra}}</div>
      <div class="stat-label">Extra</div>
    </div>
  </div>

  {{if .Feedbacks}}
  <div style="margin-top:1rem;">
    <div class="member-group-title" style="padding-left:0;">Deduction Details</div>
    <ul class="feedback-list">
      {{range .Feedbacks}}
      <li class="feedback-item">{{.}}</li>
      {{end}}
    </ul>
  </div>
  {{end}}
</div>

<div class="report-footer">
  Generated by UML Comparator • {{.Timestamp}}
</div>

</body>
</html>`
