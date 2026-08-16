const state = {
  token: localStorage.getItem("spider_token") || "",
  role: localStorage.getItem("spider_role") || "user",
  user: null,
  bindStatus: null,
  term: ""
};

const $ = (selector) => document.querySelector(selector);
const $$ = (selector) => [...document.querySelectorAll(selector)];
const authView = $("#authView");
const appView = $("#appView");
const loading = $("#loading");
const toast = $("#toast");
let toastTimer;

function showToast(message, type = "success") {
  clearTimeout(toastTimer);
  toast.textContent = message;
  toast.className = `toast show ${type === "error" ? "error" : ""}`;
  toastTimer = setTimeout(() => toast.className = "toast", 3200);
}

function setLoading(show, text = "正在连接教务系统...") {
  loading.querySelector("p").textContent = text;
  loading.classList.toggle("hidden", !show);
}

async function api(path, options = {}) {
  const headers = { ...(options.body ? { "Content-Type": "application/json" } : {}), ...(options.headers || {}) };
  if (state.token) headers.Authorization = `Bearer ${state.token}`;
  let response;
  try {
    response = await fetch(path, { ...options, headers });
  } catch (error) {
    throw new Error("无法连接服务器，请检查 Docker 服务是否运行");
  }
  let payload;
  try {
    payload = await response.json();
  } catch (error) {
    throw new Error(`服务器返回了无法识别的响应 (${response.status})`);
  }
  if (!response.ok || (Object.prototype.hasOwnProperty.call(payload, "code") && payload.code !== 0)) {
    if (payload.code === 40100 || payload.code === 40101) logout(false);
    const error = new Error(payload.message || `请求失败 (${response.status})`);
    error.code = payload.code;
    error.data = payload.data;
    throw error;
  }
  return Object.prototype.hasOwnProperty.call(payload, "data") ? payload.data : payload;
}

function saveToken(token) {
  state.token = token;
  localStorage.setItem("spider_token", token);
  localStorage.setItem("spider_role", state.role);
}

function logout(notify = true) {
  state.token = "";
  state.user = null;
  localStorage.removeItem("spider_token");
  localStorage.removeItem("spider_role");
  appView.classList.add("hidden");
  authView.classList.remove("hidden");
  if (notify) showToast("已退出登录");
}

function setAuthMode(mode) {
  const login = mode === "login";
  $("#loginForm").classList.toggle("hidden", !login);
  $("#registerForm").classList.toggle("hidden", login);
  $("#authTitle").textContent = login ? "登录账户" : "创建账户";
  $("#authSubtitle").textContent = login ? "使用 Spider Go 账户继续" : "先验证邮箱，再创建你的账户";
  $$(".auth-tab").forEach(tab => tab.classList.toggle("active", tab.dataset.authMode === mode));
}

function setAuthRole(role) {
  state.role = role;
  const admin = role === "admin";
  $$(".role-tab").forEach(tab => tab.classList.toggle("active", tab.dataset.authRole === role));
  $("#userAuthTabs").classList.toggle("hidden", admin);
  if (admin) setAuthMode("login");
  $("#authTitle").textContent = admin ? "管理员登录" : "登录账户";
  $("#authSubtitle").textContent = admin ? "进入本地管理控制台" : "使用 Spider Go 账户继续";
  $("#authNote").textContent = admin ? "本地默认管理员：admin@spider-go.com" : "教务密码仅用于向学校教务系统发起查询。";
  $("#loginEmail").placeholder = admin ? "admin@spider-go.com" : "name@example.com";
  if (admin && !$("#loginEmail").value) $("#loginEmail").value = "admin@spider-go.com";
}

async function initializeApp() {
  if (!state.token) return;
  setLoading(true, "正在载入你的控制台...");
  try {
    if (state.role === "admin") {
      await initializeAdmin();
      authView.classList.add("hidden");
      appView.classList.remove("hidden");
      return;
    }
    const [user, bindStatus, termResult] = await Promise.all([
      api("/api/user/info"),
      api("/api/user/bind-status").catch(() => null),
      api("/api/config/term").catch(() => null)
    ]);
    state.user = user;
    state.bindStatus = bindStatus;
    state.term = termResult?.term || "";
    hydrateDashboard();
    setAppRole("user");
    switchView("overview");
    authView.classList.add("hidden");
    appView.classList.remove("hidden");
  } catch (error) {
    logout(false);
    showToast(error.message, "error");
  } finally {
    setLoading(false);
  }
}

