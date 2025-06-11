<script lang="ts">
    import '../app.css';
    import Navigation from '$lib/components/Navigation.svelte';
    import Toast from '$lib/components/ui/Toast.svelte';
    
    interface Props {
        children: any;
    }

    let { children }: Props = $props();
</script>

<div class="app-container">
    <Navigation />
    <main class="main-content">
        {@render children()}
    </main>
    <Toast />
</div>

<style>
    @reference "../app.css";
    .app-container {
        @apply min-h-screen bg-gray-50;
    }

    .main-content {
        @apply flex-1;
    }
</style>