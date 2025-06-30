<script lang="ts">
    import { authStore } from '$lib/stores/auth.svelte';
    import Button from '$lib/components/ui/Button.svelte';

    const user = $derived(authStore.user);
</script>

<svelte:head>
    <title>Dashboard - Banking App</title>
</svelte:head>

<div class="dashboard-container">
    <div class="dashboard-header">
        <h1 class="dashboard-title">Dashboard</h1>
        <p class="dashboard-subtitle">Welcome back, {user?.email}!</p>
    </div>

    <div class="dashboard-grid">
        <!-- Account Overview Card -->
        <div class="card">
            <h3 class="card-title">Account Overview</h3>
            <div class="account-overview">
                <div class="balance">$0.00</div>
                <p class="balance-label">Available Balance</p>
            </div>
            <Button variant="outline" onclick={() => alert('Account management coming soon!')}>
                Manage Account
            </Button>
        </div>

        <!-- Quick Actions Card -->
        <div class="card">
            <h3 class="card-title">Quick Actions</h3>
            <div class="actions-list">
                <Button 
                    variant="primary" 
                    onclick={() => alert('Transfer feature coming soon!')}
                >
                    Send Money
                </Button>
                <Button 
                    variant="outline" 
                    onclick={() => alert('Deposit feature coming soon!')}
                >
                    Deposit Funds
                </Button>
                <Button 
                    variant="outline" 
                    onclick={() => alert('Transaction history coming soon!')}
                >
                    View Transactions
                </Button>
            </div>
        </div>

        <!-- Recent Activity Card -->
        <div class="card">
            <h3 class="card-title">Recent Activity</h3>
            <div class="empty-state">
                <svg class="empty-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p class="empty-text">No recent transactions</p>
            </div>
        </div>
    </div>

</div>

<style>
    @reference "../../app.css";
    .dashboard-container {
        @apply max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8;
    }

    .dashboard-header {
        @apply mb-8;
    }

    .dashboard-title {
        @apply text-3xl font-bold text-gray-900;
    }

    .dashboard-subtitle {
        @apply mt-2 text-gray-600;
    }

    .dashboard-grid {
        @apply grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6;
    }

    .card {
        @apply bg-white p-6 rounded-lg shadow-md;
    }

    .card-title {
        @apply text-lg font-semibold text-gray-900 mb-4;
    }

    .account-overview {
        @apply text-center py-8;
    }

    .balance {
        @apply text-3xl font-bold text-primary-600 mb-2;
    }

    .balance-label {
        @apply text-gray-600;
    }

    .actions-list {
        @apply space-y-3;
    }

    .empty-state {
        @apply text-center py-8 text-gray-500;
    }

    .empty-icon {
        @apply h-12 w-12 mx-auto mb-4;
    }

    .empty-text {
        @apply text-gray-500;
    }

    .features-grid {
        @apply grid grid-cols-1 md:grid-cols-2 gap-4 text-blue-800;
    }

    .features-list {
        @apply space-y-2;
    }
</style>