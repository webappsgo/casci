// Toast.js - Toast Notification System (NO JS alerts - TEMPLATE.md compliant)

// Show Toast Notification
function showToast(title, message, type = 'info', duration = 5000) {
    const container = document.getElementById('toast-container');
    if (!container) {
        console.error('Toast container not found');
        return;
    }
    
    // Create toast element
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.setAttribute('data-type', type);
    
    const icon = getToastIcon(type);
    
    toast.innerHTML = `
        <div class="toast-icon">${icon}</div>
        <div class="toast-content">
            <div class="toast-title">${title}</div>
            <div class="toast-message">${message}</div>
        </div>
        <button class="toast-close" onclick="this.parentElement.remove()">×</button>
    `;
    
    container.appendChild(toast);
    
    // Auto-remove after duration
    if (duration > 0) {
        setTimeout(() => {
            removeToast(toast);
        }, duration);
    }
    
    return toast;
}

// Get Icon for Toast Type
function getToastIcon(type) {
    const icons = {
        success: '✓',
        error: '✗',
        warning: '⚠',
        info: 'ℹ'
    };
    return icons[type] || icons.info;
}

// Remove Toast with Animation
function removeToast(toast) {
    toast.style.animation = 'toastSlideOut 0.3s ease';
    setTimeout(() => {
        toast.remove();
    }, 300);
}

// Toast Shortcuts
function successToast(title, message, duration) {
    return showToast(title, message, 'success', duration);
}

function errorToast(title, message, duration) {
    return showToast(title, message, 'error', duration);
}

function warningToast(title, message, duration) {
    return showToast(title, message, 'warning', duration);
}

function infoToast(title, message, duration) {
    return showToast(title, message, 'info', duration);
}

// Add slide out animation
const style = document.createElement('style');
style.textContent = `
    @keyframes toastSlideOut {
        to {
            opacity: 0;
            transform: translateX(100%);
        }
    }
`;
document.head.appendChild(style);

// Example usage in comments:
/*
// Basic toast
showToast('Success', 'Operation completed successfully', 'success');

// Using shortcuts
successToast('Saved', 'Settings saved successfully');
errorToast('Error', 'Failed to save settings');
warningToast('Warning', 'This action cannot be undone');
infoToast('Info', 'New updates available');

// Persistent toast (doesn't auto-hide)
showToast('Loading', 'Please wait...', 'info', 0);
*/
