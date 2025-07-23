<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { authStore } from '$lib/stores/auth.svelte';
    import { accountStore } from '$lib/stores/account.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';
    import Button from '$lib/components/ui/Button.svelte';
    import Input from '$lib/components/ui/Input.svelte';

    const user = $derived(authStore.user);
    const isLoading = $derived(accountStore.isLoading);

    let balance = $state('');
    let balanceError = $state<string | null>(null);

    onMount(() => {
        if (!user) {
            goto('/login');
        }
    });

    function validateBalance(): boolean {
        const balanceNum = parseFloat(balance);
        
        if (isNaN(balanceNum)) {
            balanceError = 'Please enter a valid number';
            return false;
        }
        
        if (balanceNum <= 0) {
            balanceError = 'Initial balance cannot be negative or 0';
            return false;
        }
        
        if (balanceNum > 1000000) {
            balanceError = 'Initial balance cannot exceed $1,000,000';
            return false;
        }
        
        balanceError = null;
        return true;
    }

    async function handleSubmit(e: SubmitEvent) {
        e.preventDefault();
        
        if (!validateBalance()) return;

        try {
            const balanceInCents = Math.round(parseFloat(balance) * 100);
            const newAccount = await accountStore.createAccount(balanceInCents);
            
            toastStore.success(`Account #${newAccount.accountNumber} created successfully!`);
            goto('/dashboard');
        } catch (error) {
            console.error('Failed to create account:', error);
            toastStore.error('Failed to create account. Please try again.');
        }
    }

    function handleCancel() {
        goto('/dashboard');
    }

    function formatPreviewBalance(): string {
        if (!balance || isNaN(parseFloat(balance))) return '$0.00';
        return `$${parseFloat(balance).toLocaleString('en-US', { 
            minimumFractionDigits: 2, 
            maximumFractionDigits: 2 
        })}`;
    }
</script>

<svelte:head>
    <title>Create Account - Banking App</title>
</svelte:head>

<div class="create-account-container">
    <div class="create-account-header">
        <div class="header-navigation">
            <Button variant="outline" onclick={handleCancel} disabled={isLoading}>
                Back to Dashboard
            </Button>
        </div>
        
        <div class="header-content">
            <h1 class="page-title">Create New Account</h1>
            <p class="page-subtitle">Set up a new banking account with an initial deposit</p>
        </div>
    </div>

    <div class="create-account-content">
        <div class="account-form-card">
            <form onsubmit={handleSubmit} class="account-form">
                <div class="form-section">
                    <h2 class="section-title">Account Details</h2>
                    <div class="form-field">
                        <Input
                            label="Initial Balance"
                            type="number"
                            placeholder="0.00"
                            bind:value={balance}
                            error={balanceError}
                            required={true}
                            disabled={isLoading}
                        />
                        <p class="field-help">Initial deposit amount</p>
                    </div>
                </div>

                <div class="preview-section">
                    <h3 class="preview-title">Account Preview</h3>
                    <div class="preview-card">
                        <div class="preview-header">
                            <span class="preview-label">New Account</span>
                            <span class="preview-balance">{formatPreviewBalance()}</span>
                        </div>
                        <div class="preview-details">
                            <p class="preview-detail">
                                <span class="detail-label">Account Holder:</span>
                                <span class="detail-value">{user?.email || 'Loading...'}</span>
                            </p>
                            <p class="preview-detail">
                                <span class="detail-label">Initial Deposit:</span>
                                <span class="detail-value">{formatPreviewBalance()}</span>
                            </p>
                        </div>
                    </div>
                </div>

                <div class="form-actions">
                    <Button
                        type="button"
                        variant="outline"
                        onclick={handleCancel}
                        disabled={isLoading}
                    >
                        Cancel
                    </Button>
                    <Button
                        type="submit"
                        variant="primary"
                        loading={isLoading}
                        disabled={isLoading}
                    >
                        {isLoading ? 'Creating Account...' : 'Create Account'}
                    </Button>
                </div>
            </form>
        </div>
    </div>
</div>

<style>
    @reference "../../../app.css";

    .create-account-container {
        @apply max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8;
    }

    .create-account-header {
        @apply mb-8 space-y-6;
    }

    .header-navigation {
        @apply flex;
    }

    .header-content {
        @apply text-center;
    }

    .page-title {
        @apply text-3xl font-bold text-gray-900;
    }

    .page-subtitle {
        @apply mt-2 text-gray-600;
    }

    .create-account-content {
        @apply grid grid-cols-1 lg:grid-cols-3 gap-8;
    }

    .account-form-card {
        @apply lg:col-span-2 bg-white rounded-lg shadow-sm p-6;
    }

    .account-form {
        @apply space-y-8;
    }

    .form-section {
        @apply space-y-4;
    }

    .section-title {
        @apply text-xl font-semibold text-gray-900 mb-4;
    }

    .form-field {
        @apply space-y-2;
    }

    .field-help {
        @apply text-sm text-gray-500;
    }

    .preview-section {
        @apply space-y-4;
    }

    .preview-title {
        @apply text-lg font-semibold text-gray-900;
    }

    .preview-card {
        @apply border border-gray-200 rounded-lg p-4 bg-gray-50;
    }

    .preview-header {
        @apply flex justify-between items-center mb-4 pb-4 border-b border-gray-200;
    }

    .preview-label {
        @apply text-sm font-medium text-gray-600;
    }

    .preview-balance {
        @apply text-2xl font-bold text-primary-600;
    }

    .preview-details {
        @apply space-y-2;
    }

    .preview-detail {
        @apply flex justify-between text-sm;
    }

    .detail-label {
        @apply text-gray-600;
    }

    .detail-value {
        @apply font-medium text-gray-900;
    }

    .form-actions {
        @apply flex gap-4 justify-end pt-6 border-t border-gray-200;
    }
</style>