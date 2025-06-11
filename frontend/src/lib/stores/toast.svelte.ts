interface Toast {
    id: string;
    message: string;
    type: 'success' | 'error' | 'info';
    duration?: number;
}

class ToastStore {
    toasts = $state<Toast[]>([]);

    show(message: string, type: Toast['type'] = 'info', duration = 5000) {
        const id = Math.random().toString(36).substring(2, 2 + 9);
        const toast: Toast = { id, message, type, duration };
        
        this.toasts.push(toast);

        if (duration > 0) {
            setTimeout(() => {
                this.remove(id);
            }, duration);
        }
    }

    remove(id: string) {
        const index = this.toasts.findIndex(t => t.id === id);
        if (index > -1) {
            this.toasts.splice(index, 1);
        }
    }

    success(message: string, duration?: number) {
        this.show(message, 'success', duration);
    }

    error(message: string, duration?: number) {
        this.show(message, 'error', duration);
    }

    info(message: string, duration?: number) {
        this.show(message, 'info', duration);
    }
}

export const toastStore = new ToastStore();