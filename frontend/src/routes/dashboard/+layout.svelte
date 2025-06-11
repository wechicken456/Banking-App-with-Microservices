<script lang="ts">
    import ProtectedRoutes from '$lib/components/ProtectedRoutes.svelte';
    
    interface Props {
        children: any;
    }

    let { children }: Props = $props();
</script>

<ProtectedRoutes>
    {@render children()}
</ProtectedRoutes>