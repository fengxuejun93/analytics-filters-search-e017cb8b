// ========== 全局状态 ==========
let currentView = 'list';
let currentItem = null;
let currentItemId = null;
let savedFilters = {};
let filterOptions = null;
let debounceTimer = null;
let loadingLock = {};  // 防重复提交锁
let cachedFilteredItems = []; // 缓存当前筛选结果列表

// ========== 工具函数 ==========
const API = '/api';

async function api(path, options = {}) {
    let resp;
    try {
        resp = await fetch(API + path, {
            headers: { 'Content-Type': 'application/json' },
            ...options,
        });
    } catch (e) {
        throw new Error('网络请求失败，请检查网络连接后重试');
    }
    let data;
    try {
        data = await resp.json();
    } catch (e) {
        throw new Error('服务器返回格式异常');
    }
    if (!resp.ok) {
        throw new Error(data.error || `请求失败 (${resp.status})`);
    }
    return data;
}

function showToast(msg, type = 'success') {
    const t = document.getElementById('toast');
    t.textContent = msg;
    t.className = 'toast toast-' + type;
    t.style.display = 'block';
    clearTimeout(t._timer);
    t._timer = setTimeout(() => { t.style.display = 'none'; }, 3000);
}

function setLoading(btn, loading) {
    if (!btn) return;
    if (loading) {
        btn._origText = btn.textContent;
        btn.textContent = '处理中...';
        btn.disabled = true;
    } else {
        btn.textContent = btn._origText || btn.textContent;
        btn.disabled = false;
    }
}

function formatTime(ts) {
    if (!ts) return '-';
    const d = new Date(ts);
    if (isNaN(d.getTime())) return '-';
    const now = new Date();
    const diff = now - d;
    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return Math.floor(diff / 60000) + ' 分钟前';
    if (diff < 86400000) return Math.floor(diff / 3600000) + ' 小时前';
    if (diff < 604800000) return Math.floor(diff / 86400000) + ' 天前';
    return d.toLocaleDateString('zh-CN');
}

function formatTimeFull(ts) {
    if (!ts) return '-';
    const d = new Date(ts);
    if (isNaN(d.getTime())) return '-';
    return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
}

const statusMap = {
    listed: '上架中', exchanged: '已置换', delisted: '已下架',
    pending: '待处理', accepted: '已接受', rejected: '已拒绝', cancelled: '已取消',
};
const statusLabel = (s) => statusMap[s] || s || '未知';

const historyIcons = { '发布上架': '🟢', '编辑货品信息': '✏️', '主动下架': '🔴', '重新上架/恢复上架': '🔄' };
function historyIcon(reason) {
    for (const [key, icon] of Object.entries(historyIcons)) { if (reason && reason.includes(key)) return icon; }
    if (reason && reason.includes('发起置换申请')) return '📋';
    if (reason && reason.includes('接受')) return '✅';
    if (reason && reason.includes('拒绝')) return '❌';
    if (reason && reason.includes('取消')) return '↩️';
    if (reason && reason.includes('自动拒绝')) return '⛔';
    return '•';
}

function esc(s) {
    if (s == null) return '';
    const d = document.createElement('div');
    d.textContent = String(s);
    return d.innerHTML;
}

function debounceLoadItems() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(loadItems, 300);
}

// ========== 全量刷新（保持筛选状态） ==========
async function refreshAll() {
    await loadStats();
    if (currentView === 'detail' && currentItemId) {
        await loadDetail(currentItemId);
    } else if (currentView === 'list') {
        await loadItems();
    }
}

// ========== 导航 ==========
function navigateTo(view, data) {
    document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
    currentView = view;
    const el = document.getElementById('view-' + view);
    if (el) el.classList.add('active');
    if (view === 'list') {
        currentItemId = null;
        currentItem = null;
        restoreFilterState();
        loadItems();
        loadStats();
    } else if (view === 'detail' && data) {
        currentItemId = data;
        loadDetail(data);
        loadStats();
    }
}

