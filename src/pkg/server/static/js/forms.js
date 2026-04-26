// Forms.js - Form Handling and Validation

// Form Validation
function validateForm(formId) {
    const form = document.getElementById(formId);
    if (!form) return false;
    
    let isValid = true;
    const errors = [];
    
    // Check required fields
    form.querySelectorAll('[required]').forEach(field => {
        if (!field.value.trim()) {
            isValid = false;
            field.classList.add('error');
            errors.push(`${field.name || field.id} is required`);
        } else {
            field.classList.remove('error');
        }
    });
    
    // Check email fields
    form.querySelectorAll('input[type="email"]').forEach(field => {
        if (field.value && !isValidEmail(field.value)) {
            isValid = false;
            field.classList.add('error');
            errors.push(`Invalid email format: ${field.value}`);
        }
    });
    
    // Check URL fields
    form.querySelectorAll('input[type="url"]').forEach(field => {
        if (field.value && !isValidURL(field.value)) {
            isValid = false;
            field.classList.add('error');
            errors.push(`Invalid URL format: ${field.value}`);
        }
    });
    
    if (!isValid) {
        showToast('Validation Error', errors.join('<br>'), 'error');
    }
    
    return isValid;
}

// Email Validation
function isValidEmail(email) {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email);
}

// URL Validation
function isValidURL(url) {
    try {
        new URL(url);
        return true;
    } catch {
        return false;
    }
}

// Clear Form Errors
function clearFormErrors(formId) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    form.querySelectorAll('.error').forEach(field => {
        field.classList.remove('error');
    });
}

// Serialize Form Data to JSON
function serializeForm(formId) {
    const form = document.getElementById(formId);
    if (!form) return {};
    
    const formData = new FormData(form);
    const data = {};
    
    for (const [key, value] of formData.entries()) {
        // Handle nested objects (dot notation)
        const keys = key.split('.');
        let current = data;
        
        for (let i = 0; i < keys.length - 1; i++) {
            if (!current[keys[i]]) current[keys[i]] = {};
            current = current[keys[i]];
        }
        
        const finalKey = keys[keys.length - 1];
        const element = form.elements[key];
        
        // Handle different input types
        if (element.type === 'checkbox') {
            current[finalKey] = element.checked;
        } else if (element.type === 'number') {
            current[finalKey] = parseFloat(value) || 0;
        } else if (element.type === 'radio') {
            if (element.checked) {
                current[finalKey] = value;
            }
        } else {
            current[finalKey] = value;
        }
    }
    
    return data;
}

// Populate Form from JSON
function populateForm(formId, data) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    function setNestedValue(obj, path) {
        const keys = path.split('.');
        let current = data;
        
        for (const key of keys) {
            if (current && typeof current === 'object' && key in current) {
                current = current[key];
            } else {
                return null;
            }
        }
        
        return current;
    }
    
    // Iterate through all form elements
    Array.from(form.elements).forEach(element => {
        if (!element.name) return;
        
        const value = setNestedValue(data, element.name);
        if (value === null || value === undefined) return;
        
        if (element.type === 'checkbox') {
            element.checked = Boolean(value);
        } else if (element.type === 'radio') {
            if (element.value === String(value)) {
                element.checked = true;
            }
        } else {
            element.value = value;
        }
    });
}

// Auto-save Form (with debounce)
function enableAutoSave(formId, saveCallback, delay = 2000) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    let timeout;
    
    form.addEventListener('input', () => {
        clearTimeout(timeout);
        timeout = setTimeout(() => {
            const data = serializeForm(formId);
            saveCallback(data);
        }, delay);
    });
}

// Disable Form During Submit
function disableForm(formId) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    form.querySelectorAll('input, select, textarea, button').forEach(element => {
        element.disabled = true;
    });
}

// Enable Form
function enableForm(formId) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    form.querySelectorAll('input, select, textarea, button').forEach(element => {
        element.disabled = false;
    });
}

// Show Form Loading State
function showFormLoading(formId, message = 'Saving...') {
    const form = document.getElementById(formId);
    if (!form) return;
    
    const submitBtn = form.querySelector('[type="submit"]');
    if (submitBtn) {
        submitBtn.dataset.originalText = submitBtn.textContent;
        submitBtn.innerHTML = `<span class="spinner"></span> ${message}`;
        submitBtn.disabled = true;
    }
    
    disableForm(formId);
}

// Hide Form Loading State
function hideFormLoading(formId) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    const submitBtn = form.querySelector('[type="submit"]');
    if (submitBtn && submitBtn.dataset.originalText) {
        submitBtn.textContent = submitBtn.dataset.originalText;
        delete submitBtn.dataset.originalText;
        submitBtn.disabled = false;
    }
    
    enableForm(formId);
}

// Confirm Before Submit
function confirmBeforeSubmit(formId, message) {
    const form = document.getElementById(formId);
    if (!form) return;
    
    form.addEventListener('submit', (e) => {
        e.preventDefault();
        
        confirmDialog('Confirm', message, () => {
            // Remove this listener and submit
            form.removeEventListener('submit', arguments.callee);
            form.submit();
        });
    });
}

// Add CSS for error styling
const errorStyle = document.createElement('style');
errorStyle.textContent = `
    .form-control.error {
        border-color: var(--color-danger);
    }
    
    .form-control.error:focus {
        box-shadow: 0 0 0 3px rgba(255, 85, 85, 0.1);
    }
`;
document.head.appendChild(errorStyle);

// Example usage in comments:
/*
// Validate form before submit
if (validateForm('settings-form')) {
    // Form is valid, proceed with submission
}

// Auto-save form
enableAutoSave('settings-form', (data) => {
    console.log('Auto-saving:', data);
});

// Show loading during async operation
showFormLoading('settings-form', 'Saving...');
setTimeout(() => {
    hideFormLoading('settings-form');
    successToast('Saved', 'Settings saved successfully');
}, 2000);

// Populate form from API response
fetch('/api/v1/admin/config')
    .then(r => r.json())
    .then(data => populateForm('settings-form', data));
*/
