<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth.svelte';
	import { accountStore } from '$lib/stores/account.svelte';
	import type { Account } from '$lib/types/account';
	import type { Transaction } from '$lib/types/account';
	import Button from '$lib/components/ui/Button.svelte';
	import AccountCard from '$lib/components/AccountCard.svelte';
	import ActivityControls from '$lib/components/ActivityControls.svelte';
	import TransactionList from '$lib/components/TransactionList.svelte';

	const user = $derived(authStore.user);
	const accounts = $derived(accountStore.accounts);
	const isLoading = $derived(accountStore.isLoading);

	let activeTab = $state<'accounts' | 'activities'>('accounts');
	let allTransactions = $state<Transaction[]>([]);
	let sortBy = $state<'date' | 'amount' | 'account'>('date');
	let sortOrder = $state<'asc' | 'desc'>('desc');
	let filterAccount = $state<string>('all');

	onMount(async () => {
		if (!user) {
			goto('/login');
		}
		try {
			await accountStore.fetchAllAccounts();
			if (activeTab == 'activities') {
				await loadAllTransactions();
			}
		} catch (err) {
			console.error('Failed to load dashboard data: ', err);
		}
	});

	async function loadAllTransactions() {
		if (!accounts || accounts.length === 0) return;

		const transactionPromises = accounts.map((account) =>
			accountStore.fetchTransactionsByAccountId(account.accountId)
		);

		try {
			const transactionArrays = await Promise.all(transactionPromises);
			allTransactions = transactionArrays.flat();
		} catch (error) {
			console.error('Failed to load transactions:', error);
			return null;
		}
	}


	// filteredTransactions as the sorted and filtered list
    let filteredTransactions = $state<Transaction[]>([]);
    // Create a reactive effect to update filteredTransactions
    // sort and filter transactions based on user selection
    $effect(() => {
        let filtered = [...allTransactions];

        // if we're only showing transactions for a specific account
        if (filterAccount !== 'all') {
            filtered = filtered.filter((tx) => tx.accountId === filterAccount);
        }

        // sort based on selected criteria, default to descending order and sorting by date.
        filtered.sort((a, b) => {
            let comparison = 0;
            if (sortBy === 'date') {
                comparison = a.timestamp - b.timestamp;
            } else if (sortBy === 'amount') {
                comparison = a.amount - b.amount;
            }
            return sortOrder === 'desc' ? -comparison : comparison;
        });

        filteredTransactions = filtered;
    });

	function handleTabChange(tab: 'accounts' | 'activities') {
		activeTab = tab;
		if (tab === 'activities') {
			loadAllTransactions();
		}
	}

	function formatBalance(balance: number): string {
		return `$${balance.toLocaleString()}`;
	}

	function formatDate(timestamp: number): string {
		return new Date(timestamp * 1000).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function getAccountNumber(accountId: string): string | undefined {
		if (accounts.length == 0) return undefined;
		return accounts.find((acc) => acc.accountId === accountId)?.accountNumber;
	}

	function getTotalBalance(): number {
		if (accounts.length == 0) return 0;
		return accounts.reduce((sum, account) => sum + account.balance, 0) || 0;
	}

	function getRecentTransactions(accountId: string): Transaction[] {
		return allTransactions
			.filter((tx) => tx.accountId === accountId)
			.sort((a, b) => b.timestamp - a.timestamp)
			.slice(0, 3);
	}

	// Component event handlers
	function handleAccountClick(accountNumber: string) {
		goto(`/account?accountNumber=${accountNumber}`);
	}

	function handleFilterChange(value: string) {
		filterAccount = value;
	}

	function handleSortByChange(value: 'date' | 'amount' | 'account') {
		sortBy = value;
	}

	function handleSortOrderChange(value: 'asc' | 'desc') {
		sortOrder = value;
	}
</script>

<svelte:head>
	<title>Dashboard - Banking App</title>
</svelte:head>

<div class="dashboard-container">
	<div class="dashboard-header">
		<h1 class="dashboard-title">Dashboard</h1>
		<p class="dashboard-subtitle">Welcome back, {user?.email}!</p>

		<!-- Total Balance Summary -->
		<div class="total-balance-card">
			<div class="total-balance">
				<span class="balance-amount">{formatBalance(getTotalBalance())}</span>
				<span class="balance-label">Total Balance</span>
			</div>
			<Button onclick={() => goto('/account/create')}>Create Account</Button>
		</div>
	</div>

	<!-- Tab Navigation -->
	<nav class="tab-navigation">
		<button
			class="tab-button"
			class:active={activeTab === 'accounts'}
			onclick={() => handleTabChange('accounts')}
		>
			All Accounts
		</button>
		<button
			class="tab-button"
			class:active={activeTab === 'activities'}
			onclick={() => handleTabChange('activities')}
		>
			All Activities
		</button>
	</nav>

	<!-- Tab Content -->
	<main class="tab-content">
		{#if activeTab === 'accounts'}
			<section class="accounts-content">
				{#if isLoading}
					<div class="loading-state">
						<div class="loading-spinner"></div>
						<p>Loading accounts...</p>
					</div>
				{:else if accounts && accounts.length > 0}
					<div class="accounts-list">
						{#each accounts as account (account.accountId)}
							<AccountCard
								{account}
								recentTransactions={getRecentTransactions(account.accountId)}
								onAccountClick={handleAccountClick}
								{formatBalance}
								{formatDate}
							/>
						{/each}
					</div>
				{:else}
					<div class="empty-state">
						<div class="empty-icon">
							<svg class="empty-icon-svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
								/>
							</svg>
						</div>
						<h3 class="empty-title">No accounts found</h3>
						<p class="empty-description">Create your first account to get started.</p>
						<Button onclick={() => goto('/account/create')}>Create Account</Button>
					</div>
				{/if}
			</section>
		{:else}
			<section class="activities-content">
				<ActivityControls
					{accounts}
					{filterAccount}
					{sortBy}
					{sortOrder}
					onFilterChange={handleFilterChange}
					onSortByChange={handleSortByChange}
					onSortOrderChange={handleSortOrderChange}
				/>

				{#if isLoading}
					<div class="loading-state">
						<div class="loading-spinner"></div>
						<p>Loading transactions...</p>
					</div>
				{:else if filteredTransactions.length > 0}
					<TransactionList
						transactions={filteredTransactions}
						{getAccountNumber}
						{formatBalance}
						{formatDate}
					/>
				{:else}
					<div class="empty-state">
						<div class="empty-icon">
							<svg class="empty-icon-svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
								/>
							</svg>
						</div>
						<h3 class="empty-title">No transactions found</h3>
						<p class="empty-description">No transactions match your current filters.</p>
					</div>
				{/if}
			</section>
		{/if}
	</main>
</div>

<style>
	@reference "../../app.css";

	.dashboard-container {
		@apply max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8;
	}

	.dashboard-header {
		@apply mb-8;
	}

	.header-content {
		@apply mb-6;
	}

	.dashboard-title {
		@apply text-3xl font-bold text-gray-900;
	}

	.dashboard-subtitle {
		@apply mt-2 text-gray-600;
	}

	.total-balance-card {
		@apply bg-primary-50 rounded-lg p-6 flex justify-between items-center;
	}

	.total-balance {
		@apply flex flex-col;
	}

	.balance-amount {
		@apply text-3xl font-bold text-primary-600;
	}

	.balance-label {
		@apply text-sm text-primary-500 mt-1;
	}

	/* Tab Navigation */
	.tab-navigation {
		@apply flex border-b border-gray-200 mb-8;
	}

	.tab-button {
		@apply px-6 py-3 text-sm font-medium text-gray-500 hover:text-gray-700 border-b-2 border-transparent hover:border-gray-300 transition-colors cursor-pointer;
	}

	.tab-button.active {
		@apply text-primary-600 border-primary-600;
	}

	/* Tab Content */
	.tab-content {
		@apply min-h-96;
	}

	/* Accounts Tab */
	.accounts-content {
		@apply space-y-4;
	}

	.accounts-list {
		@apply space-y-4;
	}

	/* Activities Tab */
	.activities-content {
		@apply space-y-6;
	}

	/* Common styles */
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
</style>