// ========== 统计 ==========
async function loadStats() {
    try {
        const stats = await api('/stats');
        document.getElementById('stat-listed').textContent = stats.listed_count ?? '-';
        document.getElementById('stat-pending').textContent = stats.pending_count ?? '-';
        document.getElementById('stat-exchanged').textContent = stats.exchanged_count ?? '-';
        document.getElementById('stat-delisted').textContent = stats.delisted_count ?? '-';
        document.getElementById('stat-rejected').textContent = stats.rejected_count ?? '-';
        document.getElementById('stat-cancelled').textContent = stats.cancelled_count ?? '-';
    } catch (e) {
        ['listed','pending','exchanged','delisted','rejected','cancelled'].forEach(k => {
            document.getElementById('stat-' + k).textContent = '--';
        });
    }
}

function filterByStat(status) {
    // pending/rejected/cancelled 是申请维度 → 使用申请状态筛选
    if (status === 'pending' || status === 'rejected' || status === 'cancelled') {
        currentView = 'list';
        document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
        document.getElementById('view-list').classList.add('active');
        currentItemId = null;
        currentItem = null;
        // 清除其他筛选，设置申请状态
        document.getElementById('filter-keyword').value = '';
        document.getElementById('filter-category').value = '';
        document.getElementById('filter-city').value = '';
        document.getElementById('filter-condition').value = '';
        document.getElementById('filter-status').value = '';
        document.getElementById('filter-app-status').value = status;
        loadItems();
        loadStats();
        return;
    }
    currentView = 'list';
    document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
    document.getElementById('view-list').classList.add('active');
    currentItemId = null;
    currentItem = null;
    document.getElementById('filter-keyword').value = '';
    document.getElementById('filter-category').value = '';
    document.getElementById('filter-city').value = '';
    document.getElementById('filter-condition').value = '';
    document.getElementById('filter-app-status').value = '';
    document.getElementById('filter-status').value = status;
    loadItems();
    loadStats();
}

// ========== 筛选器 ==========
async function loadFilterOptions() {
    try {
        filterOptions = await api('/filters');
        const catSel = document.getElementById('filter-category');
        catSel.innerHTML = '<option value="">全部品类</option>';
        (filterOptions.categories || []).forEach(c => { catSel.innerHTML += `<option value="${esc(c)}">${esc(c)}</option>`; });

        const citySel = document.getElementById('filter-city');
        citySel.innerHTML = '<option value="">全部城市</option>';
        (filterOptions.cities || []).forEach(c => { citySel.innerHTML += `<option value="${esc(c)}">${esc(c)}</option>`; });

        const condSel = document.getElementById('filter-condition');
        condSel.innerHTML = '<option value="">全部成色</option>';
        (filterOptions.conditions || []).forEach(c => { condSel.innerHTML += `<option value="${esc(c)}">${esc(c)}</option>`; });
    } catch (e) { console.error('加载筛选选项失败:', e); }
}

function getFilterValues() {
    return {
        keyword: document.getElementById('filter-keyword').value.trim(),
        category: document.getElementById('filter-category').value,
        city: document.getElementById('filter-city').value,
        condition: document.getElementById('filter-condition').value,
        status: document.getElementById('filter-status').value,
        app_status: document.getElementById('filter-app-status').value,
    };
}

function saveFilterState() {
    savedFilters = getFilterValues();
}

function restoreFilterState() {
    if (savedFilters.keyword !== undefined) {
        document.getElementById('filter-keyword').value = savedFilters.keyword;
        document.getElementById('filter-category').value = savedFilters.category;
        document.getElementById('filter-city').value = savedFilters.city;
        document.getElementById('filter-condition').value = savedFilters.condition || '';
        document.getElementById('filter-status').value = savedFilters.status;
        document.getElementById('filter-app-status').value = savedFilters.app_status || '';
    }
}

function clearFilters() {
    document.getElementById('filter-keyword').value = '';
    document.getElementById('filter-category').value = '';
    document.getElementById('filter-city').value = '';
    document.getElementById('filter-condition').value = '';
    document.getElementById('filter-status').value = '';
    document.getElementById('filter-app-status').value = '';
    savedFilters = {};
    loadItems();
}

