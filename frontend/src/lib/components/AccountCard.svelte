<script lang="ts">
    import type { Account, Transaction } from '$lib/types/account';

    interface Props {
        account: Account;
        recentTransactions: Transaction[];
        onAccountClick: (accountNumber: number) => void;
        formatBalance: (balance: number) => string;
        formatDate: (timestamp: number) => string;
    }

    let { account, recentTransactions, onAccountClick, formatBalance, formatDate }: Props = $props();
</script>

<div 
    class="account-card" 
    onclick={() => onAccountClick(account.accountNumber)}
    role="button"
    tabindex="0"
    onkeydown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            onAccountClick(account.accountNumber);
        }
    }}
>
    <div class="account-header">
        <h3 class="account-title">Account #{account.accountNumber}</h3>
        <div class="account-balance">
            {formatBalance(account.balance)}
        </div>
    </div>
    
    <div class="recent-activities">
        <h4 class="activities-title">Recent Activity</h4>
        {#if recentTransactions.length > 0}
            <div class="activity-list">
                {#each recentTransactions as transaction (transaction.transactionId)}
                    <div class="activity-item">
                        <span class="activity-type">{transaction.transactionType}</span>
                        <span 
                            class="activity-amount"
                            class:credit={transaction.amount > 0}
                            class:debit={transaction.amount < 0}
                        >
                            {transaction.amount > 0 ? '+' : ''}{formatBalance(Math.abs(transaction.amount))}
                        </span>
                        <span class="activity-date">{formatDate(transaction.timestamp)}</span>
                    </div>
                {/each}
            </div>
        {:else}
            <p class="no-activity">No recent activity</p>
        {/if}
    </div>

    <!-- Visual indicator that this is clickable -->
    <div class="click-indicator">
        <span class="click-text">Click to view details and actions</span>
        <svg class="click-arrow" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
    </div>

</div>

<style>
    .account-card {
        @apply bg-white rounded-lg shadow-md p-6 cursor-pointer hover:shadow-lg transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2;
    }

    .account-card:hover {
        @apply transform -translate-y-1;
    }

    .account-header {
        @apply flex justify-between items-center mb-4;
    }

    .account-title {
        @apply text-lg font-semibold text-gray-900;
    }

    .account-balance {
        @apply text-xl font-bold text-primary-600;
    }

    .recent-activities {
        @apply space-y-3 mb-4;
    }

    .activities-title {
        @apply text-sm font-medium text-gray-700 mb-3;
    }

    .activity-list {
        @apply space-y-2;
    }

    .activity-item {
        @apply flex justify-between items-center text-sm;
    }

    .activity-type {
        @apply text-gray-600 min-w-0 flex-1;
    }

    .activity-amount {
        @apply font-medium mx-2;
    }

    .activity-amount.credit {
        @apply text-green-600;
    }

    .activity-amount.debit {
        @apply text-red-600;
    }

    .activity-date {
        @apply text-gray-500 text-xs;
    }

    .no-activity {
        @apply text-sm text-gray-500;
    }

    .click-indicator {
        @apply flex items-center justify-center text-sm text-gray-500 mt-4 pt-3 border-t border-gray-200;
    }

    .click-text {
        @apply mr-2;
    }

    .click-arrow {
        @apply w-4 h-4;
    }

    .account-card:hover .click-indicator {
        @apply text-primary-600;
    }
</style>