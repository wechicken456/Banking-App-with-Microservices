import { authStore } from '$lib/stores/auth.svelte';
import { browser } from '$app/environment';
import { goto } from '$app/navigation';

export function requireAuth() : boolean {
    if (browser && !authStore.isAuthenticated) {
        goto('/login');
        return false;
    }
    return true;
}

export function redirectIfAuthenticated() : boolean {
    if (browser && authStore.isAuthenticated) {
        goto('/dashboard');
        return true;
    }
    return false;
}