function clearSingleFilter(field) {
    if (field === 'keyword') {
        document.getElementById('filter-keyword').value = '';
    } else if (field === 'category') {
        document.getElementById('filter-category').value = '';
    } else if (field === 'city') {
        document.getElementById('filter-city').value = '';
    } else if (field === 'condition') {
        document.getElementById('filter-condition').value = '';
    } else if (field === 'status') {
        document.getElementById('filter-status').value = '';
    } else if (field === 'app_status') {
        document.getElementById('filter-app-status').value = '';
    }
    loadItems();
}

function hasActiveFilters() {
    const f = getFilterValues();
    return !!f.keyword || !!f.category || !!f.city || !!f.condition || !!f.status || !!f.app_status;
}

// 渲染激活的筛选标签
function renderActiveFilterTags() {
    const container = document.getElementById('active-filters');
    const f = getFilterValues();
    const tags = [];

    const labelMap = {
        keyword: '关键词',
        category: '品类',
        city: '城市',
        condition: '成色',
        status: '货品状态',
        app_status: '申请状态',
    };

    for (const [key, label] of Object.entries(labelMap)) {
        const val = f[key];
        if (val) {
            const displayVal = statusLabel(val) || val;
            tags.push(`<span class="filter-tag">${esc(label)}: ${esc(displayVal)} <span class="filter-tag-remove" onclick="clearSingleFilter('${key}')" title="移除此筛选">&times;</span></span>`);
        }
    }

    if (tags.length > 0) {
        container.innerHTML = tags.join('');
        container.style.display = 'flex';
    } else {
        container.innerHTML = '';
        container.style.display = 'none';
    }
}

// ========== 货品列表 ==========
async function loadItems() {
    const params = new URLSearchParams();
    const f = getFilterValues();
    if (f.keyword) params.set('keyword', f.keyword);
    if (f.category) params.set('category', f.category);
    if (f.city) params.set('city', f.city);
    if (f.condition) params.set('condition', f.condition);
    if (f.status) params.set('status', f.status);
    if (f.app_status) params.set('app_status', f.app_status);

    const grid = document.getElementById('item-grid');
    const empty = document.getElementById('empty-state');
    const resultCount = document.getElementById('result-count');

    // 渲染筛选标签
    renderActiveFilterTags();

    // 重置空状态为默认内容
    function resetEmptyState() {
        empty.innerHTML = '<p>暂无符合条件的货品</p><button class="btn btn-secondary btn-sm" onclick="clearFilters()" style="margin-top:8px;">清除筛选，查看全部</button>';
    }

    try {
        const resp = await api('/items?' + params.toString());
        const items = resp.items || (Array.isArray(resp) ? resp : []);
        const totalCount = resp.total_count ?? items.length;

        // 缓存筛选结果
        cachedFilteredItems = items;

        if (hasActiveFilters() && totalCount > 0) {
            resultCount.textContent = `当前结果: ${items.length} / 全部: ${totalCount}`;
            resultCount.style.display = 'block';
        } else if (hasActiveFilters() && totalCount === 0) {
            resultCount.textContent = `当前结果: 0 / 全部: ${totalCount}`;
            resultCount.style.display = 'block';
        } else {
            resultCount.style.display = 'none';
        }

        if (!items.length) {
            grid.innerHTML = '';
            resetEmptyState();
            empty.style.display = 'block';
            return;
        }
        empty.style.display = 'none';
        grid.innerHTML = items.map(item => `
            <div class="item-card ${currentItemId === item.id ? 'item-card-active' : ''}" onclick="openDetail('${esc(item.id)}')">
                <div class="item-card-image">📷</div>
                <div class="item-card-body">
                    <div class="item-card-title">${esc(item.title || '无标题')}</div>
                    <div class="item-card-tags">
                        <span class="tag tag-category">${esc(item.category || '未分类')}</span>
                        <span class="tag tag-condition">${esc(item.condition || '未填写')}</span>
                        <span class="tag tag-city">${esc(item.city || '未填写')}</span>
                    </div>
                    <div class="item-card-bottom">
                        <span class="item-card-exchange">期望: ${esc(item.expected_exchange || '不限')}</span>
                        <span class="status-badge status-${item.status || 'listed'}">${statusLabel(item.status)}</span>
                    </div>
                    <div class="item-card-bottom" style="margin-top:6px;">
                        <span>${esc(item.publisher || '-')}</span>
                        <span>${formatTime(item.updated_at)}</span>
                    </div>
                </div>
            </div>
        `).join('');
    } catch (e) {
        grid.innerHTML = '';
        resultCount.style.display = 'none';
        empty.innerHTML = `<p>加载失败: ${esc(e.message)}</p><button class="btn btn-secondary btn-sm" onclick="loadItems()" style="margin-top:8px;">重新加载</button> <button class="btn btn-secondary btn-sm" onclick="clearFilters()" style="margin-top:8px;">清除筛选，查看全部</button>`;
        empty.style.display = 'block';
    }
}

