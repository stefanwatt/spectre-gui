<script>
	import Toast from '$lib/notification/Toast.svelte';
	import { fade } from 'svelte/transition';
	import { state } from '$lib/picker/live-grep-results/results.service.svelte';
</script>

<!-- <div -->
<!-- 	class="min-w-screen flex h-full min-h-screen w-full flex-col overflow-hidden bg-base px-2 py-4" -->
<!-- > -->
<slot />
<!-- </div> -->
{#if state.toast}
	<div transition:fade={{ delay: 250, duration: 300 }}>
		<Toast text={state.toast.text} level={state.toast.level}></Toast>
	</div>
{/if}
