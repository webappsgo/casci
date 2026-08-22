// Modal.js - Custom Modal System (no native alert/confirm/prompt)
// Modals are built with DOM nodes and delegated listeners: the Content-Security-Policy
// serves a per-request nonce and allows neither inline handlers nor eval, so
// onclick attributes and new Function() callbacks would silently never run.

// Callbacks are held here rather than serialized into data attributes so they
// survive as real closures instead of being re-parsed from source text.
const modalCallbacks = new Map();

// Open Modal
function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (!modal) return;

    modal.style.display = 'block';
    document.body.style.overflow = 'hidden';

    const focusable = modal.querySelectorAll('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
    if (focusable.length > 0) {
        focusable[0].focus();
    }
}

// Close Modal
function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (!modal) return;

    modal.style.display = 'none';
    document.body.style.overflow = '';
}

// Build one modal button carrying the delegated action metadata
function createModalButton(modalId, action, label, className) {
    const button = document.createElement('button');
    button.className = className;
    button.textContent = label;
    button.dataset.action = action;
    button.dataset.modal = modalId;
    return button;
}

// Create Custom Modal
// content may be a DOM node or a plain string; strings are inserted as text so
// caller-supplied values can never inject markup.
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

    const existing = document.getElementById(id);
    if (existing) {
        existing.remove();
    }

    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.id = id;

    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay';
    overlay.dataset.action = 'modal-close';
    overlay.dataset.modal = id;

    const container = document.createElement('div');
    container.className = 'modal-container';

    const header = document.createElement('div');
    header.className = 'modal-header';
    const heading = document.createElement('h3');
    heading.className = 'modal-title';
    heading.textContent = title;
    const close = createModalButton(id, 'modal-close', '×', 'modal-close');
    close.setAttribute('aria-label', 'Close');
    header.append(heading, close);

    const body = document.createElement('div');
    body.className = 'modal-body';
    if (content instanceof Node) {
        body.appendChild(content);
    } else if (content) {
        const paragraph = document.createElement('p');
        paragraph.textContent = content;
        body.appendChild(paragraph);
    }

    const footer = document.createElement('div');
    footer.className = 'modal-footer';
    if (showCancel) {
        footer.appendChild(createModalButton(id, 'modal-cancel', cancelText, 'btn btn-secondary'));
    }
    if (showConfirm) {
        footer.appendChild(createModalButton(id, 'modal-confirm', confirmText, 'btn btn-primary'));
    }

    container.append(header, body, footer);
    modal.append(overlay, container);
    document.body.appendChild(modal);

    modalCallbacks.set(id, { onConfirm, onCancel });

    openModal(id);

    return id;
}

// Handle Modal Actions
function handleModalAction(modalId, action) {
    const modal = document.getElementById(modalId);
    if (!modal) return;

    const callbacks = modalCallbacks.get(modalId);
    if (callbacks) {
        const callback = action === 'confirm' ? callbacks.onConfirm : callbacks.onCancel;
        if (typeof callback === 'function') {
            try {
                callback();
            } catch (e) {
                console.error('Error in modal ' + action + ' callback:', e);
            }
        }
    }

    closeModal(modalId);

    // Remove modal after the close animation has run
    setTimeout(() => {
        modalCallbacks.delete(modalId);
        modal.remove();
    }, 300);
}

// Confirm Dialog (replacement for JS confirm)
function confirmDialog(title, message, onConfirm, onCancel) {
    return createModal({
        title,
        content: message,
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
        content: message,
        showCancel: false,
        confirmText: 'OK',
        onConfirm: onClose || (() => {})
    });
}

// Prompt Dialog (replacement for JS prompt)
function promptDialog(title, message, defaultValue = '', onConfirm, onCancel) {
    const inputId = 'prompt-input-' + Date.now();

    const wrapper = document.createElement('div');
    const text = document.createElement('p');
    text.textContent = message;
    const input = document.createElement('input');
    input.type = 'text';
    input.id = inputId;
    input.className = 'form-control';
    input.value = defaultValue;
    // Enter submits the prompt without needing an inline onkeypress attribute
    input.addEventListener('keydown', (e) => {
        if (e.key !== 'Enter') return;
        e.preventDefault();
        const confirmButton = wrapper.closest('.modal').querySelector('[data-action="modal-confirm"]');
        if (confirmButton) confirmButton.click();
    });
    wrapper.append(text, input);

    const modalId = createModal({
        title,
        content: wrapper,
        confirmText: 'OK',
        onConfirm: () => {
            if (onConfirm) {
                onConfirm(input.value);
            }
        },
        onCancel: onCancel || (() => {})
    });

    input.focus();

    return modalId;
}

// Delegated dispatcher for every modal control, including modals injected after
// page load. Registered once at module scope.
document.addEventListener('click', (e) => {
    const target = e.target.closest('[data-action][data-modal]');
    if (!target) return;

    const modalId = target.dataset.modal;
    switch (target.dataset.action) {
        case 'modal-open':
            openModal(modalId);
            break;
        case 'modal-close':
            closeModal(modalId);
            break;
        case 'modal-cancel':
            handleModalAction(modalId, 'cancel');
            break;
        case 'modal-confirm':
            handleModalAction(modalId, 'confirm');
            break;
        default:
            break;
    }
});
