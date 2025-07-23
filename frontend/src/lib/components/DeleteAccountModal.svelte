<script lang="ts">
    import Button from '$lib/components/ui/Button.svelte';
    import type { Account } from '$lib/types/account';

    interface Props {
        isOpen: boolean;
        account: Account;
        onClose: () => void;
        onConfirm: () => void;
        isLoading: boolean;
    }

    let { isOpen, account, onClose, onConfirm, isLoading }: Props = $props();

    function handleClose() {
        if (!isLoading) {
            onClose();
        }
    }

    function handleBackdropClick(e: MouseEvent) {
        if (e.target === e.currentTarget) {
            handleClose();
        }
    }

    function handleBackdropKeydown(e: KeyboardEvent) {
        if (e.key === 'Escape') {
            handleClose();
        }
    }
</script>

{#if isOpen}
    <div 
        class="modal-backdrop" 
        onclick={handleBackdropClick}
        onkeydown={handleBackdropKeydown}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        tabindex="-1"
    >        tabindex="-1"

        <div class="modal-container">
            <div class="modal-header">
                <h2 class="modal-title">Delete Account</h2>
                <button 
                    class="modal-close-btn" 
                    onclick={onClose}
                    disabled={isLoading}
                    aria-label="Close"
                >
                    <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>

            <div class="modal-body">
                <div class="warning-icon">
                    <svg class="w-12 h-12 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.728-.833-2.498 0L4.316 16.5c-.77.833.192 2.5 1.732 2.5z" />
                    </svg>
                </div>
hatgpt.com
                <div class="warning-content">
                    <h3 class="warning-title">Are you absolutely sure?</h3>
                    <p class="warning-message">
                        This will permanently delete Account #{account.accountNumber} with a balance of ${account.balance.toLocaleString()}. 
                        This action cannot be undone and will also delete all associated transaction history.
                    </p>
                </div>

                <div class="form-actions">
                    <Button 
                        type="button" 
                        variant="outline" 
                        onclick={onClose}
                        disabled={isLoading}
                    >
                        Cancel
                    </Button>
                    <Button 
                        type="button" 
                        variant="destructive"
                        onclick={onConfirm}
                        loading={isLoading}
                        disabled={isLoading}
                    >
                        {isLoading ? 'Deleting...' : 'Delete Account'}
                    </Button>
                </div>
            </div>
        </div>
    </div>
{/if}

<style>
    @reference '../../app.css';

    .modal-backdrop {
        @apply fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4;
    }

    .modal-container {
        @apply bg-white rounded-lg shadow-xl max-w-md w-full;
    }

    .modal-header {
        @apply flex items-center justify-between p-6 border-b border-gray-200;
    }

    .modal-title {
        @apply text-lg font-semibold text-gray-900;
    }

    .modal-close-btn {
        @apply p-2 hover:bg-gray-100 rounded-full transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
    }

    .modal-body {
        @apply p-6 text-center;
    }

    .warning-icon {
        @apply mb-4;
    }

    .warning-content {
        @apply mb-6;
    }

    .warning-title {
        @apply text-lg font-semibold text-gray-900 mb-2;
    }

    .warning-message {
        @apply text-sm text-gray-600;
    }

    .form-actions {
        @apply flex gap-3 justify-end;
    }
</style>