async function initializeAdmin() {
  const [admin, userCount, dau, termResult, notices] = await Promise.all([
    api("/api/admin/info"),
    api("/api/admin/statistics/user/count").catch(() => ({ total_count: 0 })),
    api("/api/admin/statistics/dau").catch(() => ({ count: 0, date: "" })),
    api("/api/config/term").catch(() => ({ term: "" })),
    api("/api/admin/notices").catch(() => [])
  ]);
  state.user = admin;
  state.term = termResult?.term || "";
  hydrateAdmin(admin, userCount, dau, notices);
  setAppRole("admin");
  switchView("adminOverview");
}

function setAppRole(role) {
  $$('[data-app-role]').forEach(element => element.classList.toggle("hidden", element.dataset.appRole !== role));
}

function hydrateAdmin(admin, userCount, dau, notices) {
  const name = admin?.name || "管理员";
  $("#profileName").textContent = name;
  $("#profileInitial").textContent = name.slice(0, 1);
  $("#profileSid").textContent = "系统管理员";
  $("#adminUserCount").textContent = userCount?.total_count ?? 0;
  $("#adminDauCount").textContent = dau?.count ?? 0;
  $("#adminDauDate").textContent = dau?.date ? `${dau.date} 活跃用户` : "今日 DAU";
  $("#adminCurrentTerm").textContent = state.term || "未设置";
  $("#adminTermInput").value = state.term;
  $("#datesTermInput").value = state.term;
  $("#adminAccountDetails").innerHTML = detailsMarkup([
    ["姓名", name], ["邮箱", admin?.email || "--"], ["管理员编号", admin?.uid ?? "--"], ["创建时间", formatDate(admin?.created_at)]
  ]);
  renderAdminNotices(notices);
}

function hydrateDashboard() {
  const user = state.user || {};
  const initial = (user.name || "同").slice(0, 1);
  $("#welcomeName").textContent = user.name || "同学";
  $("#profileName").textContent = user.name || "同学";
  $("#profileInitial").textContent = initial;
  $("#profileSid").textContent = user.sid || "未绑定学号";
  $("#emailStat").textContent = user.email || "--";
  $("#currentTermText").textContent = state.term || "尚未设置学期";
  ["gradesTerm", "coursesTerm", "examsTerm"].forEach(id => { if (state.term) $(`#${id}`).value = state.term; });

  const bound = Boolean(user.is_bind || state.bindStatus?.is_bound);
  $("#bindStat").textContent = bound ? "已绑定" : "待绑定";
  $("#bindHint").textContent = bound ? `学号 ${user.sid || state.bindStatus?.current_sid || "已验证"}` : "完成绑定后查询教务数据";
  $("#bindPanel").classList.toggle("hidden", bound);

  $("#accountDetails").innerHTML = detailsMarkup([
    ["姓名", user.name || "--"], ["邮箱", user.email || "--"], ["用户编号", user.uid ?? "--"],
    ["注册时间", formatDate(user.created_at)]
  ]);
  $("#bindDetails").innerHTML = detailsMarkup([
    ["状态", bound ? "已绑定" : "未绑定"], ["学号", user.sid || state.bindStatus?.current_sid || "--"],
    ["累计绑定", state.bindStatus?.total_bind_count ?? "--"], ["最近绑定", formatDate(state.bindStatus?.last_bind_at)]
  ]);
}

function detailsMarkup(items) {
  return items.map(([key, value]) => `<div><dt>${escapeHtml(String(key))}</dt><dd>${escapeHtml(String(value))}</dd></div>`).join("");
}

function formatDate(value) {
  if (!value) return "--";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString("zh-CN", { dateStyle: "medium", timeStyle: "short" });
}

function escapeHtml(value) {
  const node = document.createElement("div");
  node.textContent = value ?? "";
  return node.innerHTML;
}

