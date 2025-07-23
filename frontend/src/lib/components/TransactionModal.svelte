<script lang="ts">
    import { api } from '$lib/services/api';
    import Button from '$lib/components/ui/Button.svelte';
    import type { Account } from '$lib/types/account';

    interface Props {
        isOpen: boolean;
        transactionType: 'DEPOSIT' | 'WITHDRAWAL' | 'TRANSFER';
        account: Account;
        onClose: () => void;
        onSuccess: () => void;
        accounts?: Account[];
    }

    let { isOpen, transactionType, account, onClose, onSuccess, accounts = [] }: Props = $props();

    let amount = $state<number>(0);
    let destinationAccountNumber = $state<number>(0);
    let isLoading = $state<boolean>(false);
    let error = $state<string | null>(null);

    // Reset form when modal opens
    $effect(() => {
        if (isOpen) {
            amount = 0;
            destinationAccountNumber = 0;
            error = null;
        }
    });

    const modalTitle = $derived(() => {
        switch (transactionType) {
            case 'DEPOSIT': return 'Deposit Money';
            case 'WITHDRAWAL': return 'Withdraw Money';
            case 'TRANSFER': return 'Transfer Money';
            default: return 'Transaction';
        }
    });

    const submitButtonText = $derived(() => {
        switch (transactionType) {
            case 'DEPOSIT': return 'Deposit';
            case 'WITHDRAWAL': return 'Withdraw';
            case 'TRANSFER': return 'Transfer';
            default: return 'Submit';
        }
    });

    function validateForm(): boolean {
        error = null;

        if (amount <= 0) {
            error = 'Amount must be greater than 0';
            return false;
        }

        if (transactionType === 'WITHDRAWAL' && amount > account.balance) {
            error = 'Insufficient funds';
            return false;
        }

        if (transactionType === 'TRANSFER') {
            if (!destinationAccountNumber) {
                error = 'Please select a destination account';
                return false;
            }
            if (destinationAccountNumber === account.accountNumber) {
                error = 'Cannot transfer to the same account';
                return false;
            }
            if (amount > account.balance) {
                error = 'Insufficient funds';
                return false;
            }
        }

        return true;
    }

    async function handleSubmit() {
        if (!validateForm()) return;

        try {
            isLoading = true;
            error = null;

            if (transactionType === 'TRANSFER') {
                // For transfers, we'll need to call a different endpoint
                // TODO: add transfer 
                // await api.createTransaction({
                //     accountId: account.accountId,
                //     amount: -amount,
                //     transactionType: 'TRANSFER_DEBIT'
                // });
                alert("Transfer feature coming soon...");
            } else {
                // For deposits and withdrawals
                const transactionAmount = transactionType === 'DEPOSIT' ? amount : -amount;
                await api.createTransaction({
                    accountId: account.accountId,
                    amount: transactionAmount,
                    transactionType: transactionType
                });
            }

            onSuccess();
            onClose();
        } catch (err) {
            console.error('Transaction failed:', err);
            error = err instanceof Error ? err.message : 'Transaction failed. Please try again.';
        } finally {
            isLoading = false;
        }
    }

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
    >
        <div class="modal-container">
            <div class="modal-header">
                <h2 class="modal-title" id="modal-title">{modalTitle}</h2>
                <button 
                    class="modal-close-btn" 
                    onclick={handleClose}
                    disabled={isLoading}
                    aria-label="Close modal"
                >
                    <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>

            <div class="modal-body">
                <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
                    <div class="form-group">
                        <label for="amount" class="form-label">
                            Amount ($)
                        </label>
                        <input 
                            id="amount" 
                            type="number" 
                            bind:value={amount}
                            step="0.01"
                            min="0.01"
                            max={transactionType === 'DEPOSIT' ? undefined : account.balance}
                            required
                            disabled={isLoading}
                            class="form-input"
                            placeholder="Enter amount"
                        />
                    </div>

                    {#if transactionType === 'TRANSFER'}
                        <div class="form-group">
                            <label for="destination" class="form-label">
                                Destination Account
                            </label>
                            <select 
                                id="destination"
                                bind:value={destinationAccountNumber}
                                required
                                disabled={isLoading}
                                class="form-input"
                            >
                                <option value={0}>Select destination account</option>
                                {#each accounts.filter(acc => acc.accountNumber !== account.accountNumber) as acc}
                                    <option value={acc.accountNumber}>
                                        Account #{acc.accountNumber} (${acc.balance.toLocaleString()})
                                    </option>
                                {/each}
                            </select>
                        </div>
                    {/if}

                    {#if error}
                        <div class="error-message" role="alert">
                            {error}
                        </div>
                    {/if}

                    <div class="form-actions">
                        <Button 
                            type="button" 
                            variant="outline" 
                            onclick={handleClose}
                            disabled={isLoading}
                        >
                            Cancel
                        </Button>
                        <Button 
                            type="submit" 
                            loading={isLoading}
                            disabled={isLoading}
                        >
                            {isLoading ? 'Processing...' : `${submitButtonText} $${amount.toLocaleString()}`}
                        </Button>
                    </div>
                </form>
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
        @apply bg-white rounded-lg shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto;
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
        @apply p-6;
    }

    .form-group {
        @apply mb-4;
    }

    .form-label {
        @apply block text-sm font-medium text-gray-700 mb-2;
    }

    .form-input {
        @apply w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500 disabled:bg-gray-100 disabled:cursor-not-allowed;
    }

    .error-message {
        @apply mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded;
    }

    .form-actions {
        @apply flex gap-3 justify-end;
    }
</style>