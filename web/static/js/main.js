// Minimal JavaScript for GoTH stack - Theme management and essential interactions only

// Theme Management (client-side state required)
class ThemeManager {
    constructor() {
        this.theme = localStorage.getItem('theme') || 'light';
        this.init();
    }

    init() {
        this.applyTheme();
        this.setupToggle();
    }

    applyTheme() {
        if (this.theme === 'dark') {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
        localStorage.setItem('theme', this.theme);
    }

    toggle() {
        this.theme = this.theme === 'light' ? 'dark' : 'light';
        this.applyTheme();
    }

    setupToggle() {
        const toggle = document.getElementById('theme-toggle');
        if (toggle) {
            toggle.addEventListener('click', () => this.toggle());
        }
    }
}

// Mobile Menu Management
class MobileMenuManager {
    constructor() {
        this.isOpen = false;
        this.init();
    }

    init() {
        this.setupToggle();
        this.setupCloseOnOutsideClick();
        this.setupKeyboardNavigation();
    }

    setupToggle() {
        const toggle = document.getElementById('mobile-menu-button');
        const menu = document.getElementById('mobile-menu');
        
        if (toggle && menu) {
            toggle.addEventListener('click', () => this.toggle());
        }
    }

    toggle() {
        const menu = document.getElementById('mobile-menu');
        if (!menu) return;

        this.isOpen = !this.isOpen;
        
        if (this.isOpen) {
            menu.classList.remove('hidden');
            menu.classList.add('mobile-menu-enter');
            setTimeout(() => {
                menu.classList.remove('mobile-menu-enter');
                menu.classList.add('mobile-menu-enter-active');
            }, 10);
        } else {
            menu.classList.add('mobile-menu-enter');
            menu.classList.remove('mobile-menu-enter-active');
            setTimeout(() => {
                menu.classList.add('hidden');
                menu.classList.remove('mobile-menu-enter');
            }, 200);
        }
    }

    close() {
        if (this.isOpen) {
            this.toggle();
        }
    }

    setupCloseOnOutsideClick() {
        document.addEventListener('click', (e) => {
            const menu = document.getElementById('mobile-menu');
            const toggle = document.getElementById('mobile-menu-button');
            
            if (this.isOpen && menu && toggle && 
                !menu.contains(e.target) && 
                !toggle.contains(e.target)) {
                this.close();
            }
        });
    }

    setupKeyboardNavigation() {
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen) {
                this.close();
            }
        });
    }
}

// Basic file upload enhancement (for drag & drop)
class FileUploadManager {
    constructor() {
        this.setupFileUpload();
    }

    setupFileUpload() {
        const uploadArea = document.querySelector('.upload-area');
        const fileInput = document.querySelector('#csv_file');

        if (!uploadArea || !fileInput) return;

        // Drag and drop functionality
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('border-blue-500', 'bg-blue-50', 'dark:bg-blue-900');
        });

        uploadArea.addEventListener('dragleave', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('border-blue-500', 'bg-blue-50', 'dark:bg-blue-900');
        });

        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('border-blue-500', 'bg-blue-50', 'dark:bg-blue-900');
            
            const files = e.dataTransfer.files;
            if (files.length > 0 && files[0].name.endsWith('.csv')) {
                fileInput.files = files;
                this.updateFileDisplay(files[0]);
            }
        });

        // File input change
        fileInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) {
                this.updateFileDisplay(e.target.files[0]);
            }
        });
    }

    updateFileDisplay(file) {
        const uploadArea = document.querySelector('.upload-area');
        if (!uploadArea) return;

        const content = uploadArea.querySelector('.space-y-1');
        if (content) {
            content.innerHTML = `
                <svg class="mx-auto h-12 w-12 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
                <div class="text-sm text-gray-600 dark:text-gray-400">
                    <span class="font-medium text-green-600 dark:text-green-400">${file.name}</span>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                    ${this.formatFileSize(file.size)} • Ready to upload
                </p>
            `;
        }
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
}

// HTMX event handlers for enhanced UX
document.addEventListener('htmx:beforeRequest', (event) => {
    // Add loading state
    const target = event.detail.elt;
    if (target) {
        target.classList.add('htmx-loading');
        target.style.pointerEvents = 'none';
        target.style.opacity = '0.7';
    }
});

document.addEventListener('htmx:afterRequest', (event) => {
    // Remove loading state
    const target = event.detail.elt;
    if (target) {
        target.classList.remove('htmx-loading');
        target.style.pointerEvents = '';
        target.style.opacity = '';
    }
});

document.addEventListener('htmx:responseError', (event) => {
    console.error('HTMX Error:', event.detail);
    // Show error message if flash messages container exists
    const flashContainer = document.getElementById('flash-messages');
    if (flashContainer) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'flash-message rounded-lg p-4 mb-2 bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-700';
        errorDiv.innerHTML = `
            <div class="flex items-center">
                <div class="flex-shrink-0">
                    <svg class="w-5 h-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"/>
                    </svg>
                </div>
                <div class="ml-3 flex-1">
                    <p class="text-sm font-medium text-red-800 dark:text-red-200">Network error occurred</p>
                </div>
                <div class="ml-4 flex-shrink-0">
                    <button onclick="this.parentElement.parentElement.parentElement.remove()" class="inline-flex text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600 transition ease-in-out duration-150">
                        <svg class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/>
                        </svg>
                    </button>
                </div>
            </div>
        `;
        flashContainer.appendChild(errorDiv);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            if (errorDiv.parentNode) {
                errorDiv.remove();
            }
        }, 5000);
    }
});

// Auto-remove flash messages
document.addEventListener('htmx:afterSettle', () => {
    const flashMessages = document.querySelectorAll('.flash-message');
    flashMessages.forEach(message => {
        if (!message.dataset.autoRemoved) {
            message.dataset.autoRemoved = 'true';
            setTimeout(() => {
                if (message.parentNode) {
                    message.style.transition = 'opacity 0.5s ease-out';
                    message.style.opacity = '0';
                    setTimeout(() => message.remove(), 500);
                }
            }, 5000);
        }
    });
});

// Keyboard shortcuts for accessibility
document.addEventListener('keydown', (event) => {
    // Escape to close modals
    if (event.key === 'Escape') {
        const modal = document.getElementById('rule-modal');
        if (modal) {
            modal.remove();
        }
    }

    // Ctrl/Cmd + K for search focus
    if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
        event.preventDefault();
        const searchInput = document.querySelector('input[name="search"]');
        if (searchInput) {
            searchInput.focus();
        }
    }
});

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Initialize theme manager
    window.themeManager = new ThemeManager();
    
    // Initialize mobile menu manager
    window.mobileMenuManager = new MobileMenuManager();
    
    // Initialize file upload manager
    window.fileUploadManager = new FileUploadManager();

    // Set active navigation link
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.nav-link');
    navLinks.forEach(link => {
        link.classList.remove('text-gray-900', 'dark:text-gray-100');
        link.classList.add('text-gray-500', 'dark:text-gray-400');
        
        if (link.getAttribute('href') === currentPath) {
            link.classList.remove('text-gray-500', 'dark:text-gray-400');
            link.classList.add('text-gray-900', 'dark:text-gray-100');
        }
    });

    // Smooth scroll for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({ behavior: 'smooth' });
            }
        });
    });
}); 