function switchView(view) {
  const titles = { overview: ["今日概览", `你好，${state.user?.name || "同学"}`], grades: ["ACADEMIC RECORD", "成绩查询"], courses: ["WEEKLY SCHEDULE", "课程表"], exams: ["EXAM PLAN", "考试安排"], account: ["ACCOUNT", "账户信息"], adminOverview: ["ADMIN CONSOLE", `你好，${state.user?.name || "管理员"}`], adminConfig: ["SEMESTER CONFIG", "学期配置"], adminNotices: ["NOTICE MANAGEMENT", "通知管理"], adminAccount: ["ADMIN ACCOUNT", "管理员账户"] };
  $$(".nav-item").forEach(item => item.classList.toggle("active", item.dataset.view === view));
  $$(".view-section").forEach(section => section.classList.toggle("active", section.id === `${view}View`));
  $("#pageEyebrow").textContent = titles[view][0];
  $("#pageTitle").textContent = titles[view][1];
  $(".sidebar").classList.remove("open");
}

function requireBinding() {
  const bound = Boolean(state.user?.is_bind || state.bindStatus?.is_bound);
  if (!bound) {
    showToast("请先在概览页绑定教务系统", "error");
    switchView("overview");
  }
  return bound;
}

$("#loginForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  setLoading(true, "正在登录...");
  try {
    const endpoint = state.role === "admin" ? "/api/admin/login" : "/api/user/login";
    const data = await api(endpoint, { method: "POST", body: JSON.stringify({ email: $("#loginEmail").value.trim(), password: $("#loginPassword").value }) });
    saveToken(data.token);
    await initializeApp();
    showToast("登录成功");
  } catch (error) { showToast(error.message, "error"); }
  finally { setLoading(false); }
});

$("#registerForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  setLoading(true, "正在创建账户...");
  try {
    const data = await api("/api/user/register", { method: "POST", body: JSON.stringify({ name: $("#registerName").value.trim(), email: $("#registerEmail").value.trim(), password: $("#registerPassword").value, captcha: $("#registerCaptcha").value.trim() }) });
    saveToken(data.token);
    await initializeApp();
    showToast("账户创建成功");
  } catch (error) { showToast(error.message, "error"); }
  finally { setLoading(false); }
});

$("#sendCaptchaButton").addEventListener("click", async () => {
  const email = $("#registerEmail").value.trim();
  if (!email) return showToast("请先填写邮箱", "error");
  const button = $("#sendCaptchaButton");
  try {
    await api("/api/captcha/send", { method: "POST", body: JSON.stringify({ email }) });
    showToast("验证码已发送");
    let seconds = 60;
    button.disabled = true;
    const timer = setInterval(() => {
      seconds -= 1;
      button.textContent = `${seconds}s`;
      if (seconds <= 0) { clearInterval(timer); button.disabled = false; button.textContent = "发送验证码"; }
    }, 1000);
  } catch (error) { showToast(error.message, "error"); }
});

function showBindMfa(phone, message) {
  const hint = phone ? `请输入手机 ${phone} 收到的短信验证码` : (message || "请输入绑定手机收到的短信验证码");
  $("#bindMfaHint").textContent = hint;
  $("#bindMfaForm").classList.remove("hidden");
  $("#bindMfaCode").value = "";
  $("#bindMfaCode").focus();
}

function showMfaDialog(phone, message) {
  $("#mfaPhoneHint").textContent = phone ? `请输入手机 ${phone} 收到的短信验证码` : (message || "请输入绑定手机收到的短信验证码");
  $("#mfaCode").value = "";
  $("#mfaDialog").classList.remove("hidden");
  $("#mfaCode").focus();
}

function hideMfaDialog() {
  $("#mfaDialog").classList.add("hidden");
}

function handleApiError(error) {
  if (error.code === 40011) {
    showMfaDialog(error.data?.phone, error.message);
    showToast(error.message);
    return;
  }
  showToast(error.message, "error");
}

async function refreshBoundUser() {
  state.user = await api("/api/user/info");
  state.bindStatus = await api("/api/user/bind-status");
  hydrateDashboard();
}

$("#bindForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  setLoading(true, "正在验证教务账户，这可能需要一点时间...");
  try {
    await api("/api/user/bind", { method: "POST", body: JSON.stringify({ sid: $("#bindSid").value.trim(), spwd: $("#bindPassword").value }) });
    await refreshBoundUser();
    $("#bindMfaForm").classList.add("hidden");
    showToast("教务账户绑定成功");
  } catch (error) {
    if (error.code === 40011) {
      showBindMfa(error.data?.phone, error.message);
      showToast(error.message);
    } else {
      showToast(error.message, "error");
    }
  } finally { setLoading(false); }
});

