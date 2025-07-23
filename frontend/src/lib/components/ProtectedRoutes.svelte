<script lang="ts">
    import { authStore } from '$lib/stores/auth.svelte';
    import { browser } from '$app/environment';
    import { goto } from '$app/navigation';

    interface Props {
        children: any;
    }

    let { children }: Props = $props();

    const isAuthenticated = $derived(authStore.isAuthenticated);
    const isLoading = $derived(authStore.isLoading);

    $effect(() => {
        if (browser && !isLoading && !isAuthenticated) {
            goto('/login');
        }
    });
</script>

{#if isLoading}
    <div class="loading-container">
        <div class="loading-spinner"></div>
    </div>
{:else if isAuthenticated}
    {@render children()}
{:else}
    <div class="redirect-container">
        <p class="redirect-text">Redirecting to login...</p>
    </div>
{/if}

<style>
    @reference "../../app.css";
    
    .loading-container {
        @apply flex items-center justify-center min-h-screen;
    }

    .loading-spinner {
        @apply animate-spin rounded-full h-32 w-32 border-b-2 border-primary-600;
    }

    .redirect-container {
        @apply flex items-center justify-center min-h-screen;
    }

    .redirect-text {
        @apply text-gray-600;
    }
</style>