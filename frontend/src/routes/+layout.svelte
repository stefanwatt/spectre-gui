<script>
	import Toast from '$lib/notification/Toast.svelte';
	import { toast } from '$lib/store';
	import { fade } from 'svelte/transition';
</script>

<!-- <div -->
<!-- 	class="min-w-screen flex h-full min-h-screen w-full flex-col overflow-hidden bg-base px-2 py-4" -->
<!-- > -->
<slot />
<!-- </div> -->
{#if $toast}
	<div transition:fade={{ delay: 250, duration: 300 }}>
		<Toast text={$toast.text} level={$toast.level}></Toast>
	</div>
{/if}