$("#bindMfaForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  setLoading(true, "正在校验手机验证码...");
  try {
    await api("/api/user/mfa/verify", { method: "POST", body: JSON.stringify({ code: $("#bindMfaCode").value.trim() }) });
    await refreshBoundUser();
    $("#bindMfaForm").classList.add("hidden");
    showToast("手机验证通过，教务账户已绑定");
  } catch (error) { showToast(error.message, "error"); }
  finally { setLoading(false); }
});

$("#bindMfaResend").addEventListener("click", async () => {
  try {
    await api("/api/user/mfa/resend", { method: "POST" });
    showToast("验证码已重新发送");
  } catch (error) { showToast(error.message, "error"); }
});

$("#mfaForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  setLoading(true, "正在校验手机验证码...");
  try {
    await api("/api/user/mfa/verify", { method: "POST", body: JSON.stringify({ code: $("#mfaCode").value.trim() }) });
    hideMfaDialog();
    await refreshBoundUser();
    showToast("手机验证通过，请重新查询");
  } catch (error) { showToast(error.message, "error"); }
  finally { setLoading(false); }
});

$("#mfaResend").addEventListener("click", async () => {
  try {
    await api("/api/user/mfa/resend", { method: "POST" });
    showToast("验证码已重新发送");
  } catch (error) { showToast(error.message, "error"); }
});

$("#gradesFilter").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!requireBinding()) return;
  const term = $("#gradesTerm").value.trim();
  setLoading(true, "正在查询成绩...");
  try {
    const data = await api(`/api/user/grades${term ? `?term=${encodeURIComponent(term)}` : ""}`);
    renderGrades(data);
  } catch (error) { handleApiError(error); }
  finally { setLoading(false); }
});

function renderGrades(data) {
  const grades = data?.grades || [];
  const gpa = data?.gpa || {};
  $("#gradesEmpty").classList.toggle("hidden", grades.length > 0);
  $("#gradesTableWrap").classList.toggle("hidden", grades.length === 0);
  $("#gpaCards").classList.toggle("hidden", grades.length === 0);
  $("#gpaCards").innerHTML = [["平均绩点", gpa.averageGPA ?? gpa.average_gpa ?? "--"], ["平均成绩", gpa.averageScore ?? gpa.average_score ?? "--"], ["基本分", gpa.basicScore ?? gpa.basic_score ?? "--"]].map(([name, value]) => `<article class="metric"><p>${name}</p><strong>${escapeHtml(String(value))}</strong></article>`).join("");
  $("#gradesBody").innerHTML = grades.map(item => `<tr><td><strong>${escapeHtml(item.subject)}</strong><br><small>${escapeHtml(item.code)}</small></td><td>${escapeHtml(item.term)}</td><td><span class="score-badge">${escapeHtml(item.score)}</span></td><td>${item.credit ?? "--"}</td><td>${item.gpa ?? "--"}</td><td>${escapeHtml(item.property || "--")}</td></tr>`).join("");
  if (!grades.length) $("#gradesEmpty").textContent = "没有查询到该学期的成绩。";
}

$("#coursesFilter").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!requireBinding()) return;
  const term = $("#coursesTerm").value.trim();
  const week = $("#coursesWeek").value;
  setLoading(true, "正在查询课表...");
  try {
    const data = await api(`/api/user/courses?week=${encodeURIComponent(week)}&term=${encodeURIComponent(term)}`);
    renderCourses(data);
  } catch (error) { handleApiError(error); }
  finally { setLoading(false); }
});

function renderCourses(data) {
  const weekdays = ["周一", "周二", "周三", "周四", "周五", "周六", "周日"];
  const days = data?.days || [];
  $("#scheduleMeta").classList.remove("hidden");
  $("#scheduleMeta").textContent = `第 ${data?.weekno ?? $("#coursesWeek").value} 周 · ${data?.starttime || "--"} 至 ${data?.endtime || "--"}`;
  $("#coursesGrid").innerHTML = weekdays.map((name, index) => {
    const day = days.find(item => Number(item.weekday) === index + 1) || { courses: [] };
    const cards = (day.courses || []).map(course => `<article class="course-card"><strong>${escapeHtml(course.name)}</strong><span>${escapeHtml(course.teacher || "教师待定")}</span><span>${escapeHtml(course.classroom || "教室待定")}</span><span>第 ${course.start_period}-${course.end_period} 节</span></article>`).join("");
    return `<div class="day-column"><div class="day-heading">${name}</div>${cards || '<div class="no-course">暂无课程</div>'}</div>`;
  }).join("");
}