function openDetail(itemId) {
    saveFilterState();
    navigateTo('detail', itemId);
}

// ========== 货品详情 ==========
async function loadDetail(itemId) {
    const container = document.querySelector('.detail-layout');
    try {
        const detail = await api('/items/' + itemId + '/detail');
        if (!detail || !detail.item) throw new Error('详情数据异常');
        currentItem = detail.item;
        currentItemId = detail.item.id;
        renderDetail(detail);
    } catch (e) {
        showToast(e.message, 'error');
        const nav = document.querySelector('.detail-nav');
        const errorHtml = `<div class="detail-error"><p>加载详情失败: ${esc(e.message)}</p><button class="btn btn-primary btn-sm" onclick="loadDetail('${esc(itemId)}')">重新加载</button> <button class="btn btn-secondary btn-sm" onclick="navigateTo('list')">返回列表</button></div>`;
        if (container) container.innerHTML = errorHtml;
    }
}

function renderDetail(detail) {
    const item = detail.item;
    const apps = detail.applications || [];
    const history = detail.status_history || [];

    document.getElementById('detail-nav-item').textContent = item.title || '货品详情';
    document.getElementById('detail-image').textContent = '📷';
    document.getElementById('detail-title').textContent = item.title || '无标题';

    document.getElementById('detail-meta').innerHTML = `
        <span class="tag tag-category">${esc(item.category || '未分类')}</span>
        <span class="tag tag-condition">${esc(item.condition || '未填写')}</span>
        <span class="tag tag-city">${esc(item.city || '未填写')}</span>
        <span class="status-badge status-${item.status || 'listed'}">${statusLabel(item.status)}</span>
    `;

    document.getElementById('detail-fields').innerHTML = `
        <div class="field-row"><span class="field-label">品类</span><span class="field-value">${esc(item.category || '-')}</span></div>
        <div class="field-row"><span class="field-label">成色</span><span class="field-value">${esc(item.condition || '-')}</span></div>
        <div class="field-row"><span class="field-label">所在城市</span><span class="field-value">${esc(item.city || '-')}</span></div>
        <div class="field-row"><span class="field-label">发布人</span><span class="field-value">${esc(item.publisher || '-')}</span></div>
        <div class="field-row"><span class="field-label">期望置换物</span><span class="field-value">${esc(item.expected_exchange || '不限')}</span></div>
        <div class="field-row"><span class="field-label">发布时间</span><span class="field-value">${formatTimeFull(item.created_at)}</span></div>
        <div class="field-row"><span class="field-label">最后更新</span><span class="field-value">${formatTimeFull(item.updated_at)}</span></div>
    `;

    document.getElementById('detail-desc').textContent = item.description || '暂无描述';
    document.getElementById('detail-exchange').textContent = item.expected_exchange || '不限';

    renderDetailActions(item);
    renderApplications(apps, item);
    renderStatusHistory(history);
}

function renderDetailActions(item) {
    const actions = document.getElementById('detail-actions');
    const hint = document.getElementById('detail-status-hint');
    let btns = '';
    let hintText = '';

    switch (item.status) {
        case 'listed':
            btns += `<button class="btn btn-sm btn-primary" onclick="showEditForm('${item.id}')">编辑</button>`;
            btns += `<button class="btn btn-sm btn-secondary" onclick="delistItem('${item.id}')">下架</button>`;
            break;
        case 'delisted':
            btns += `<button class="btn btn-sm btn-success" onclick="relistItem('${item.id}')">重新上架</button>`;
            btns += `<button class="btn btn-sm btn-primary" onclick="showEditForm('${item.id}')">编辑</button>`;
            hintText = '此货品已下架，无法发起新的置换申请。可重新上架或编辑信息。';
            break;
        case 'exchanged':
            hintText = '此货品已达成置换，不可编辑、下架或发起新申请。如需恢复，请取消已接受的置换申请。';
            break;
        default:
            hintText = `货品状态异常 (${esc(item.status)})`;
    }

    actions.innerHTML = btns;
    if (hintText) {
        hint.innerHTML = hintText;
        hint.style.display = 'flex';
    } else {
        hint.style.display = 'none';
    }
}

