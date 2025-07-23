<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';
    import { authStore } from '$lib/stores/auth.svelte';
    import { accountStore } from '$lib/stores/account.svelte';
    import type { Account, Transaction } from '$lib/types/account';
    import Button from '$lib/components/ui/Button.svelte';
    import TransactionList from '$lib/components/TransactionList.svelte';
    import TransactionModal from '$lib/components/TransactionModal.svelte';
    import DeleteAccountModal from '$lib/components/DeleteAccountModal.svelte';

    const user = $derived(authStore.user);
    const isLoading = $derived(accountStore.isLoading);
    const accounts = $derived(accountStore.accounts);
    
    let account = $state<Account | null>(null);
    let transactions = $state<Transaction[]>([]);
    let accountNumber = $state<number>(0);
    
    // Modal states
    let showTransactionModal = $state<boolean>(false);
    let transactionType = $state<'DEPOSIT' | 'WITHDRAWAL' | 'TRANSFER'>('DEPOSIT');
    let showDeleteModal = $state<boolean>(false);
    let actionLoading = $state<boolean>(false);

    // Get account number from URL query params
    $effect(() => {
        const url = new URL(window.location.href);
        const accountNumberParam = url.searchParams.get('accountNumber');
        if (accountNumberParam) {
            accountNumber = parseInt(accountNumberParam);
        }
    });

    onMount(async () => {
        if (!user) {
            goto('/login');
            return;
        }

        if (!accountNumber) {
            goto('/dashboard');
            return;
        }

        await loadAccountData();
        // Load all accounts for transfer functionality
        await accountStore.fetchAllAccounts();
    });

    async function loadAccountData() {
        try {
            account = await accountStore.fetchAccountByAccountNumber(accountNumber);
            transactions = await accountStore.fetchTransactionsByAccountId(account?.accountId || '');
        } catch (error) {
            console.error('Failed to load account data:', error);
            goto('/dashboard');
        }
    }

    function formatBalance(balance: number): string {
        return `$${balance.toLocaleString()}`;
    }

    function formatDate(timestamp: number): string {
        return new Date(timestamp * 1000).toLocaleDateString('en-US', {
            month: 'long',
            day: 'numeric',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    function getAccountNumber(accountId: string): number | undefined {
        return account?.accountNumber;
    }

    // Action handlers
    function handleDeposit() {
        transactionType = 'DEPOSIT';
        showTransactionModal = true;
    }

    function handleWithdraw() {
        transactionType = 'WITHDRAWAL';
        showTransactionModal = true;
    }

    function handleTransfer() {
        transactionType = 'TRANSFER';
        showTransactionModal = true;
    }

    function handleDeleteAccount() {
        showDeleteModal = true;
    }

    async function handleTransactionSuccess() {
        // Refresh account data and transactions
        await loadAccountData();
        // Also refresh accounts list for updated balances
        await accountStore.fetchAllAccounts();
    }

    async function handleDeleteConfirm() {
        if (!account) return;

        try {
            actionLoading = true;
            await accountStore.deleteAccount(account.accountNumber);
            goto('/dashboard');
        } catch (error) {
            console.error('Failed to delete account:', error);
            alert('Failed to delete account. Please try again.');
        } finally {
            actionLoading = false;
        }
    }

    function closeTransactionModal() {
        showTransactionModal = false;
    }

    function closeDeleteModal() {
        showDeleteModal = false;
    }
</script>

<svelte:head>
    <title>Account #{accountNumber} - Banking App</title>
</svelte:head>

<div class="account-container">
    <div class="account-header">
        <div class="header-navigation">
            <Button variant="outline" onclick={() => goto('/dashboard')}>
                ← Back to Dashboard
            </Button>
        </div>
        
        {#if account}
            <div class="account-summary">
                <h1 class="account-title">Account #{account.accountNumber}</h1>
                <div class="balance-display">
                    <span class="balance-amount">{formatBalance(account.balance)}</span>
                    <span class="balance-label">Current Balance</span>
                </div>
            </div>

            <!-- Account Actions -->
            <div class="account-actions">
                <Button onclick={handleDeposit} class="action-button deposit">
                    Deposit Money
                </Button>
                <Button variant="secondary" onclick={handleWithdraw} class="action-button withdraw">
                    Withdraw Money
                </Button>
                <Button variant="secondary" onclick={handleTransfer} class="action-button transfer">
                    Transfer Funds
                </Button>
                <Button variant="destructive" onclick={handleDeleteAccount} class="action-button delete">
                    Delete Account
                </Button>
            </div>
        {/if}
    </div>

    <div class="account-content">
        {#if isLoading}
            <div class="loading-state">
                <div class="loading-spinner"></div>
                <p>Loading account details...</p>
            </div>
        {:else if account}
            <div class="transactions-section">
                <h2 class="section-title">Transaction History</h2>
                {#if transactions.length > 0}
                    <TransactionList
                        {transactions}
                        {getAccountNumber}
                        {formatBalance}
                        {formatDate}
                    />
                {:else}
                    <div class="empty-state">
                        <div class="empty-icon">
                            <svg class="empty-icon-svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                        </div>
                        <h3 class="empty-title">No transactions found</h3>
                        <p class="empty-description">This account has no transaction history yet.</p>
                    </div>
                {/if}
            </div>
        {:else}
            <div class="error-state">
                <h3>Account not found</h3>
                <p>The requested account could not be found or you don't have access to it.</p>
                <Button onclick={() => goto('/dashboard')}>
                    Return to Dashboard
                </Button>
            </div>
        {/if}
    </div>
</div>

<!-- Modals -->
{#if account}
    <TransactionModal 
        isOpen={showTransactionModal}
        {transactionType}
        {account}
        accounts={accounts || []}
        onClose={closeTransactionModal}
        onSuccess={handleTransactionSuccess}
    />

    <DeleteAccountModal 
        isOpen={showDeleteModal}
        {account}
        onClose={closeDeleteModal}
        onConfirm={handleDeleteConfirm}
        isLoading={actionLoading}
    />
{/if}

<style>
    .account-container {
        @apply max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8;
    }

    .account-header {
        @apply mb-8 space-y-6;
    }

    .header-navigation {
        @apply flex;
    }

    .account-summary {
        @apply bg-white rounded-lg shadow-sm p-6;
    }

    .account-title {
        @apply text-2xl font-bold text-gray-900 mb-4;
    }

    .balance-display {
        @apply flex flex-col;
    }

    .balance-amount {
        @apply text-4xl font-bold text-primary-600;
    }

    .balance-label {
        @apply text-sm text-gray-500 mt-1;
    }

    .account-actions {
        @apply flex gap-4 flex-wrap;
    }

    .action-button {
        @apply min-w-[140px];
    }

    .account-content {
        @apply space-y-6;
    }

    .transactions-section {
        @apply bg-white rounded-lg shadow-sm p-6;
    }

    .section-title {
        @apply text-xl font-semibold text-gray-900 mb-4;
    }

    .loading-state {
        @apply text-center py-12;
    }

    .loading-spinner {
        @apply inline-block w-8 h-8 border-4 border-gray-200 border-t-primary-600 rounded-full animate-spin mb-4;
    }

    .empty-state {
        @apply text-center py-16;
    }

    .empty-icon {
        @apply mx-auto mb-4;
    }

    .empty-icon-svg {
        @apply h-16 w-16 text-gray-400;
    }

    .empty-title {
        @apply text-xl font-semibold text-gray-900 mb-2;
    }

    .empty-description {
        @apply text-gray-600 mb-4;
    }

    .error-state {
        @apply text-center py-16;
    }
</style>