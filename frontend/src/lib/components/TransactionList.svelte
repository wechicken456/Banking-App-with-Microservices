<script lang="ts">
    import type { Transaction } from '$lib/types/account';

    interface Props {
        transactions: Transaction[];
        getAccountNumber: (accountId: string) => number | undefined;
        formatBalance: (balance: number) => string;
        formatDate: (timestamp: number) => string;
    }

    let { transactions, getAccountNumber, formatBalance, formatDate }: Props = $props();
</script>

<div class="transactions-container">
    {#each transactions as transaction (transaction.transactionId)}
        <div class="transaction-row">
            <div class="transaction-main">
                <div class="transaction-type">
                    {transaction.transactionType}
                </div>
                <div class="transaction-meta">
                    <span class="transaction-account">Account #{getAccountNumber(transaction.accountId)}</span>
                    <span class="transaction-date">{formatDate(transaction.timestamp)}</span>
                </div>
            </div>
            <div 
                class="transaction-amount"
                class:credit={transaction.amount > 0}
                class:debit={transaction.amount < 0}
            >
                {transaction.amount > 0 ? '+' : ''}{formatBalance(Math.abs(transaction.amount))}
            </div>
        </div>
    {/each}
</div>

<style>
    .transactions-container {
        @apply space-y-3;
    }

    .transaction-row {
        @apply bg-white rounded-lg shadow-sm p-4 flex justify-between items-center;
    }

    .transaction-main {
        @apply flex flex-col space-y-1;
    }

    .transaction-type {
        @apply font-medium text-gray-900;
    }

    .transaction-meta {
        @apply flex gap-3 text-sm text-gray-600;
    }

    .transaction-amount {
        @apply text-lg font-semibold;
    }

    .transaction-amount.credit {
        @apply text-green-600;
    }

    .transaction-amount.debit {
        @apply text-red-600;
    }
</style>