// ========== 置换申请渲染 ==========
function renderApplications(apps, item) {
    const pendingCount = apps.filter(a => a.status === 'pending').length;
    document.getElementById('detail-app-count').textContent = pendingCount;
    document.getElementById('detail-app-count').title = `共 ${apps.length} 条申请，${pendingCount} 条待处理`;

    const applyToggle = document.getElementById('btn-apply-toggle');
    applyToggle.style.display = item.status === 'listed' ? '' : 'none';
    document.getElementById('apply-form').style.display = 'none';

    const list = document.getElementById('application-list');
    if (!apps.length) {
        list.innerHTML = '<p class="empty-hint">暂无置换申请</p>';
        return;
    }

    const statusColors = { pending: '#d97706', accepted: '#16a34a', rejected: '#dc2626', cancelled: '#64748b' };
    const statusBgs = { pending: '#fef3c7', accepted: '#dcfce7', rejected: '#fee2e2', cancelled: '#f1f5f9' };

    list.innerHTML = apps.map(a => {
        let actionBtns = '';
        if (a.status === 'pending' && item.status === 'listed') {
            actionBtns = `
                <button class="btn btn-sm btn-success" onclick="handleApp(this,'${a.id}','accept')">接受</button>
                <button class="btn btn-sm btn-danger" onclick="handleApp(this,'${a.id}','reject')">拒绝</button>
                <button class="btn btn-sm btn-secondary" onclick="handleApp(this,'${a.id}','cancel')">取消</button>
            `;
        } else if (a.status === 'accepted') {
            actionBtns = `<button class="btn btn-sm btn-warning" onclick="handleApp(this,'${a.id}','cancel')">取消置换</button>`;
        } else if (a.status === 'pending' && item.status !== 'listed') {
            actionBtns = `<span class="op-hint">货品已${statusLabel(item.status)}，此申请暂不可操作</span>`;
        } else if (a.status === 'rejected') {
            actionBtns = `<span class="op-hint">已拒绝，不可操作</span>`;
        } else if (a.status === 'cancelled') {
            actionBtns = `<span class="op-hint">已取消，不可操作</span>`;
        } else {
            actionBtns = `<span class="op-hint">状态异常</span>`;
        }

        const offerText = a.offer_item ? `提供: ${esc(a.offer_item)}` : '未指定置换物';

        return `
            <div class="app-card app-${a.status || 'pending'}">
                <div class="app-header-row">
                    <span class="app-applicant">${esc(a.applicant || '匿名')}</span>
                    <span class="app-status-badge" style="color:${statusColors[a.status]||'#64748b'};background:${statusBgs[a.status]||'#f1f5f9'}">${statusLabel(a.status)}</span>
                </div>
                <div class="app-offer">${offerText}</div>
                ${a.message ? `<div class="app-message">${esc(a.message)}</div>` : ''}
                <div class="app-time">申请于 ${formatTimeFull(a.created_at)}${a.updated_at && a.updated_at !== a.created_at ? ' · 处理于 ' + formatTimeFull(a.updated_at) : ''}</div>
                <div class="app-actions">${actionBtns}</div>
            </div>
        `;
    }).join('');
}

