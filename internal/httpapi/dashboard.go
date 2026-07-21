package httpapi

import "net/http"

func writeDashboard(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(dashboardHTML))
}

const dashboardHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>AI Code Tracker</title>
  <style>
    :root { color: #172124; background: #f6f7f2; font-family: Inter, "Microsoft YaHei", sans-serif; }
    * { box-sizing: border-box; }
    body { margin: 0; }
    main { max-width: 1180px; margin: 0 auto; padding: 40px 24px 64px; }
    header { display: flex; justify-content: space-between; align-items: baseline; border-bottom: 1px solid #ccd4d0; padding-bottom: 22px; }
    h1 { font-size: 26px; font-weight: 700; letter-spacing: 0; margin: 0; }
    .status { color: #52625e; font-size: 14px; }
    .stats { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 12px; margin: 28px 0 38px; }
    .stat { border: 1px solid #ccd4d0; background: #fff; padding: 18px; min-height: 112px; }
    .stat span { display: block; color: #52625e; font-size: 13px; margin-bottom: 12px; }
    .stat strong { font-size: 30px; font-weight: 650; }
    h2 { font-size: 17px; margin: 0 0 14px; }
    .table-wrap { overflow-x: auto; border-top: 2px solid #172124; }
    table { width: 100%; border-collapse: collapse; min-width: 820px; background: #fff; }
    th, td { border-bottom: 1px solid #dce2df; padding: 13px 12px; text-align: left; vertical-align: top; font-size: 13px; }
    th { color: #52625e; font-size: 12px; font-weight: 600; background: #edf1ee; }
    .message { max-width: 380px; word-break: break-word; }
    .commit { color: #52625e; font-family: ui-monospace, Consolas, monospace; }
    .ai { color: #0c6c54; font-weight: 600; }
    @media (max-width: 800px) { main { padding: 24px 16px; } header { align-items: flex-start; flex-direction: column; gap: 8px; } .stats { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
  </style>
</head>
<body>
  <main>
    <header><h1>AI Code Tracker</h1><span id="status" class="status">正在加载</span></header>
    <section class="stats" aria-label="统计概览">
      <div class="stat"><span>提交总数</span><strong id="total-commits">-</strong></div>
      <div class="stat"><span>AI 提交</span><strong id="ai-commits">-</strong></div>
      <div class="stat"><span>AI 代码行</span><strong id="ai-lines">-</strong></div>
      <div class="stat"><span>代码总行</span><strong id="total-lines">-</strong></div>
      <div class="stat"><span>仓库数量</span><strong id="repositories">-</strong></div>
    </section>
    <section><h2>最近提交</h2><div class="table-wrap"><table><thead><tr><th>时间</th><th>仓库</th><th>作者</th><th>AI 行</th><th>总行</th><th>提交</th><th>信息</th></tr></thead><tbody id="records"></tbody></table></div></section>
  </main>
  <script>
    const format = new Intl.NumberFormat("zh-CN");
    const set = (id, value) => document.getElementById(id).textContent = format.format(value);
    const cell = (row, text, className) => { const element = document.createElement("td"); element.textContent = text; if (className) element.className = className; row.append(element); };
    fetch("/v1/dashboard").then(response => { if (!response.ok) throw new Error("load failed"); return response.json(); }).then(data => {
      set("total-commits", data.total_commits); set("ai-commits", data.ai_commits); set("ai-lines", data.ai_lines); set("total-lines", data.total_lines); set("repositories", data.repositories);
      const table = document.getElementById("records");
      for (const record of data.recent_records) { const row = document.createElement("tr"); cell(row, record.date); cell(row, record.repository_url); cell(row, record.author); cell(row, record.ai_lines, record.is_ai_commit ? "ai" : ""); cell(row, record.total_lines); cell(row, record.commit_id.slice(0, 12), "commit"); cell(row, record.message, "message"); table.append(row); }
      document.getElementById("status").textContent = "已更新";
    }).catch(() => { document.getElementById("status").textContent = "数据暂不可用"; });
  </script>
</body>
</html>
`