$("#examsFilter").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!requireBinding()) return;
  const term = $("#examsTerm").value.trim();
  setLoading(true, "正在查询考试安排...");
  try {
    const data = await api(`/api/user/exams?term=${encodeURIComponent(term)}`);
    renderExams(Array.isArray(data) ? data : data?.exams || []);
  } catch (error) { handleApiError(error); }
  finally { setLoading(false); }
});

function renderExams(exams) {
  $("#examsList").innerHTML = exams.length ? exams.map((exam, index) => `<article class="exam-card"><span class="exam-index">${String(index + 1).padStart(2, "0")}</span><div><h4>${escapeHtml(exam.class_name || "未命名课程")}</h4><p>${escapeHtml(exam.time || "时间待定")} · ${escapeHtml(exam.class_no || "")}</p></div><span class="exam-place">${escapeHtml(exam.place || "地点待定")}</span></article>`).join("") : '<div class="empty-state">没有查询到该学期的考试安排。</div>';
}

function renderAdminNotices(notices) {
  const list = $("#adminNoticeList");
  if (!Array.isArray(notices) || notices.length === 0) {
    list.innerHTML = '<div class="empty-state panel">暂无通知，可在左侧发布第一条通知。</div>';
    return;
  }
  list.innerHTML = notices.map(item => `<article class="notice-item"><header><div class="notice-flags">${item.is_top ? "<i>置顶</i>" : ""}${item.is_show ? "<i>显示中</i>" : "<i>已隐藏</i>"}</div><span>${escapeHtml(formatDate(item.create_time))}</span></header><p>${escapeHtml(item.content)}</p></article>`).join("");
}

$("#termForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const term = $("#adminTermInput").value.trim();
  setLoading(true, "正在保存学期配置...");
  try {
    await api("/api/admin/config/term", { method: "POST", body: JSON.stringify({ term }) });
    state.term = term;
    $("#adminCurrentTerm").textContent = term;
    $("#datesTermInput").value = term;
    showToast("当前学期已更新");
  } catch (error) { showToast(error.message, "error"); }
  finally { setLoading(false); }
});

$("#semesterDatesForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  setLoading(true, "正在保存学期日期...");
  try {
    await api("/api/admin/config/semester-dates", { method: "POST", body: JSON.stringify({ term: $("#datesTermInput").value.trim(), start_date: $("#semesterStart").value, end_date: $("#semesterEnd").value }) });
    showToast("学期日期已保存");
  } catch (error) { showToast(error.message, "error"); }
  finally { setLoading(false); }
});

$("#noticeForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  setLoading(true, "正在发布通知...");
  try {
    await api("/api/admin/notices", { method: "POST", body: JSON.stringify({ content: $("#noticeContent").value.trim(), notice_type: $("#noticeType").value.trim(), is_show: $("#noticeShow").checked, is_top: $("#noticeTop").checked, is_html: false }) });
    $("#noticeContent").value = "";
    renderAdminNotices(await api("/api/admin/notices"));
    showToast("通知已发布");
  } catch (error) { showToast(error.message, "error"); }
  finally { setLoading(false); }
});

$("#adminPasswordForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  setLoading(true, "正在更新管理员密码...");
  try {
    await api("/api/admin/reset", { method: "POST", body: JSON.stringify({ old_password: $("#adminOldPassword").value, new_password: $("#adminNewPassword").value }) });
    event.target.reset();
    showToast("管理员密码已更新");
  } catch (error) { showToast(error.message, "error"); }
  finally { setLoading(false); }
});

$$('[data-auth-mode]').forEach(button => button.addEventListener("click", () => setAuthMode(button.dataset.authMode)));
$$('[data-auth-role]').forEach(button => button.addEventListener("click", () => setAuthRole(button.dataset.authRole)));
$$(".nav-item").forEach(button => button.addEventListener("click", () => switchView(button.dataset.view)));
$$('[data-jump]').forEach(button => button.addEventListener("click", () => switchView(button.dataset.jump)));
$("#logoutButton").addEventListener("click", () => logout());
$("#menuButton").addEventListener("click", () => $(".sidebar").classList.toggle("open"));

setAuthRole(state.role);
initializeApp();
