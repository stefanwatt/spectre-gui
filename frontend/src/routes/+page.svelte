<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { GetRoute } from '$lib/wailsjs/go/main/App';

	onMount(async () => {
		const route = await GetRoute();
		console.log(route);
		goto(route);
	});
</script>
