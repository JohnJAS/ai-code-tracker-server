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
    .stats { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 12px; margin: 28px 0 38px; }
    .stat { border: 1px solid #ccd4d0; background: #fff; padding: 18px; min-height: 112px; }
    .stat span { display: block; color: #52625e; font-size: 13px; margin-bottom: 12px; }
    .stat strong { font-size: 30px; font-weight: 650; }
    .records-heading { display: flex; justify-content: space-between; align-items: end; gap: 16px; margin-bottom: 14px; }
    h2 { font-size: 17px; margin: 0; }
    .filters { display: grid; grid-template-columns: 1.2fr 1.2fr 1fr 1fr auto; gap: 8px; align-items: end; }
    .filter { display: grid; gap: 5px; }
    .filter label { color: #52625e; font-size: 12px; }
    input, button { font: inherit; }
    input { border: 1px solid #b9c4bf; background: #fff; color: #172124; padding: 9px 10px; min-width: 0; }
    button { border: 1px solid #172124; background: #172124; color: #fff; padding: 9px 14px; cursor: pointer; }
    button:disabled { border-color: #c7cfcb; background: #e7ebe8; color: #7a8581; cursor: not-allowed; }
    .table-wrap { overflow-x: auto; border-top: 2px solid #172124; }
    table { width: 100%; border-collapse: collapse; min-width: 820px; background: #fff; }
    th, td { border-bottom: 1px solid #dce2df; padding: 13px 12px; text-align: left; vertical-align: top; font-size: 13px; }
    th { color: #52625e; font-size: 12px; font-weight: 600; background: #edf1ee; }
    .message { max-width: 380px; word-break: break-word; }
    .commit { color: #52625e; font-family: ui-monospace, Consolas, monospace; }
    .ai { color: #0c6c54; font-weight: 600; }
    .empty { color: #52625e; text-align: center; padding: 28px 12px; }
    .pagination { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-top: 14px; }
    .pagination-actions { display: flex; gap: 8px; }
    @media (max-width: 980px) { .filters { grid-template-columns: repeat(2, minmax(0, 1fr)); } .filters button { width: max-content; } }
    @media (max-width: 800px) { main { padding: 24px 16px; } header { align-items: flex-start; flex-direction: column; gap: 8px; } .stats { grid-template-columns: repeat(2, minmax(0, 1fr)); } .records-heading { align-items: flex-start; flex-direction: column; } .filters { width: 100%; } }
    @media (max-width: 520px) { .filters { grid-template-columns: 1fr; } .pagination { align-items: flex-start; flex-direction: column; } }
  </style>
</head>
<body>
  <main>
    <header><h1>AI Code Tracker</h1><span id="status" class="status">&#27491;&#22312;&#21152;&#36733;</span></header>
    <section class="stats" aria-label="&#32479;&#35745;&#27010;&#35272;">
      <div class="stat"><span>&#25552;&#20132;&#24635;&#25968;</span><strong id="total-commits">-</strong></div>
      <div class="stat"><span>AI &#21442;&#19982;&#25552;&#20132;</span><strong id="ai-commits">-</strong></div>
      <div class="stat"><span>AI &#20195;&#30721;&#34892;</span><strong id="ai-lines">-</strong></div>
      <div class="stat"><span>&#20195;&#30721;&#24635;&#34892;</span><strong id="total-lines">-</strong></div>
      <div class="stat"><span>&#20179;&#24211;&#25968;&#37327;</span><strong id="repositories">-</strong></div>
      <div class="stat"><span>&#36129;&#29486;&#20154;&#25968;</span><strong id="contributors">-</strong></div>
    </section>
    <section>
      <div class="records-heading">
        <h2>&#25552;&#20132;&#35760;&#24405;</h2>
        <form id="filters" class="filters">
          <div class="filter"><label for="author">&#20316;&#32773;&#20851;&#38190;&#35789;</label><input id="author" name="author" type="search"></div>
          <div class="filter"><label for="repository">&#20179;&#24211;&#20851;&#38190;&#35789;</label><input id="repository" name="repository" type="search"></div>
          <div class="filter"><label for="start-date">&#24320;&#22987;&#26085;&#26399;</label><input id="start-date" name="start_date" type="date"></div>
          <div class="filter"><label for="end-date">&#32467;&#26463;&#26085;&#26399;</label><input id="end-date" name="end_date" type="date"></div>
          <button type="submit">&#26597;&#35810;</button>
        </form>
      </div>
      <div class="table-wrap">
        <table><thead><tr><th>&#26102;&#38388;</th><th>&#20179;&#24211;</th><th>&#20316;&#32773;</th><th>AI &#24037;&#20855;</th><th>AI &#34892;</th><th>&#24635;&#34892;</th><th>&#25552;&#20132;</th><th>&#20449;&#24687;</th></tr></thead><tbody id="records"></tbody></table>
      </div>
      <div class="pagination">
        <span id="result-summary" class="status"></span>
        <div class="pagination-actions">
          <button id="previous-page" type="button" disabled>&#19978;&#19968;&#39029;</button>
          <button id="next-page" type="button" disabled>&#19979;&#19968;&#39029;</button>
        </div>
      </div>
    </section>
  </main>
  <script>
    const format = new Intl.NumberFormat("zh-CN");
    const set = (id, value) => document.getElementById(id).textContent = format.format(value);
    const cell = (row, text, className) => { const element = document.createElement("td"); element.textContent = text; if (className) element.className = className; row.append(element); };
    const form = document.getElementById("filters");
    const controls = Array.from(form.elements);
    const previousPage = document.getElementById("previous-page");
    const nextPage = document.getElementById("next-page");
    const records = document.getElementById("records");
    let currentPage = 1;
    let totalPages = 0;

    const filters = () => ({
      author: document.getElementById("author").value.trim(),
      repository: document.getElementById("repository").value.trim(),
      start_date: document.getElementById("start-date").value,
      end_date: document.getElementById("end-date").value
    });

    const setBusy = (busy) => {
      for (const control of controls) control.disabled = busy;
      previousPage.disabled = busy || currentPage <= 1;
      nextPage.disabled = busy || currentPage >= totalPages || totalPages === 0;
    };

    const renderRecords = (data) => {
      set("total-commits", data.total_commits);
      set("ai-commits", data.ai_commits);
      set("ai-lines", data.ai_lines);
      set("total-lines", data.total_lines);
      set("repositories", data.repositories);
      set("contributors", data.contributors);
      records.replaceChildren();
      if (data.records.length === 0) {
        const row = document.createElement("tr");
        const empty = document.createElement("td");
        empty.colSpan = 8;
        empty.className = "empty";
        empty.textContent = "没有找到匹配的提交记录";
        row.append(empty);
        records.append(row);
      } else {
        for (const record of data.records) {
          const row = document.createElement("tr");
          cell(row, record.date);
          cell(row, record.repository_url);
          cell(row, record.author);
          cell(row, record.ai_tool);
          cell(row, record.ai_lines, record.is_ai_commit ? "ai" : "");
          cell(row, record.total_lines);
          cell(row, record.commit_id.slice(0, 12), "commit");
          cell(row, record.message, "message");
          records.append(row);
        }
      }
      currentPage = data.page;
      totalPages = data.total_pages;
      document.getElementById("result-summary").textContent = "共 " + format.format(data.total_records) + " 条，第 " + data.page + " / " + data.total_pages + " 页";
      document.getElementById("status").textContent = "已更新";
    };

    const load = async (page) => {
      const parameters = new URLSearchParams();
      for (const [key, value] of Object.entries(filters())) {
        if (value) parameters.set(key, value);
      }
      parameters.set("page", page);
      parameters.set("page_size", 20);
      setBusy(true);
      try {
        const response = await fetch("/v1/records?" + parameters.toString());
        if (!response.ok) throw new Error("load failed");
        renderRecords(await response.json());
      } catch (_) {
        document.getElementById("status").textContent = "数据暂不可用";
      } finally {
        setBusy(false);
      }
    };

    form.addEventListener("submit", (event) => { event.preventDefault(); load(1); });
    previousPage.addEventListener("click", () => { if (currentPage > 1) load(currentPage - 1); });
    nextPage.addEventListener("click", () => { if (currentPage < totalPages) load(currentPage + 1); });
    load(1);
  </script>
</body>
</html>
`
