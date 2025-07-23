<script lang="ts">
    interface Props {
        variant?: 'primary' | 'secondary' | 'outline' | 'destructive';
        size?: 'sm' | 'md' | 'lg';
        disabled?: boolean;
        loading?: boolean;
        type?: 'button' | 'submit' | 'reset';
        onclick?: () => void;
        children: any;
    }

    let {
        variant = 'primary',
        size = 'md',
        disabled = false,
        loading = false,
        type = 'button',
        onclick,
        children
    }: Props = $props();

    const baseClasses = 'btn';
    const variantClass = `btn-${variant}`;
    const sizeClass = `btn-${size}`;
    const classes = `${baseClasses} ${variantClass} ${sizeClass}`;
</script>

<button
    {type}
    class={classes}
    disabled={disabled || loading}
    onclick={onclick}
>
    {#if loading}
        <svg class="spinner" fill="none" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="opacity-25"></circle>
            <path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" class="opacity-75"></path>
        </svg>
    {/if}
    {@render children()}
</button>

<style>
    @reference "../../../app.css";

    .btn {
        @apply inline-flex items-center justify-center rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none;
    }

    .btn-primary {
        @apply bg-primary-600 text-white hover:bg-primary-700 focus:ring-primary-500;
    }

    .btn-secondary {
        @apply bg-gray-600 text-white hover:bg-gray-700 focus:ring-gray-500;
    }

    .btn-outline {
        @apply border border-gray-300 bg-white text-gray-700 hover:bg-gray-50 focus:ring-primary-500;
    }

    .btn-sm {
        @apply px-3 py-2 text-sm;
    }

    .btn-md {
        @apply px-4 py-2 text-base;
    }

    .btn-lg {
        @apply px-6 py-3 text-lg;
    }

    .spinner {
        @apply animate-spin -ml-1 mr-2 h-4 w-4;
    }
</style>