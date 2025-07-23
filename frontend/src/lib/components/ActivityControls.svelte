<script lang="ts">
    import type { Account } from '$lib/types/account';

    interface Props {
        accounts: Account[] | null;
        filterAccount: string;
        sortBy: 'date' | 'amount' | 'account';
        sortOrder: 'asc' | 'desc';
        onFilterChange: (value: string) => void;
        onSortByChange: (value: 'date' | 'amount' | 'account') => void;
        onSortOrderChange: (value: 'asc' | 'desc') => void;
    }

    let { 
        accounts, 
        filterAccount, 
        sortBy, 
        sortOrder, 
        onFilterChange, 
        onSortByChange, 
        onSortOrderChange 
    }: Props = $props();
</script>

<div class="controls-container">
    <div class="control-group">
        <label for="account-filter" class="control-label">Filter by Account:</label>
        <select 
            id="account-filter" 
            value={filterAccount}
            onchange={(e) => onFilterChange(e.currentTarget.value)}
            class="control-select"
        >
            <option value="all">All Accounts</option>
            {#if accounts}
                {#each accounts as account (account.accountId)}
                    <option value={account.accountId}>
                        Account #{account.accountNumber}
                    </option>
                {/each}
            {/if}
        </select>
    </div>

    <div class="control-group">
        <label for="sort-by" class="control-label">Sort by:</label>
        <select 
            id="sort-by" 
            value={sortBy}
            onchange={(e) => onSortByChange(e.currentTarget.value as 'date' | 'amount' | 'account')}
            class="control-select"
        >
            <option value="date">Date</option>
            <option value="amount">Amount</option>
            <option value="account">Account</option>
        </select>
    </div>

    <div class="control-group">
        <label for="sort-order" class="control-label">Order:</label>
        <select 
            id="sort-order"
            value={sortOrder}
            onchange={(e) => onSortOrderChange(e.currentTarget.value as 'asc' | 'desc')}
            class="control-select"
        >
            <option value="desc">Newest First</option>
            <option value="asc">Oldest First</option>
        </select>
    </div>
</div>

<style>
    .controls-container {
        @apply bg-white rounded-lg shadow-sm p-4 flex flex-wrap gap-4 mb-6;
    }

    .control-group {
        @apply flex flex-col gap-1;
    }

    .control-label {
        @apply text-sm font-medium text-gray-700;
    }

    .control-select {
        @apply px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500;
    }
</style>