// ========== 状态历史时间线 ==========
function renderStatusHistory(histories) {
    const el = document.getElementById('status-history');
    if (!histories || !histories.length) {
        el.innerHTML = '<p class="empty-hint">暂无状态变更记录</p>';
        return;
    }
    el.innerHTML = '<div class="timeline">' + histories.map((h, i) => {
        const isLast = i === histories.length - 1;
        const fromLabel = h.from_status ? statusLabel(h.from_status) : '新建';
        const toLabel = statusLabel(h.to_status);
        const icon = historyIcon(h.reason);
        const isStateChange = h.from_status !== h.to_status;
        return `
            <div class="timeline-item${isLast ? ' timeline-item-active' : ''}">
                <div class="timeline-dot${isStateChange ? ' timeline-dot-event' : ''}"></div>
                <div class="timeline-content">
                    <div class="timeline-title">${icon} ${isStateChange ? esc(fromLabel) + ' → ' + esc(toLabel) : esc(toLabel)}</div>
                    <div class="timeline-reason">${esc(h.reason || '-')}</div>
                    <div class="timeline-meta">${esc(h.operator || '-')} · ${formatTime(h.created_at)}</div>
                </div>
            </div>
        `;
    }).join('') + '</div>';
}

// ========== 申请交互 ==========
function showApplyForm() {
    if (!currentItem || currentItem.status !== 'listed') {
        showToast('该货品当前状态不允许发起申请', 'error');
        return;
    }
    const f = document.getElementById('apply-form');
    f.style.display = f.style.display === 'none' ? 'flex' : 'none';
}

async function submitApplication() {
    const btn = event.target;
    if (loadingLock['apply']) return;
    const applicant = document.getElementById('apply-applicant').value.trim();
    const offerItem = document.getElementById('apply-offer-item').value.trim();
    const message = document.getElementById('apply-message').value.trim();
    if (!applicant) { showToast('请填写申请人名字', 'error'); return; }
    if (!offerItem) { showToast('请填写你能提供的置换物', 'error'); return; }

    loadingLock['apply'] = true;
    setLoading(btn, true);
    try {
        await api('/items/' + currentItem.id + '/applications', {
            method: 'POST',
            body: JSON.stringify({ applicant, offer_item: offerItem, message }),
        });
        showToast('申请已提交');
        document.getElementById('apply-applicant').value = '';
        document.getElementById('apply-offer-item').value = '';
        document.getElementById('apply-message').value = '';
        document.getElementById('apply-form').style.display = 'none';
        refreshAll();
    } catch (e) {
        showToast(e.message, 'error');
    } finally {
        loadingLock['apply'] = false;
        setLoading(btn, false);
    }
}

async function handleApp(btn, appId, action) {
    if (btn && btn.disabled) return;
    if (loadingLock[appId]) return;
    loadingLock[appId] = true;
    if (btn) setLoading(btn, true);
    try {
        await api('/applications/' + appId, { method: 'PUT', body: JSON.stringify({ action }) });
        const actionLabels = { accept: '已接受', reject: '已拒绝', cancel: '已取消' };
        showToast(actionLabels[action] || '操作成功');
        // 操作后检查当前详情项是否仍符合筛选条件
        await afterDetailAction();
    } catch (e) {
        showToast(e.message, 'error');
    } finally {
        loadingLock[appId] = false;
        if (btn) setLoading(btn, false);
    }
}

// ========== 详情操作后：刷新数据并检查筛选匹配 ==========
async function afterDetailAction() {
    await loadStats();
    if (currentView === 'detail' && currentItemId) {
        // 重新加载详情
        try {
            const detail = await api('/items/' + currentItemId + '/detail');
            if (!detail || !detail.item) throw new Error('详情数据异常');
            currentItem = detail.item;
            renderDetail(detail);
        } catch (e) {
            showToast(e.message, 'error');
            return;
        }

        // 检查当前详情项是否仍符合筛选条件
        if (hasActiveFilters() && !itemMatchesFilters(currentItem)) {
            showToast('当前货品已不再符合筛选条件，即将返回列表', 'info');
            setTimeout(() => {
                navigateTo('list');
            }, 1200);
        }
    } else if (currentView === 'list') {
        await loadItems();
    }
}

// 检查一个货品是否符合当前筛选条件（前端侧校验，用于操作后判断）
function itemMatchesFilters(item) {
    const f = getFilterValues();
    if (f.keyword) {
        const kw = f.keyword.toLowerCase();
        if (!(item.title || '').toLowerCase().includes(kw) &&
            !(item.description || '').toLowerCase().includes(kw) &&
            !(item.expected_exchange || '').toLowerCase().includes(kw)) {
            return false;
        }
    }
    if (f.category && item.category !== f.category) return false;
    if (f.city && item.city !== f.city) return false;
    if (f.condition && item.condition !== f.condition) return false;
    if (f.status && item.status !== f.status) return false;
    // app_status 需要后端数据，此处仅做简化判断
    // 因为前端缓存了 filteredItems，可以用 ID 判断
    if (f.app_status) {
        return cachedFilteredItems.some(i => i.id === item.id);
    }
    return true;
}

