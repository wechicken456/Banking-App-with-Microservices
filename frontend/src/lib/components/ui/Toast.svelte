<script lang="ts">
    import { toastStore } from '$lib/stores/toast.svelte';
    
    const toasts = $derived(toastStore.toasts);
</script>

{#if toasts.length > 0}
    <div class="toast-container">
        {#each toasts as toast (toast.id)}
            <div class="toast" role="alert">
                <div class="toast-content">
                    <div class="toast-body">
                        <div class="toast-icon">
                            {#if toast.type === 'success'}
                                <svg class="icon icon-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                            {:else if toast.type === 'error'}
                                <svg class="icon icon-error" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                            {:else}
                                <svg class="icon icon-info" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                            {/if}
                        </div>
                        <div class="toast-message">
                            <p class="message-text">{toast.message}</p>
                        </div>
                        <div class="toast-close">
                            <button
                                class="close-button"
                                onclick={() => toastStore.remove(toast.id)}
                            >
                                <span class="sr-only">Close</span>
                                <svg class="close-icon" viewBox="0 0 20 20" fill="currentColor">
                                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        {/each}
    </div>
{/if}

<style>
    @reference "../../../app.css";
    .toast-container {
        @apply fixed top-4 right-4 z-50 space-y-2;
    }

    .toast {
        @apply max-w-sm w-full bg-white shadow-lg rounded-lg pointer-events-auto ring-1 ring-black/5 overflow-hidden transform transition-all duration-300 ease-in-out;
    }

    .toast-content {
        @apply p-4;
    }

    .toast-body {
        @apply flex items-start;
    }

    .toast-icon {
        @apply flex-shrink-0;
    }

    .icon {
        @apply h-6 w-6;
    }

    .icon-success {
        @apply text-green-400;
    }

    .icon-error {
        @apply text-red-400;
    }

    .icon-info {
        @apply text-blue-400;
    }

    .toast-message {
        @apply ml-3 w-0 flex-1 pt-0.5;
    }

    .message-text {
        @apply text-sm font-medium text-gray-900;
    }

    .toast-close {
        @apply ml-4 flex-shrink-0 flex;
    }

    .close-button {
        @apply bg-white rounded-md inline-flex text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500;
    }

    .close-icon {
        @apply h-5 w-5;
    }

    .sr-only {
        @apply sr-only;
    }
</style>