// Main Application JS - TEMPLATE.md Compliant (NO JS alerts)

// Theme Management
function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', savedTheme);
    updateThemeIcon(savedTheme);
}

function toggleTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    updateThemeIcon(newTheme);
    
    // Toggle stylesheets
    document.getElementById('theme-dark').disabled = newTheme === 'light';
    document.getElementById('theme-light').disabled = newTheme === 'dark';
}

function updateThemeIcon(theme) {
    const icon = document.querySelector('.icon-theme');
    if (icon) {
        icon.textContent = theme === 'dark' ? '☀️' : '🌙';
    }
}

// User Menu Dropdown
function initUserMenu() {
    const btn = document.getElementById('user-menu-btn');
    const dropdown = document.getElementById('user-dropdown');
    
    if (btn && dropdown) {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            dropdown.classList.toggle('show');
        });
        
        // Close on outside click
        document.addEventListener('click', () => {
            dropdown.classList.remove('show');
        });
    }
}

// Logout. Sessions are stateless bearer tokens with no server-side revocation
// list, so signing out is purely a matter of discarding the local credentials.
function logout() {
    clearAuthToken();
    showToast('Logged out successfully', 'You have been logged out', 'success');
    setTimeout(() => {
        window.location.href = '/';
    }, 1000);
}

// Auth Token Management
function getAuthToken() {
    return localStorage.getItem('auth_token');
}

function setAuthToken(token) {
    localStorage.setItem('auth_token', token);
}

function clearAuthToken() {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user');
}

// Route prefixes are operator-configurable (server.api_version, server.admin_path),
// so they are published as meta tags by head.tmpl rather than hardcoded here.
function metaContent(name, fallback) {
    const tag = document.querySelector(`meta[name="${name}"]`);
    return tag && tag.content ? tag.content : fallback;
}

function apiBase() {
    return '/api/' + metaContent('api-version', 'v1') + '/server';
}

function adminPath() {
    return metaContent('admin-path', 'administration');
}

// API Request Helper
async function apiRequest(url, options = {}) {
    const token = getAuthToken();
    const headers = {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers
    };
    
    try {
        const response = await fetch(url, {
            ...options,
            headers
        });
        
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'Request failed');
        }
        
        return data;
    } catch (error) {
        console.error('API Request Error:', error);
        throw error;
    }
}

// Settings Form Handler
function initSettingsForm() {
    const form = document.getElementById('settings-form');
    if (!form) return;
    
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const formData = new FormData(form);
        const settings = {};
        
        // Build nested object from form data
        for (const [key, value] of formData.entries()) {
            const keys = key.split('.');
            let current = settings;
            
            for (let i = 0; i < keys.length - 1; i++) {
                if (!current[keys[i]]) current[keys[i]] = {};
                current = current[keys[i]];
            }
            
            // Handle checkboxes
            if (form.elements[key].type === 'checkbox') {
                current[keys[keys.length - 1]] = form.elements[key].checked;
            } else if (form.elements[key].type === 'number') {
                current[keys[keys.length - 1]] = parseInt(value);
            } else {
                current[keys[keys.length - 1]] = value;
            }
        }
        
        try {
            await apiRequest(`${apiBase()}/${adminPath()}/config/settings`, {
                method: 'PUT',
                body: JSON.stringify(settings)
            });
            
            showToast('Settings Saved', 'Your settings have been updated successfully', 'success');
        } catch (error) {
            showToast('Save Error', error.message, 'error');
        }
    });
}

// Settings Tabs
function initSettingsTabs() {
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');
    
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const tab = btn.getAttribute('data-tab');
            
            // Remove active from all
            tabBtns.forEach(b => b.classList.remove('active'));
            tabContents.forEach(c => c.classList.remove('active'));
            
            // Add active to clicked
            btn.classList.add('active');
            document.querySelector(`.tab-content[data-tab="${tab}"]`).classList.add('active');
        });
    });
}

// Dashboard Refresh
async function refreshDashboard(btn) {
    if (!btn) return;
    const originalHTML = btn.innerHTML;
    btn.innerHTML = '<span class="spinner"></span> Refreshing...';
    btn.disabled = true;
    
    try {
        // Reload the page to get fresh data
        await new Promise(resolve => setTimeout(resolve, 500));
        window.location.reload();
    } catch (error) {
        showToast('Refresh Error', error.message, 'error');
        btn.innerHTML = originalHTML;
        btn.disabled = false;
    }
}

// Form Reset
function resetForm(formId) {
    const form = document.getElementById(formId || 'settings-form');
    if (!form) return;
    confirmDialog('Reset Form', 'Are you sure you want to reset the form?', () => form.reset());
}

// Initialize on DOMContentLoaded
document.addEventListener('DOMContentLoaded', () => {
    initTheme();
    initUserMenu();
    initSettingsForm();
    initSettingsTabs();
    
    // Theme toggle button
    const themeBtn = document.getElementById('theme-toggle');
    if (themeBtn) {
        themeBtn.addEventListener('click', toggleTheme);
    }
    
    // Auto-hide toasts after 5 seconds
    setTimeout(() => {
        document.querySelectorAll('.toast').forEach(toast => {
            if (!toast.hasAttribute('data-persist')) {
                toast.remove();
            }
        });
    }, 5000);
});

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
    // Escape key closes modals
    if (e.key === 'Escape') {
        document.querySelectorAll('.modal').forEach(modal => {
            modal.style.display = 'none';
        });
    }
    
    // Ctrl/Cmd + K for search (future)
    if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        // Open search modal
    }
});

// Delegated dispatcher for controls that previously carried onclick attributes.
// The Content-Security-Policy allows only nonced script, so inline handlers
// would never fire.
document.addEventListener('click', (e) => {
    const target = e.target.closest('[data-action]');
    if (!target) return;

    switch (target.dataset.action) {
        case 'logout':
            logout();
            break;
        case 'reset-form':
            resetForm(target.dataset.form);
            break;
        case 'refresh-dashboard':
            refreshDashboard(target.closest('button'));
            break;
        case 'toast-close':
            target.closest('.toast').remove();
            break;
        default:
            break;
    }
});
