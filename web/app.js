const notice = document.querySelector("#notice");
const expenses = document.querySelector("#expenses");
const total = document.querySelector("#total");
const byCategory = document.querySelector("#by-category");
const filterForm = document.querySelector("#filter-form");

function showError(message) {
  notice.hidden = false;
  notice.textContent = message.trim();
}

function clearError() {
  notice.hidden = true;
  notice.textContent = "";
}

function queryString() {
  const data = new FormData(filterForm);
  const params = new URLSearchParams();
  for (const [key, value] of data) {
    const text = String(value).trim();
    if (text !== "") {
      params.set(key, text);
    }
  }
  const query = params.toString();
  return query === "" ? "" : `?${query}`;
}

async function readError(response) {
  const text = (await response.text()).trim();
  return text || `request failed (${response.status})`;
}

function renderStats(stats) {
  total.textContent = stats.total;
  byCategory.replaceChildren();
  const entries = Object.entries(stats.by_category);
  if (entries.length === 0) {
    const item = document.createElement("li");
    item.className = "empty";
    item.textContent = "No categories";
    byCategory.append(item);
    return;
  }
  for (const [category, amount] of entries) {
    const item = document.createElement("li");
    item.textContent = `${category} ${amount}`;
    byCategory.append(item);
  }
}

function renderList(rows) {
  expenses.replaceChildren();
  if (rows.length === 0) {
    const item = document.createElement("li");
    item.className = "empty";
    item.textContent = "No expenses";
    expenses.append(item);
    return;
  }
  for (const row of rows) {
    const item = document.createElement("li");
    const amount = document.createElement("span");
    amount.className = "amount";
    amount.textContent = row.amount;
    const category = document.createElement("span");
    category.textContent = row.category;
    const meta = document.createElement("span");
    meta.className = "meta";
    meta.textContent = row.description ? `${row.description} · ${row.created_at}` : row.created_at;
    const button = document.createElement("button");
    button.type = "button";
    button.className = "ghost";
    button.textContent = "Delete";
    button.addEventListener("click", () => removeExpense(row.id));
    item.append(amount, category, meta, button);
    expenses.append(item);
  }
}

async function load() {
  const query = queryString();
  const [listResponse, statsResponse] = await Promise.all([
    fetch(`/expenses${query}`),
    fetch(`/stats${query}`),
  ]);
  if (!listResponse.ok) {
    showError(await readError(listResponse));
    return;
  }
  if (!statsResponse.ok) {
    showError(await readError(statsResponse));
    return;
  }
  clearError();
  renderList(await listResponse.json());
  renderStats(await statsResponse.json());
}

async function removeExpense(id) {
  const response = await fetch(`/expenses/${id}`, { method: "DELETE" });
  if (!response.ok) {
    showError(await readError(response));
    return;
  }
  await load();
}

document.querySelector("#create-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const form = event.currentTarget;
  const data = new FormData(form);
  const body = {
    amount: String(data.get("amount")).trim(),
    category: String(data.get("category")).trim(),
  };
  const description = String(data.get("description")).trim();
  if (description !== "") {
    body.description = description;
  }
  const response = await fetch("/expenses", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    showError(await readError(response));
    return;
  }
  form.reset();
  await load();
});

filterForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  await load();
});

load().catch((error) => showError(error.message));
