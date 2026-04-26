// Modal.js - Custom Modal System (NO JS alerts - TEMPLATE.md compliant)

// Open Modal
function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.style.display = 'block';
        document.body.style.overflow = 'hidden';
        
        // Focus trap
        const focusableElements = modal.querySelectorAll('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
        if (focusableElements.length > 0) {
            focusableElements[0].focus();
        }
    }
}

// Close Modal
function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.style.display = 'none';
        document.body.style.overflow = '';
    }
}

// Create Custom Modal
function createModal(options) {
    const {
        id = 'custom-modal-' + Date.now(),
        title = 'Modal',
        content = '',
        showCancel = true,
        showConfirm = true,
        cancelText = 'Cancel',
        confirmText = 'Confirm',
        onConfirm = () => {},
        onCancel = () => {}
    } = options;
    
    // Remove existing modal with same ID
    const existing = document.getElementById(id);
    if (existing) {
        existing.remove();
    }
    
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.id = id;
    modal.innerHTML = `
        <div class="modal-overlay" onclick="closeModal('${id}')"></div>
        <div class="modal-container">
            <div class="modal-header">
                <h3 class="modal-title">${title}</h3>
                <button class="modal-close" onclick="closeModal('${id}')" aria-label="Close">×</button>
            </div>
            <div class="modal-body">
                ${content}
            </div>
            <div class="modal-footer">
                ${showCancel ? `<button class="btn btn-secondary" onclick="handleModalAction('${id}', 'cancel')">${cancelText}</button>` : ''}
                ${showConfirm ? `<button class="btn btn-primary" onclick="handleModalAction('${id}', 'confirm')">${confirmText}</button>` : ''}
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    
    // Store callbacks
    modal.dataset.onConfirm = onConfirm.toString();
    modal.dataset.onCancel = onCancel.toString();
    
    openModal(id);
    
    return id;
}

// Handle Modal Actions
function handleModalAction(modalId, action) {
    const modal = document.getElementById(modalId);
    if (!modal) return;
    
    if (action === 'confirm' && modal.dataset.onConfirm) {
        try {
            const fn = new Function('return ' + modal.dataset.onConfirm)();
            fn();
        } catch (e) {
            console.error('Error in modal confirm callback:', e);
        }
    } else if (action === 'cancel' && modal.dataset.onCancel) {
        try {
            const fn = new Function('return ' + modal.dataset.onCancel)();
            fn();
        } catch (e) {
            console.error('Error in modal cancel callback:', e);
        }
    }
    
    closeModal(modalId);
    
    // Remove modal after animation
    setTimeout(() => {
        modal.remove();
    }, 300);
}

// Confirm Dialog (replacement for JS confirm)
function confirmDialog(title, message, onConfirm, onCancel) {
    return createModal({
        title,
        content: `<p>${message}</p>`,
        confirmText: 'Yes',
        cancelText: 'No',
        onConfirm: onConfirm || (() => {}),
        onCancel: onCancel || (() => {})
    });
}

// Alert Dialog (replacement for JS alert)
function alertDialog(title, message, onClose) {
    return createModal({
        title,
        content: `<p>${message}</p>`,
        showCancel: false,
        confirmText: 'OK',
        onConfirm: onClose || (() => {})
    });
}

// Prompt Dialog (replacement for JS prompt)
function promptDialog(title, message, defaultValue = '', onConfirm, onCancel) {
    const inputId = 'prompt-input-' + Date.now();
    const content = `
        <p>${message}</p>
        <input type="text" id="${inputId}" class="form-control" value="${defaultValue}" 
               onkeypress="if(event.key==='Enter') handlePromptSubmit('${inputId}', arguments[0])">
    `;
    
    const modalId = createModal({
        title,
        content,
        confirmText: 'OK',
        onConfirm: () => {
            const input = document.getElementById(inputId);
            if (input && onConfirm) {
                onConfirm(input.value);
            }
        },
        onCancel: onCancel || (() => {})
    });
    
    // Focus input after modal opens
    setTimeout(() => {
        const input = document.getElementById(inputId);
        if (input) input.focus();
    }, 100);
    
    return modalId;
}

function handlePromptSubmit(inputId, event) {
    if (event.key === 'Enter') {
        const input = document.getElementById(inputId);
        const modal = input.closest('.modal');
        if (modal) {
            modal.querySelector('.btn-primary').click();
        }
    }
}

// Example usage in comments:
/*
// Replace confirm():
confirmDialog('Delete Project', 'Are you sure you want to delete this project?', () => {
    // User clicked Yes
    console.log('Confirmed');
}, () => {
    // User clicked No
    console.log('Cancelled');
});

// Replace alert():
alertDialog('Success', 'Project created successfully!');

// Replace prompt():
promptDialog('New Project', 'Enter project name:', '', (value) => {
    console.log('User entered:', value);
});
*/
