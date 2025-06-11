<script lang="ts">
    import { authStore } from '$lib/stores/auth.svelte';
    import Button from '$lib/components/ui/Button.svelte';

    const isAuthenticated = $derived(authStore.isAuthenticated);
    const user = $derived(authStore.user);

    async function handleLogout() {
        await authStore.logout();
    }
</script>

<nav class="navbar">
    <div class="navbar-container">
        <div class="navbar-content">
            <div class="navbar-brand">
                <a href="/" class="brand-link">
                    Banking App
                </a>
            </div>

            <div class="navbar-actions">
                {#if isAuthenticated && user}
                    <span class="user-welcome">Welcome, {user.email}</span>
                    <Button variant="outline" onclick={handleLogout}>
                        Logout
                    </Button>
                {:else}
                    <a href="/login" class="nav-link">Login</a>
                    <a href="/register" class="nav-link">Register</a>
                {/if}
            </div>
        </div>
    </div>
</nav>

<style>
    @reference "../../app.css";
    .navbar {
        @apply bg-white shadow-sm border-b;
    }

    .navbar-container {
        @apply max-w-7xl mx-auto px-4 sm:px-6 lg:px-8;
    }

    .navbar-content {
        @apply flex justify-between h-16;
    }

    .navbar-brand {
        @apply flex items-center;
    }

    .brand-link {
        @apply text-xl font-bold text-primary-600;
    }

    .navbar-actions {
        @apply flex items-center space-x-4;
    }

    .user-welcome {
        @apply text-gray-700;
    }

    .nav-link {
        @apply text-gray-600 hover:text-gray-900;
    }
</style>