// ========== 货品状态操作 ==========
async function delistItem(id) {
    try {
        await api('/items/' + id + '/status', { method: 'PUT', body: JSON.stringify({ status: 'delisted' }) });
        showToast('已下架');
        await afterDetailAction();
    } catch (e) { showToast(e.message, 'error'); }
}

async function relistItem(id) {
    try {
        await api('/items/' + id + '/status', { method: 'PUT', body: JSON.stringify({ status: 'listed' }) });
        showToast('已重新上架');
        await afterDetailAction();
    } catch (e) { showToast(e.message, 'error'); }
}

// ========== 创建/编辑表单 ==========
function showCreateForm() {
    document.getElementById('form-title').textContent = '发布新货品';
    document.getElementById('form-item-id').value = '';
    document.getElementById('item-form').reset();
    document.getElementById('form-back-btn').onclick = () => navigateTo('list');
    document.getElementById('form-cancel-btn').onclick = () => navigateTo('list');
    navigateTo('form');
}

function showEditForm(id) {
    const item = currentItem;
    if (!item) return;
    if (item.status === 'exchanged') { showToast('已置换的货品不可编辑', 'error'); return; }
    document.getElementById('form-title').textContent = '编辑货品';
    document.getElementById('form-item-id').value = item.id;
    document.getElementById('form-field-title').value = item.title || '';
    document.getElementById('form-field-category').value = item.category || '';
    document.getElementById('form-field-condition').value = item.condition || '';
    document.getElementById('form-field-city').value = item.city || '';
    document.getElementById('form-field-exchange').value = item.expected_exchange || '';
    document.getElementById('form-field-publisher').value = item.publisher || '';
    document.getElementById('form-field-description').value = item.description || '';
    // 编辑返回时保持筛选状态
    document.getElementById('form-back-btn').onclick = () => { restoreFilterState(); navigateTo('detail', item.id); };
    document.getElementById('form-cancel-btn').onclick = () => { restoreFilterState(); navigateTo('detail', item.id); };
    saveFilterState();
    navigateTo('form');
}

async function submitItemForm(e) {
    e.preventDefault();
    const submitBtn = e.target.querySelector('button[type=submit]');
    if (submitBtn && submitBtn.disabled) return;

    const id = document.getElementById('form-item-id').value;
    const payload = {
        title: document.getElementById('form-field-title').value.trim(),
        category: document.getElementById('form-field-category').value,
        condition: document.getElementById('form-field-condition').value,
        city: document.getElementById('form-field-city').value.trim(),
        expected_exchange: document.getElementById('form-field-exchange').value.trim(),
        publisher: document.getElementById('form-field-publisher').value.trim(),
        description: document.getElementById('form-field-description').value.trim(),
    };

    if (!payload.title || !payload.category || !payload.publisher) {
        showToast('标题、品类、发布人为必填项', 'error');
        return;
    }

    setLoading(submitBtn, true);
    try {
        if (id) {
            await api('/items/' + id, { method: 'PUT', body: JSON.stringify(payload) });
            showToast('编辑成功');
            // 编辑后回到详情，保持筛选
            restoreFilterState();
            navigateTo('detail', id);
        } else {
            const newItem = await api('/items', { method: 'POST', body: JSON.stringify(payload) });
            showToast('发布成功');
            restoreFilterState();
            navigateTo('detail', newItem.id);
        }
    } catch (e) {
        showToast(e.message, 'error');
    } finally {
        setLoading(submitBtn, false);
    }
}

// ========== 初始化 ==========
document.addEventListener('DOMContentLoaded', () => {
    const urlEl = document.getElementById('header-url');
    urlEl.innerHTML = '<strong>' + esc(window.location.origin) + '</strong> ← 访问入口';

    loadFilterOptions();
    loadItems();
    loadStats();
});
