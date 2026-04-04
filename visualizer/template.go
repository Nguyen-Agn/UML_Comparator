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
    max-width: 1200px;
    margin: 0 auto;
  }
  a { color: #114665; }

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
    font-size: 2rem;
    font-weight: 800;
    color: #16202b;
    letter-spacing: -0.02em;
    margin-bottom: 0.8rem;
  }
  .score-display {
    font-size: 3rem;
    font-weight: 800;
    margin: 0.5rem 0;
    letter-spacing: -0.04em;
  }
  .score-green  { color: #10b981; }
  .score-yellow { color: #d79b00; }
  .score-red    { color: #720f32; }

  /* ── Progress Bar ─────────────────────────────── */
  .progress-wrap {
    width: 380px;
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

  /* ── Section Container ────────────────────────── */
  .section {
    background: #ffffff;
    border: 1px solid rgba(22, 32, 43, 0.08);
    border-radius: 20px;
    padding: 2.5rem;
    margin-bottom: 2.5rem;
    box-shadow: 0 4px 20px rgba(0,0,0,0.03);
  }
  .section-title {
    font-size: 1.4rem;
    font-weight: 800;
    margin-bottom: 1.5rem;
    padding-bottom: 0.8rem;
    border-bottom: 2px solid #f1f5f9;
    color: #114665;
    letter-spacing: -0.01em;
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
  .col-student  .col-header { background: #f1f5f9; color: #114665; }
  .col-solution .col-header { background: rgba(114, 15, 50, 0.05); color: #720f32; }

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
  .rel-missing { border-left: 4px solid #720f32; opacity: 0.6; }
  .rel-wrong   { border-left: 4px solid #d79b00; }
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
  .card-correct  { background: rgba(16,185,129,0.05); border-color: rgba(16,185,129,0.1); }
  .card-correct  .stat-value { color: #10b981; }
  .card-missing  { background: rgba(114,15,50,0.05); border-color: rgba(114,15,50,0.1); }
  .card-missing  .stat-value { color: #720f32; }
  .card-wrong    { background: rgba(215,155,0,0.05); border-color: rgba(215,155,0,0.1); }
  .card-wrong    .stat-value { color: #d79b00; }
  .card-extra    { background: rgba(17,70,101,0.05); border-color: rgba(17,70,101,0.1); }
  .card-extra    .stat-value { color: #114665; }

  /* ── Feedbacks ────────────────────────────────── */
  .feedback-list {
    list-style: none;
    padding: 0;
  }
  .feedback-item {
    padding: 0.8rem 1.2rem;
    margin-bottom: 0.5rem;
    border-radius: 10px;
    background: rgba(114, 15, 50, 0.03);
    border-left: 4px solid #720f32;
    font-size: 0.85rem;
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
    <span class="rel-icon">
      {{if eq .Status "correct"}}<svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="3" fill="none"><polyline points="20 6 9 17 4 12"></polyline></svg>
      {{else if eq .Status "wrong"}}<svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="3" fill="none"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
      {{else if eq .Status "missing"}}<svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="3" fill="none"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
      {{else if eq .Status "extra"}}<svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="3" fill="none"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="12" x2="16" y2="12"></line><line x1="12" y1="8" x2="12" y2="16"></line></svg>
      {{end}}
    </span>
    <span>{{.Source}}</span>
    <span style="color:#cbd5e1">──▷</span>
    <span>{{.Target}}</span>
    <span class="rel-type-tag">{{.RelType}}</span>
    {{if .Note}}<span class="rel-note">{{.Note}}</span>{{end}}
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
